package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"testing"
)

func widgetFile(t *testing.T, payload map[string]any) string {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return richFile(t, string(data))
}

func TestLoadWidget(t *testing.T) {
	for _, input := range []string{
		`null`, `[]`, `{}`, `{"html":""}`, `{"html":42}`, `{"html":null}`,
		`{"html":"a","html":"b"}`, `{"html":"a"}{}`, `{"html":"a","token":"secret"}`,
		`{"html":"a","chat_id":42}`, `{"html":"a","message_thread_id":1}`,
		`{"html":"a","disable_notification":true}`, `{"html":"a","width":true}`,
		`{"html":"a","height":1.5}`, `{"html":"a","allow_queries":"true"}`,
		`{"html":"a","caption_entities":{}}`, `{"html":"a","caption_entities":[null]}`,
		`{"html":"a","reply_markup":[]}`, `{"html":"a","parse_mode":"xml"}`,
		`{"html":"a","parse_mode":"HTML","caption_entities":[]}`,
		"{\"html\":\"\xff\"}", strings.Repeat(" ", maxWidgetFile+1),
	} {
		if _, err := loadWidget(richFile(t, input)); err == nil {
			t.Errorf("invalid JSON accepted (length %d)", len(input))
		}
	}
	if _, err := loadWidget("/missing/widget.json"); err == nil {
		t.Fatal("missing file accepted")
	}
	for _, payload := range []map[string]any{
		{"html": "<b>{{caller-owned}}</b>", "caption": "Passed", "width": 240, "height": 100},
		{"html": strings.Repeat("é", 16384), "chat_data": strings.Repeat("é", 512), "caption": strings.Repeat("🙂", 1024)},
		{"html": "x", "caption": "<b>" + strings.Repeat("x", 1024) + "</b>", "parse_mode": "HTML"},
		{"html": "x", "caption_entities": []any{}, "reply_markup": map[string]any{}, "allow_queries": false},
	} {
		if _, err := loadWidget(widgetFile(t, payload)); err != nil {
			t.Fatal(err)
		}
	}
	for _, payload := range []map[string]any{
		{"html": strings.Repeat("é", 16385)},
		{"html": "x", "public_data": strings.Repeat("é", 513)},
		{"html": "x", "chat_data": strings.Repeat("é", 513)},
		{"html": "x", "caption": strings.Repeat("🙂", 1025)},
	} {
		if _, err := loadWidget(widgetFile(t, payload)); err == nil {
			t.Fatal("oversized field accepted")
		}
	}
	rng := rand.New(rand.NewSource(1))
	noise := make([]byte, 10000)
	for i := range noise {
		noise[i] = byte(32 + rng.Intn(90))
	}
	if _, err := loadWidget(widgetFile(t, map[string]any{"html": string(noise)})); err == nil {
		t.Fatal("poorly compressible HTML accepted")
	}
}

func TestWidgetConflictsBeforeNetwork(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	http.DefaultTransport = richTransport(func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid inputs reached network")
		return nil, errors.New("unexpected network request")
	})
	changes := []func(*Config){
		func(c *Config) { c.Message = "text" },
		func(c *Config) { c.MessageFile = "text" },
		func(c *Config) { c.RichMessageFile = "rich" },
		func(c *Config) { c.TemplateVars = "{}" },
		func(c *Config) { c.TemplateVarsFile = "vars" },
		func(c *Config) { c.Format = "HTML" },
		func(c *Config) { c.DisableWebPagePreview = true },
		func(c *Config) { c.Photo = []string{"x"} },
		func(c *Config) { c.Document = []string{"x"} },
		func(c *Config) { c.Sticker = []string{"x"} },
		func(c *Config) { c.Audio = []string{"x"} },
		func(c *Config) { c.Voice = []string{"x"} },
		func(c *Config) { c.Video = []string{"x"} },
		func(c *Config) { c.Location = []string{"x"} },
		func(c *Config) { c.Venue = []string{"x"} },
		func(c *Config) { c.To = []string{"42", "43"} },
		func(c *Config) { c.MessageThreadID = -1 },
		func(c *Config) { c.To = []string{":"} },
	}
	for _, change := range changes {
		p := Plugin{Config: Config{Token: "fake", To: []string{"42"}, WidgetFile: "unused"}}
		change(&p.Config)
		if err := p.Exec(); err == nil {
			t.Fatal("conflict accepted")
		}
	}
}

func TestExecWidget(t *testing.T) {
	for _, scenario := range []string{"success", "timeout", "api-error", "invalid-response", "redirect", "getMe-error"} {
		t.Run(scenario, func(t *testing.T) {
			original := http.DefaultTransport
			t.Cleanup(func() { http.DefaultTransport = original })
			var logs bytes.Buffer
			previous := log.Writer()
			log.SetOutput(&logs)
			t.Cleanup(func() { log.SetOutput(previous) })
			calls := 0
			http.DefaultTransport = richTransport(func(r *http.Request) (*http.Response, error) {
				body := `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test"}}`
				status := 200
				header := make(http.Header)
				if strings.HasSuffix(r.URL.Path, "/getMe") {
					if scenario == "getMe-error" {
						return nil, errors.New("secret-token")
					}
				} else {
					calls++
					if !strings.HasSuffix(r.URL.Path, "/sendWidget") ||
						r.Method != http.MethodPost {
						t.Fatal("unexpected send or redirect")
					}
					if err := r.ParseForm(); err != nil {
						t.Fatal(err)
					}
					for key, want := range map[string]string{
						"html": "<b>{{unchanged}} &amp;</b>", "caption": "Build passed",
						"chat_id": "42", "message_thread_id": "7", "disable_notification": "true",
					} {
						if r.Form.Get(key) != want {
							t.Errorf("unexpected %s", key)
						}
					}
					body = `{"ok":true,"result":{"message_id":99}}`
					switch scenario {
					case "timeout":
						return nil, errors.New("timeout secret-token")
					case "api-error":
						status = 429
						body = `{"ok":false,"description":"secret-token","parameters":{"retry_after":1}}`
					case "invalid-response":
						body = "secret-token"
					case "redirect":
						status = 307
						header.Set("Location", "https://other.invalid/leak")
						body = "redirect"
					}
				}
				return &http.Response{
					StatusCode: status,
					Header:     header,
					Body:       io.NopCloser(strings.NewReader(body)),
					Request:    r,
				}, nil
			})
			p := Plugin{Config: Config{
				Token:               "secret-token",
				To:                  []string{"42"},
				MessageThreadID:     7,
				DisableNotification: true,
				Debug:               true,
				WidgetFile: widgetFile(
					t,
					map[string]any{"html": "<b>{{unchanged}} &amp;</b>", "caption": "Build passed"},
				),
			}}
			err := p.Exec()
			if (err == nil) != (scenario == "success") {
				t.Fatalf("unexpected result: %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "secret-token") {
				t.Fatal("credential leaked")
			}
			expected := 1
			if scenario == "getMe-error" {
				expected = 0
			}
			if calls != expected {
				t.Fatalf("expected %d sends, got %d", expected, calls)
			}
			if logs.Len() != 0 {
				t.Fatalf("widget mode emitted debug logs: %s", logs.String())
			}
		})
	}
}
