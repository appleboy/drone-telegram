package main

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
)

const maxWidgetFile = 1 << 20

// loadWidget returns Bot API parameters without delivery settings. Keeping the
// request in the existing client avoids a second transport or a library fork.
func loadWidget(filename string) (tgbotapi.Params, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("cannot read widget_file")
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxWidgetFile+1))
	if err != nil || len(data) > maxWidgetFile || !utf8.Valid(data) {
		return nil, errors.New("widget_file must be UTF-8 JSON of at most 1 MiB")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return nil, errors.New("widget_file must contain a JSON object")
	}
	params := tgbotapi.Params{}
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return nil, errors.New("invalid widget JSON")
		}
		name, ok := key.(string)
		if !ok {
			return nil, errors.New("invalid widget JSON field")
		}
		if _, exists := params[name]; exists {
			return nil, errors.New("duplicate widget JSON field")
		}
		var raw json.RawMessage
		if d.Decode(&raw) != nil || bytes.Equal(raw, []byte("null")) {
			return nil, errors.New("invalid widget JSON value")
		}
		switch name {
		case "html", "public_data", "chat_data", "caption", "parse_mode", "message_effect_id":
			var value string
			if json.Unmarshal(raw, &value) != nil {
				return nil, errors.New("widget string field has invalid type")
			}
			params[name] = value
		case "width", "height":
			var value int
			if json.Unmarshal(raw, &value) != nil {
				return nil, errors.New("widget dimension must be an integer")
			}
			params[name] = strconv.Itoa(value)
		case "show_caption_above_media",
			"allow_queries",
			"is_shared",
			"protect_content",
			"allow_paid_broadcast":
			var value bool
			if json.Unmarshal(raw, &value) != nil {
				return nil, errors.New("widget flag must be a boolean")
			}
			params[name] = strconv.FormatBool(value)
		case "reply_parameters", "reply_markup":
			var value map[string]json.RawMessage
			if json.Unmarshal(raw, &value) != nil {
				return nil, errors.New("widget reply field must be an object")
			}
			params[name] = string(raw)
		case "caption_entities":
			var value []map[string]json.RawMessage
			if json.Unmarshal(raw, &value) != nil {
				return nil, errors.New("caption_entities must be an array of objects")
			}
			for _, entity := range value {
				if entity == nil {
					return nil, errors.New("caption entity cannot be null")
				}
			}
			params[name] = string(raw)
		default:
			return nil, errors.New(
				"unsupported widget field; credentials and delivery settings belong in plugin inputs",
			)
		}
	}
	if _, err := d.Token(); err != nil {
		return nil, errors.New("invalid widget JSON")
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, errors.New("unexpected data after widget JSON")
	}
	html := params["html"]
	if len(html) == 0 || len(html) > 32768 {
		return nil, errors.New("widget HTML must contain 1-32768 UTF-8 bytes")
	}
	var compressed bytes.Buffer
	writer, err := zlib.NewWriterLevel(&compressed, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := writer.Write([]byte(html)); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	if compressed.Len() > 4096 {
		return nil, errors.New("widget HTML exceeds 4096 compressed bytes")
	}
	if len(params["public_data"]) > 1024 || len(params["chat_data"]) > 1024 {
		return nil, errors.New("widget state exceeds 1024 UTF-8 bytes")
	}
	mode, hasMode := params["parse_mode"]
	if hasMode {
		if _, ok := params["caption_entities"]; ok {
			return nil, errors.New("use parse_mode or caption_entities, not both")
		}
		if mode != "HTML" && mode != "Markdown" && mode != "MarkdownV2" {
			return nil, errors.New("unsupported caption parse_mode")
		}
	} else if utf8.RuneCountInString(params["caption"]) > 1024 {
		return nil, errors.New("caption exceeds 1024 characters")
	}
	return params, nil
}

func (p *Plugin) prepareWidget() (tgbotapi.Params, error) {
	c := p.Config
	if c.Message != "" || c.MessageFile != "" || c.RichMessageFile != "" ||
		c.TemplateVars != "" || c.TemplateVarsFile != "" || c.Format != "" || c.DisableWebPagePreview ||
		len(
			c.Photo,
		)+len(
			c.Document,
		)+len(
			c.Sticker,
		)+len(
			c.Audio,
		)+len(
			c.Voice,
		)+len(
			c.Video,
		)+len(
			c.Location,
		)+len(
			c.Venue,
		) != 0 {
		return nil, errors.New(
			"widget_file cannot be combined with message, rich message, media, template, format or preview inputs",
		)
	}
	if len(c.To) != 1 {
		return nil, errors.New("widget_file requires exactly one destination")
	}
	destination := strings.TrimSpace(strings.SplitN(c.To[0], ":", 2)[0])
	if id, err := strconv.ParseInt(destination, 10, 64); err != nil || id == 0 {
		return nil, errors.New("widget_file requires a numeric nonzero destination")
	}
	if len(parseTo(c.To, p.Commit.Email, c.MatchEmail)) != 1 {
		return nil, errors.New("widget_file requires exactly one destination")
	}
	if c.MessageThreadID < 0 {
		return nil, errors.New("message_thread_id must be positive when set")
	}
	return loadWidget(c.WidgetFile)
}

func (p *Plugin) sendWidget(bot *tgbotapi.BotAPI, params tgbotapi.Params, user int64) error {
	params["chat_id"] = strconv.FormatInt(user, 10)
	if p.Config.MessageThreadID != 0 {
		params["message_thread_id"] = strconv.Itoa(p.Config.MessageThreadID)
	}
	params["disable_notification"] = strconv.FormatBool(p.Config.DisableNotification)
	// The client sends once. Do not expose server descriptions or transport errors:
	// either may contain the credential-bearing URL or caller-owned content.
	if _, err := bot.MakeRequest("sendWidget", params); err != nil {
		return errors.New("sendWidget failed; delivery may be unknown; no retry attempted")
	}
	return nil
}
