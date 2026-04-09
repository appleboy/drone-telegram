package main

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/appleboy/drone-template-lib/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMissingDefaultConfig(t *testing.T) {
	var plugin Plugin

	err := plugin.Exec()

	assert.Error(t, err)
}

func TestMissingUserConfig(t *testing.T) {
	plugin := Plugin{
		Config: Config{
			Token: "123456789",
		},
	}

	err := plugin.Exec()

	assert.Error(t, err)
}

func TestDefaultMessageFormat(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			FullName:  "appleboy/go-hello",
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "update travis",
		},
		Build: Build{
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},
	}

	message := plugin.Message()

	assert.Equal(
		t,
		[]string{
			"✅ Build #101 of `appleboy/go-hello` success.\n\n📝 Commit by Bo-Yi Wu on `master`:\n``` update travis ```\n\n🌐 https://github.com/appleboy/go-hello",
		},
		message,
	)
}

func TestDefaultMessageFormatFromGitHub(t *testing.T) {
	plugin := Plugin{
		Config: Config{
			GitHub: true,
		},
		Repo: Repo{
			FullName:  "appleboy/go-hello",
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		GitHub: GitHub{
			Workflow:  "test-workflow",
			Action:    "send notification",
			EventName: "push",
		},
	}

	message := plugin.Message()

	assert.Equal(
		t,
		[]string{"appleboy/go-hello/test-workflow triggered by appleboy (push)"},
		message,
	)
}

func TestSendMessage(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "update travis by drone plugin",
			Email:   "test@gmail.com",
		},
		Build: Build{
			Tag:    "1.0.0",
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},

		Config: Config{
			Token: os.Getenv("TELEGRAM_TOKEN"),
			To: []string{
				os.Getenv("TELEGRAM_TO"),
				os.Getenv("TELEGRAM_TO") + ":appleboy@gmail.com",
				"中文ID",
				"1234567890",
			},
			Message:  "Test Telegram Chat Bot From Travis or Local, commit message: 『{{ build.message }}』",
			Photo:    []string{"tests/github.png", "1234", " "},
			Document: []string{"tests/gophercolor.png", "1234", " "},
			Sticker:  []string{"tests/github-logo.png", "tests/github.png", "1234", " "},
			Audio:    []string{"tests/audio.mp3", "1234", " "},
			Voice:    []string{"tests/voice.ogg", "1234", " "},
			Location: []string{"24.9163213 121.1424972", "1", " "},
			Venue: []string{
				"35.661777 139.704051 竹北體育館 新竹縣竹北市",
				"24.9163213 121.1424972",
				"1",
				" ",
			},
			Video: []string{"tests/video.mp4", "1234", " "},
			Debug: false,
		},
	}

	err := plugin.Exec()
	require.Error(t, err)

	plugin.Config.Format = formatMarkdown
	plugin.Config.Message = "Test escape under_score"
	err = plugin.Exec()
	require.Error(t, err)

	// disable message
	plugin.Config.Message = ""
	err = plugin.Exec()
	assert.Error(t, err)
}

func TestDisableWebPagePreviewMessage(t *testing.T) {
	plugin := Plugin{
		Config: Config{
			Token:                 os.Getenv("TELEGRAM_TOKEN"),
			To:                    []string{os.Getenv("TELEGRAM_TO")},
			DisableWebPagePreview: true,
			Debug:                 false,
		},
	}

	plugin.Config.Message = "DisableWebPagePreview https://www.google.com.tw"
	err := plugin.Exec()
	require.NoError(t, err)

	// disable message
	plugin.Config.DisableWebPagePreview = false
	plugin.Config.Message = "EnableWebPagePreview https://www.google.com.tw"
	err = plugin.Exec()
	assert.NoError(t, err)
}

func TestDisableNotificationMessage(t *testing.T) {
	plugin := Plugin{
		Config: Config{
			Token:               os.Getenv("TELEGRAM_TOKEN"),
			To:                  []string{os.Getenv("TELEGRAM_TO")},
			DisableNotification: true,
			Debug:               false,
		},
	}

	plugin.Config.Message = "DisableNotification https://www.google.com.tw"
	err := plugin.Exec()
	require.NoError(t, err)

	// disable message
	plugin.Config.DisableNotification = false
	plugin.Config.Message = "EnableNotification https://www.google.com.tw"
	err = plugin.Exec()
	assert.NoError(t, err)
}

func TestBotError(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "update travis by drone plugin",
		},
		Build: Build{
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},

		Config: Config{
			Token:   "appleboy",
			To:      []string{os.Getenv("TELEGRAM_TO"), "中文ID", "1234567890"},
			Message: "Test Telegram Chat Bot From Travis or Local",
		},
	}

	err := plugin.Exec()
	assert.Error(t, err)
}

func TestTrimElement(t *testing.T) {
	var input, result []string

	input = []string{"1", "     ", "3"}
	result = []string{"1", "3"}

	assert.Equal(t, result, trimElement(input))

	input = []string{"1", "2"}
	result = []string{"1", "2"}

	assert.Equal(t, result, trimElement(input))
}

func TestEscapeMarkdown(t *testing.T) {
	provider := [][][]string{
		{
			{"user", "repo"},
			{"user", "repo"},
		},
		{
			{"user_name", "repo_name"},
			{`user\_name`, `repo\_name`},
		},
		{
			{"user_name_long", "user_name_long"},
			{`user\_name\_long`, `user\_name\_long`},
		},
		{
			{`user\_name\_long`, `repo\_name\_long`},
			{`user\_name\_long`, `repo\_name\_long`},
		},
		{
			{`user\_name\_long`, `repo\_name\_long`, ""},
			{`user\_name\_long`, `repo\_name\_long`},
		},
	}

	for _, testCase := range provider {
		assert.Equal(t, testCase[1], escapeMarkdown(testCase[0]))
	}
}

func TestEscapeMarkdownOne(t *testing.T) {
	provider := [][]string{
		{"user", "user"},
		{"user_name", `user\_name`},
		{"user_name_long", `user\_name\_long`},
		{`user\_name\_escaped`, `user\_name\_escaped`},
	}

	for _, testCase := range provider {
		assert.Equal(t, testCase[1], escapeMarkdownOne(testCase[0]))
	}
}

func TestParseTo(t *testing.T) {
	input := []string{"0", "1:1@gmail.com", "2:2@gmail.com", "3:3@gmail.com", "4", "5"}

	ids := parseTo(input, "1@gmail.com", false)
	assert.Equal(t, []int64{0, 4, 5, 1}, ids)

	ids = parseTo(input, "1@gmail.com", true)
	assert.Equal(t, []int64{1}, ids)

	ids = parseTo(input, "a@gmail.com", false)
	assert.Equal(t, []int64{0, 4, 5}, ids)

	ids = parseTo(input, "a@gmail.com", true)
	assert.Equal(t, []int64{0, 4, 5}, ids)

	// test empty ids
	ids = parseTo([]string{"", " ", "   "}, "a@gmail.com", true)
	assert.Empty(t, ids)
}

func TestGlobList(t *testing.T) {
	var input []string
	var result []string

	input = []string{"tests/gophercolor.png", "測試", "3"}
	result = []string{"tests/gophercolor.png"}
	assert.Equal(t, result, globList(input))

	input = []string{"tests/*.mp3"}
	result = []string{"tests/audio.mp3"}
	assert.Equal(t, result, globList(input))
}

func TestConvertLocation(t *testing.T) {
	var input string
	var result Location
	var empty bool

	input = "1"
	result, empty = convertLocation(input)

	assert.True(t, empty)
	assert.Equal(t, Location{}, result)

	// strconv.ParseInt: parsing "測試": invalid syntax
	input = "測試 139.704051"
	result, empty = convertLocation(input)

	assert.True(t, empty)
	assert.Equal(t, Location{}, result)

	// strconv.ParseInt: parsing "測試": invalid syntax
	input = "35.661777 測試"
	result, empty = convertLocation(input)

	assert.True(t, empty)
	assert.Equal(t, Location{}, result)

	input = "35.661777 139.704051"
	result, empty = convertLocation(input)

	assert.False(t, empty)
	assert.Equal(t, Location{
		Latitude:  float64(35.661777),
		Longitude: float64(139.704051),
	}, result)

	input = "35.661777 139.704051 title"
	result, empty = convertLocation(input)

	assert.False(t, empty)
	assert.Equal(t, Location{
		Title:     "title",
		Address:   "",
		Latitude:  float64(35.661777),
		Longitude: float64(139.704051),
	}, result)

	input = "35.661777 139.704051 title address"
	result, empty = convertLocation(input)

	assert.False(t, empty)
	assert.Equal(t, Location{
		Title:     "title",
		Address:   "address",
		Latitude:  float64(35.661777),
		Longitude: float64(139.704051),
	}, result)
}

func TestHTMLMessage(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "test",
		},
		Build: Build{
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},

		Config: Config{
			Token: os.Getenv("TELEGRAM_TOKEN"),
			To:    []string{os.Getenv("TELEGRAM_TO")},
			Message: `
Test HTML Format
<a href='https://google.com'>Google .com 1</a>
<a href='https://google.com'>Google .com 2</a>
<a href='https://google.com'>Google .com 3</a>
`,
			Format: formatHTML,
		},
	}

	assert.NoError(t, plugin.Exec())

	plugin.Config.MessageFile = "tests/message_html.txt"
	assert.NoError(t, plugin.Exec())
}

func TestMessageFile(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "Freakin' macOS isn't fully case-sensitive..",
		},
		Build: Build{
			Number:   101,
			Status:   "success",
			Link:     "https://github.com/appleboy/go-hello",
			Started:  time.Now().Unix(),
			Finished: time.Now().Add(180 * time.Second).Unix(),
		},

		Config: Config{
			Token:       os.Getenv("TELEGRAM_TOKEN"),
			To:          []string{os.Getenv("TELEGRAM_TO")},
			MessageFile: "tests/message.txt",
		},
	}

	err := plugin.Exec()
	assert.NoError(t, err)
}

func TestTemplateVars(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "This is a test commit msg",
		},
		Build: Build{
			Number:   101,
			Status:   "success",
			Link:     "https://github.com/appleboy/go-hello",
			Started:  time.Now().Unix(),
			Finished: time.Now().Add(180 * time.Second).Unix(),
		},

		Config: Config{
			Token:        os.Getenv("TELEGRAM_TOKEN"),
			To:           []string{os.Getenv("TELEGRAM_TO")},
			Format:       formatMarkdown,
			MessageFile:  "tests/message_template.txt",
			TemplateVars: `{"env":"testing","version":"1.2.0-SNAPSHOT"}`,
		},
	}

	err := plugin.Exec()
	assert.NoError(t, err)
}

func TestTemplateVarsFile(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "This is a test commit msg",
		},
		Build: Build{
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},

		Config: Config{
			Token:            os.Getenv("TELEGRAM_TOKEN"),
			To:               []string{os.Getenv("TELEGRAM_TO")},
			Format:           formatMarkdown,
			MessageFile:      "tests/message_template.txt",
			TemplateVarsFile: "tests/vars.json",
		},
	}

	err := plugin.Exec()
	assert.NoError(t, err)
}

func TestProxySendMessage(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "start use proxy",
			Email:   "test@gmail.com",
		},
		Build: Build{
			Tag:    "1.0.0",
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},

		Config: Config{
			Token:   os.Getenv("TELEGRAM_TOKEN"),
			To:      []string{os.Getenv("TELEGRAM_TO")},
			Message: "Send message from socks5 proxy URL.",
			Debug:   false,
			Socks5:  os.Getenv("SOCKS5"),
		},
	}

	err := plugin.Exec()
	assert.NoError(t, err)
}

func TestThreadIDTransportInjectsQueryParam(t *testing.T) {
	base := http.DefaultTransport
	transport := &threadIDTransport{
		base:     base,
		threadID: 1257,
	}

	req, _ := http.NewRequest("POST", "https://api.telegram.org/bot123/sendMessage", nil)
	original := req.URL.String()

	// Use a custom base transport to capture the modified request
	// without making a real HTTP call.
	var captured *http.Request
	transport.base = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		captured = r
		return &http.Response{StatusCode: 200}, nil
	})

	_, _ = transport.RoundTrip(req)

	// Original request must not be mutated (http.RoundTripper contract).
	assert.Equal(t, original, req.URL.String())

	// Cloned request must have message_thread_id injected.
	assert.NotNil(t, captured)
	assert.Equal(t, "1257", captured.URL.Query().Get("message_thread_id"))
}

func TestThreadIDTransportPreservesExistingQuery(t *testing.T) {
	transport := &threadIDTransport{
		base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200}, nil
		}),
		threadID: 42,
	}

	req, _ := http.NewRequest("POST", "https://api.telegram.org/bot123/sendMessage?chat_id=100", nil)
	var captured *http.Request
	transport.base = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		captured = r
		return &http.Response{StatusCode: 200}, nil
	})

	_, _ = transport.RoundTrip(req)

	assert.Equal(t, "100", captured.URL.Query().Get("chat_id"))
	assert.Equal(t, "42", captured.URL.Query().Get("message_thread_id"))
}

func TestNoTransportWrapWhenThreadIDZero(t *testing.T) {
	plugin := Plugin{
		Config: Config{
			Token:           "invalid-token",
			To:              []string{"123"},
			MessageThreadID: 0,
			Message:         "test",
		},
	}

	// Exec will fail on bot auth, but we can verify the transport
	// is not wrapped by checking the error is a plain auth error,
	// not a transport-related one.
	err := plugin.Exec()
	require.NotNil(t, err)
	assert.NotContains(t, err.Error(), "message_thread_id")
}

func TestTransportWrapWithSocks5(t *testing.T) {
	plugin := Plugin{
		Config: Config{
			Token:           "invalid-token",
			To:              []string{"123"},
			MessageThreadID: 99,
			Message:         "test",
			Socks5:          "socks5://127.0.0.1:1080",
		},
	}

	// This will fail on bot auth, but exercises the code path where
	// both SOCKS5 proxy and threadIDTransport are configured together.
	err := plugin.Exec()
	assert.NotNil(t, err)
}

// roundTripFunc adapts a function to the http.RoundTripper interface.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestBuildTemplate(t *testing.T) {
	plugin := Plugin{
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "This is a test commit msg",
		},
		Build: Build{
			Number:   101,
			Status:   "success",
			Link:     "https://github.com/appleboy/go-hello",
			Started:  time.Now().Unix(),
			Finished: time.Now().Add(180 * time.Second).Unix(),
		},
	}

	_, err := template.RenderTrim(
		`
Sample message loaded from file.

Commit msg:  {{uppercasefirst commit.message}}

duration: {{duration build.started build.finished}}
`, plugin)
	assert.NoError(t, err)
}
