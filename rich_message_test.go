package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func richFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rich.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadRichMessage(t *testing.T) {
	for _, tc := range []struct {
		input string
		valid bool
	}{
		{`{"html":"<h2>Hello</h2>"}`, true},
		{`{"markdown":"# Hello"}`, true},
		{`{"blocks":[{"type":"paragraph","text":"Hello"}]}`, true},
		{`null`, false}, {`[]`, false}, {`{}`, false}, {`{"html":42}`, false},
		{`{"html":"hello","markdown":"hello"}`, false},
		{`{"blocks":[]}`, false}, {`{"html":"hello"}{}`, false},
	} {
		t.Run(tc.input, func(t *testing.T) {
			_, err := loadRichMessage(richFile(t, tc.input))
			if (err == nil) != tc.valid {
				t.Fatalf("unexpected validation result: %v", err)
			}
		})
	}
	if _, err := loadRichMessage(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing file accepted")
	}
}

type richTransport func(*http.Request) (*http.Response, error)

func (f richTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestExecRichMessage(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	calls := 0
	const markup = "<h2>&lt;{{commit.message}}&gt;</h2>"
	http.DefaultTransport = richTransport(func(r *http.Request) (*http.Response, error) {
		body := `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test"}}`
		if !strings.HasSuffix(r.URL.Path, "/getMe") {
			calls++
			if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/sendRichMessage") {
				t.Error("unexpected request")
			}
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			var payload struct {
				HTML string `json:"html"`
				Skip bool   `json:"skip_entity_detection"`
			}
			if err := json.Unmarshal([]byte(r.Form.Get("rich_message")), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.HTML != markup || !payload.Skip {
				t.Error("rich content was altered")
			}
			if r.Form.Get("chat_id") != "42" || r.Form.Get("message_thread_id") != "7" || r.Form.Get("disable_notification") != "true" {
				t.Error("delivery settings lost")
			}
			body = `{"ok":true,"result":{"message_id":99,"date":0,"chat":{"id":42,"type":"private"}}}`
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	p := Plugin{Config: Config{Token: "fake-token", To: []string{"42"}, MessageThreadID: 7, DisableNotification: true,
		RichMessageFile: richFile(t, `{"html":"`+markup+`","skip_entity_detection":true}`)}}
	if err := p.Exec(); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected one rich message, got %d", calls)
	}
	for _, file := range []bool{false, true} {
		p.Config.Message, p.Config.MessageFile = "", ""
		if file {
			p.Config.MessageFile = "ordinary.txt"
		} else {
			p.Config.Message = "ordinary"
		}
		if err := p.Exec(); err == nil {
			t.Fatal("mixed message inputs accepted")
		}
	}
	if calls != 1 {
		t.Fatal("mixed inputs reached network")
	}
}
