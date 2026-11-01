package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/appleboy/drone-template-lib/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipIfNoTelegramSecrets(t *testing.T) {
	t.Helper()
	if os.Getenv("TELEGRAM_TOKEN") == "" || os.Getenv("TELEGRAM_TO") == "" {
		t.Skip("TELEGRAM_TOKEN/TELEGRAM_TO not set; skipping integration test")
	}
}

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
	skipIfNoTelegramSecrets(t)
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
	skipIfNoTelegramSecrets(t)
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

func TestEscapeMarkdownContexts(t *testing.T) {
	cases := []struct {
		name, input, want string
	}{
		{"inline code", "`_Some text`", "`_Some text`"},
		{"code block", "```go\nuser_name\n```", "```go\nuser_name\n```"},
		{
			"link",
			"[_Some text](https://example.com/some_path)",
			"[_Some text](https://example.com/some_path)",
		},
		{
			"bare URL",
			"https://example.com/some_path?q=a_b#c_d",
			"https://example.com/some_path?q=a_b#c_d",
		},
		{
			"mixed",
			"user_name `_code` [_label](https://example.com/a_b) tail_end",
			"user\\_name `_code` [_label](https://example.com/a_b) tail\\_end",
		},
		{"empty", "", ""},
		{"Unicode", "中文_🙂 `中文_🙂`", "中文\\_🙂 `中文_🙂`"},
		{"adjacent code", "`a_b` `c_d`", "`a_b` `c_d`"},
		{"backtick inside fence", "```a_`b`_c``` tail_end", "```a_`b`_c``` tail\\_end"},
		{"escaped code delimiter", "\\`a_b\\` tail_end", "\\`a\\_b\\` tail\\_end"},
		{"even backslashes", "\\\\`a_b`", "\\\\`a_b`"},
		{"escaped bracket", `\[a_b](path_name)`, `\[a\_b](path\_name)`},
		{"escaped label end", `[a\]_b](path_name)`, `[a\]_b](path_name)`},
		{
			"balanced target",
			`[a_b](https://example.com/a_(b_c)) tail_end`,
			`[a_b](https://example.com/a_(b_c)) tail\_end`,
		},
		{"escaped target end", `[a_b](path\)_name) tail_end`, `[a_b](path\)_name) tail\_end`},
		{"multiple links", `[a_b](one_two) [c_d](three_four)`, `[a_b](one_two) [c_d](three_four)`},
		{"preserve explicit code escape", "`a\\_b`", "`a\\_b`"},
		{"preserve explicit link escape", `[a\_b](one\_two)`, `[a\_b](one\_two)`},
		{"ordinary escapes", `a\_b c\\_d e\\\_f`, `a\_b c\\_d e\\\_f`},
		{"unclosed code", "`a_b tail_end", "`a\\_b tail\\_end"},
		{"unclosed fence", "```a_b\ntail_end", "```a\\_b\ntail\\_end"},
		{"unclosed link", `[a_b](path_name tail_end`, `[a\_b](path\_name tail\_end`},
		{"unclosed label", `[a_b tail_end`, `[a\_b tail\_end`},
		{
			"bare URL parentheses",
			`(http://example.com/a_(b_c)) tail_end`,
			`(http://example.com/a_(b_c)) tail\_end`,
		},
		{
			"URL Unicode whitespace",
			"https://example.com/a_b\u3000tail_end",
			"https://example.com/a_b\u3000tail\\_end",
		},
		{
			"URL before code",
			"https://example.com/a_b`c_d` tail_end",
			"https://example.com/a_b`c_d` tail\\_end",
		},
		{"not a URL boundary", `abchttps://example.com/a_b`, `abchttps://example.com/a\_b`},
		{"existing italic behavior", `_italic_`, `\_italic\_`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, escapeMarkdownOne(tc.input))
		})
	}
}

func TestEscapeMarkdownUnmatchedDelimiters(t *testing.T) {
	input := strings.Repeat("[(_中文🙂", 10000)
	want := strings.Repeat("[(\\_中文🙂", 10000)
	assert.Equal(t, want, escapeMarkdownOne(input))
}

func TestRenderMessagesMarkdownTemplates(t *testing.T) {
	p := Plugin{
		Config: Config{Format: formatMarkdown},
		Commit: Commit{Branch: "feature_one", Message: "fix_one"},
		Build:  Build{Tag: "tag_one", Link: "https://example.com/build_one"},
		Repo:   Repo{Name: "repo_one"},
		Tpl:    map[string]string{"value": "value_one"},
	}
	original := p
	original.Tpl = map[string]string{"value": "value_one"}
	input := []string{
		"{{commit.branch}} `{{commit.branch}}` [{{commit.branch}}]({{build.link}})",
		"{{commit.message}} {{build.tag}} {{repo.name}} {{tpl.value}} `{{tpl.value}}`",
		"&#96;entity_code&#96; entity&#95;text",
		"{{replace \"_\" \"-\" commit.branch}}",
		" ",
	}
	want := []string{
		"feature\\_one `feature_one` [feature_one](https://example.com/build_one)",
		"fix\\_one tag\\_one repo\\_one value\\_one `value_one`",
		"`entity_code` entity\\_text",
		"feature-one",
	}
	for range 2 {
		got, err := p.renderMessages(input)
		require.NoError(t, err)
		assert.Equal(t, want, got)
		assert.Equal(t, original, p, "rendering must not mutate template fields")
	}
}

func TestRenderMessagesOtherFormats(t *testing.T) {
	for _, format := range []string{formatHTML, "", "markdown", "MarkdownV2"} {
		t.Run(format, func(t *testing.T) {
			p := Plugin{
				Config: Config{Format: format},
				Tpl:    map[string]string{"value": "value_one"},
			}
			got, err := p.renderMessages([]string{" <code>{{tpl.value}}</code> text_name &amp; "})
			require.NoError(t, err)
			assert.Equal(t, []string{"<code>value_one</code> text_name &"}, got)
		})
	}
}

func TestRenderMessagesFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "message.txt")
	require.NoError(t, os.WriteFile(path, []byte("  {{tpl.value}} `{{tpl.value}}`\n"), 0o600))
	messages, err := loadTextFromFile(path)
	require.NoError(t, err)
	p := Plugin{
		Config: Config{Format: formatMarkdown},
		Tpl:    map[string]string{"value": "file_value"},
	}
	got, err := p.renderMessages(messages)
	require.NoError(t, err)
	assert.Equal(t, []string{"file\\_value `file_value`"}, got)
}

func TestRenderMessagesDefaultMarkdown(t *testing.T) {
	p := Plugin{
		Config: Config{Format: formatMarkdown},
		Repo:   Repo{FullName: "owner_name/repo_name", Namespace: "owner_name"},
		Commit: Commit{Author: "author_name", Branch: "feature_one", Message: "fix_one"},
		Build:  Build{Number: 101, Status: "success", Link: "https://example.com/build_one"},
		GitHub: GitHub{Workflow: "test_workflow", EventName: "pull_request"},
	}
	got, err := p.renderMessages(p.Message())
	require.NoError(t, err)
	assert.Equal(
		t,
		[]string{
			"✅ Build #101 of `owner_name/repo_name` success.\n\n📝 Commit by author\\_name on `feature_one`:\n``` fix_one ```\n\n🌐 https://example.com/build_one",
		},
		got,
	)
	p.Config.GitHub = true
	got, err = p.renderMessages(p.Message())
	require.NoError(t, err)
	assert.Equal(
		t,
		[]string{`owner\_name/repo\_name/test\_workflow triggered by owner\_name (pull\_request)`},
		got,
	)
}

func TestRenderMessagesErrors(t *testing.T) {
	p := Plugin{Config: Config{Format: formatMarkdown}}
	got, err := p.renderMessages([]string{"valid_text", "{{#if}}"})
	require.Error(t, err)
	assert.Nil(t, got, "a failed template must not return partial messages")

	p.Config.Token = "unused"
	p.Config.To = []string{"1"}
	p.Config.MessageFile = filepath.Join(t.TempDir(), "missing.txt")
	err = p.Exec()
	require.Error(t, err)
	require.ErrorContains(t, err, "error loading message file")
	require.ErrorIs(t, err, os.ErrNotExist)
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
	skipIfNoTelegramSecrets(t)
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
	skipIfNoTelegramSecrets(t)
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
	skipIfNoTelegramSecrets(t)
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
	skipIfNoTelegramSecrets(t)
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
	skipIfNoTelegramSecrets(t)
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
`, plugin,
	)
	assert.NoError(t, err)
}
