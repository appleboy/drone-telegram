package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
)

func loadRichMessage(filename string) (*tgbotapi.InputRichMessage, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read rich message file: %w", err)
	}
	var message *tgbotapi.InputRichMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, fmt.Errorf("invalid rich message JSON: %w", err)
	}
	count := 0
	if message != nil {
		if message.HTML != "" {
			count++
		}
		if message.Markdown != "" {
			count++
		}
		if len(message.Blocks) > 0 {
			count++
		}
	}
	if count != 1 {
		return nil, errors.New("rich message requires exactly one of html, markdown, or blocks")
	}
	return message, nil
}
