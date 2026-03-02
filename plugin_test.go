package main

import (
	"os"
	"testing"
	"time"

	"github.com/appleboy/drone-template-lib/template"
	"github.com/stretchr/testify/assert"
)

func TestMissingDefaultConfig(t *testing.T) {
	var plugin Plugin

	err := plugin.Exec()

	assert.NotNil(t, err)
}

func TestMissingUserConfig(t *testing.T) {
	plugin := Plugin{
		Config: Config{
			Token: "123456789",
		},
	}

	err := plugin.Exec()

	assert.NotNil(t, err)
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

	assert.Equal(t, []string{"✅ Build #101 of `appleboy/go-hello` success.\n\n📝 Commit by Bo-Yi Wu on `master`:\n``` update travis ```\n\n🌐 https://github.com/appleboy/go-hello"}, message)
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

	assert.Equal(t, []string{"appleboy/go-hello/test-workflow triggered by appleboy (push)"}, message)
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
			Token:    os.Getenv("TELEGRAM_TOKEN"),
			To:       []string{os.Getenv("TELEGRAM_TO"), os.Getenv("TELEGRAM_TO") + ":appleboy@gmail.com", "中文ID", "1234567890"},
			Message:  "Test Telegram Chat Bot From Travis or Local, commit message: 『{{ build.message }}』",
			Photo:    []string{"tests/github.png", "1234", " "},
			Document: []string{"tests/gophercolor.png", "1234", " "},
			Sticker:  []string{"tests/github-logo.png", "tests/github.png", "1234", " "},
			Audio:    []string{"tests/audio.mp3", "1234", " "},
			Voice:    []string{"tests/voice.ogg", "1234", " "},
			Location: []string{"24.9163213 121.1424972", "1", " "},
			Venue:    []string{"35.661777 139.704051 竹北體育館 新竹縣竹北市", "24.9163213 121.1424972", "1", " "},
			Video:    []string{"tests/video.mp4", "1234", " "},
			Debug:    false,
		},
	}

	err := plugin.Exec()
	assert.NotNil(t, err)

	plugin.Config.Format = formatMarkdown
	plugin.Config.Message = "Test escape under_score"
	err = plugin.Exec()
	assert.NotNil(t, err)

	// disable message
	plugin.Config.Message = ""
	err = plugin.Exec()
	assert.NotNil(t, err)
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
	assert.Nil(t, err)

	// disable message
	plugin.Config.DisableWebPagePreview = false
	plugin.Config.Message = "EnableWebPagePreview https://www.google.com.tw"
	err = plugin.Exec()
	assert.Nil(t, err)
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
	assert.Nil(t, err)

	// disable message
	plugin.Config.DisableNotification = false
	plugin.Config.Message = "EnableNotification https://www.google.com.tw"
	err = plugin.Exec()
	assert.Nil(t, err)
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
	assert.NotNil(t, err)
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
	input := []string{"0", "1:1@gmail.com", "2:2@gmail.com", "3:3@gmail.com", "4", "5#7"}

	ids := parseTo(input, "1@gmail.com", false)
	assert.Equal(t, []Chat{Chat{0, 0}, Chat{4, 0}, Chat{5, 7}, Chat{1, 0}}, ids)

	ids = parseTo(input, "1@gmail.com", true)
	assert.Equal(t, []Chat{Chat{1, 0}}, ids)

	ids = parseTo(input, "a@gmail.com", false)
	assert.Equal(t, []Chat{Chat{0, 0}, Chat{4, 0}, Chat{5, 7}}, ids)

	ids = parseTo(input, "a@gmail.com", true)
	assert.Equal(t, []Chat{Chat{0, 0}, Chat{4, 0}, Chat{5, 7}}, ids)

	// test empty ids
	ids = parseTo([]string{"", " ", "   "}, "a@gmail.com", true)
	assert.Equal(t, 0, len(ids))
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

	assert.Equal(t, true, empty)
	assert.Equal(t, Location{}, result)

	// strconv.ParseInt: parsing "測試": invalid syntax
	input = "測試 139.704051"
	result, empty = convertLocation(input)

	assert.Equal(t, true, empty)
	assert.Equal(t, Location{}, result)

	// strconv.ParseInt: parsing "測試": invalid syntax
	input = "35.661777 測試"
	result, empty = convertLocation(input)

	assert.Equal(t, true, empty)
	assert.Equal(t, Location{}, result)

	input = "35.661777 139.704051"
	result, empty = convertLocation(input)

	assert.Equal(t, false, empty)
	assert.Equal(t, Location{
		Latitude:  float64(35.661777),
		Longitude: float64(139.704051),
	}, result)

	input = "35.661777 139.704051 title"
	result, empty = convertLocation(input)

	assert.Equal(t, false, empty)
	assert.Equal(t, Location{
		Title:     "title",
		Address:   "",
		Latitude:  float64(35.661777),
		Longitude: float64(139.704051),
	}, result)

	input = "35.661777 139.704051 title address"
	result, empty = convertLocation(input)

	assert.Equal(t, false, empty)
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

	assert.Nil(t, plugin.Exec())

	plugin.Config.MessageFile = "tests/message_html.txt"
	assert.Nil(t, plugin.Exec())
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
	assert.Nil(t, err)
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
	assert.Nil(t, err)
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
	assert.Nil(t, err)
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
	assert.Nil(t, err)
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
	assert.Nil(t, err)
}
