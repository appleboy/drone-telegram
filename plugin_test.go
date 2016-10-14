package main

import (
	"github.com/stretchr/testify/assert"

	"os"
	"testing"
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
			Name:  "go-hello",
			Owner: "appleboy",
		},
		Build: Build{
			Number:  101,
			Status:  "success",
			Link:    "https://github.com/appleboy/go-hello",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "update travis",
			Commit:  "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
		},
	}

	message := plugin.Message(plugin.Repo, plugin.Build)

	assert.Equal(t, []string{"[success] <https://github.com/appleboy/go-hello> (master)『update travis』by Bo-Yi Wu"}, message)
}

func TestSendMessage(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:  "go-hello",
			Owner: "appleboy",
		},
		Build: Build{
			Number:  101,
			Status:  "success",
			Link:    "https://github.com/appleboy/go-hello",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "update travis by drone plugin",
			Commit:  "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
		},

		Config: Config{
			Token:    os.Getenv("TELEGRAM_TOKEN"),
			To:       []string{os.Getenv("TELEGRAM_TO"), "中文ID", "1234567890"},
			Message:  []string{"Test Telegram Chat Bot From Travis or Local", " "},
			Photo:    []string{"tests/github.png", "1234", " "},
			Document: []string{"tests/gophercolor.png", "1234", " "},
			Sticker:  []string{"tests/github-logo.png", "tests/github.png", "1234", " "},
			Audio:    []string{"tests/audio.mp3", "1234", " "},
			Voice:    []string{"tests/voice.ogg", "1234", " "},
			Location: []string{"24.9163213,121.1424972", "1", " "},
			Venue:    []string{"35.661777,139.704051,竹北體育館,新竹縣竹北市", "24.9163213,121.1424972", "1", " "},
			Video:    []string{"tests/video.mp4", "1234", " "},
			Debug:    false,
		},
	}

	err := plugin.Exec()
	assert.Nil(t, err)

	// disable message
	plugin.Config.Message = []string{}
	err = plugin.Exec()
	assert.Nil(t, err)
}

func TestBotError(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:  "go-hello",
			Owner: "appleboy",
		},
		Build: Build{
			Number:  101,
			Status:  "success",
			Link:    "https://github.com/appleboy/go-hello",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "update travis by drone plugin",
			Commit:  "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
		},

		Config: Config{
			Token:   "appleboy",
			To:      []string{os.Getenv("TELEGRAM_TO"), "中文ID", "1234567890"},
			Message: []string{"Test Telegram Chat Bot From Travis or Local", " "},
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

func TestParseID(t *testing.T) {
	var input []string
	var result []int64

	input = []string{"1", "測試", "3"}
	result = []int64{int64(1), int64(3)}

	assert.Equal(t, result, parseID(input))

	input = []string{"1", "2"}
	result = []int64{int64(1), int64(2)}

	assert.Equal(t, result, parseID(input))
}

func TestCheckFileExist(t *testing.T) {
	var input []string
	var result []string

	input = []string{"tests/gophercolor.png", "測試", "3"}
	result = []string{"tests/gophercolor.png"}

	assert.Equal(t, result, fileExist(input))
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
	input = "測試,139.704051"
	result, empty = convertLocation(input)

	assert.Equal(t, true, empty)
	assert.Equal(t, Location{}, result)

	// strconv.ParseInt: parsing "測試": invalid syntax
	input = "35.661777,測試"
	result, empty = convertLocation(input)

	assert.Equal(t, true, empty)
	assert.Equal(t, Location{}, result)

	input = "35.661777,139.704051"
	result, empty = convertLocation(input)

	assert.Equal(t, false, empty)
	assert.Equal(t, Location{
		Latitude:  float64(35.661777),
		Longitude: float64(139.704051),
	}, result)

	input = "35.661777,139.704051,title"
	result, empty = convertLocation(input)

	assert.Equal(t, false, empty)
	assert.Equal(t, Location{
		Title:     "title",
		Address:   "",
		Latitude:  float64(35.661777),
		Longitude: float64(139.704051),
	}, result)

	input = "35.661777,139.704051,title,address"
	result, empty = convertLocation(input)

	assert.Equal(t, false, empty)
	assert.Equal(t, Location{
		Title:     "title",
		Address:   "address",
		Latitude:  float64(35.661777),
		Longitude: float64(139.704051),
	}, result)
}
