package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/appleboy/drone-facebook/template"
	"gopkg.in/telegram-bot-api.v4"
)

type (
	// Repo information.
	Repo struct {
		Owner string
		Name  string
	}

	// Build information.
	Build struct {
		Tag      string
		Event    string
		Number   int
		Commit   string
		Message  string
		Branch   string
		Author   string
		Email    string
		Status   string
		Link     string
		Started  float64
		Finished float64
		PR       string
	}

	// Config for the plugin.
	Config struct {
		Token      string
		Debug      bool
		MatchEmail bool
		To         []string
		Message    []string
		Photo      []string
		Document   []string
		Sticker    []string
		Audio      []string
		Voice      []string
		Location   []string
		Video      []string
		Venue      []string
		Format     string
	}

	// Plugin values.
	Plugin struct {
		Repo   Repo
		Build  Build
		Config Config
	}

	// Location format
	Location struct {
		Title     string
		Address   string
		Latitude  float64
		Longitude float64
	}
)

func trimElement(keys []string) []string {
	var newKeys []string

	for _, value := range keys {
		value = strings.Trim(value, " ")
		if len(value) == 0 {
			continue
		}
		newKeys = append(newKeys, value)
	}

	return newKeys
}

func escapeMarkdown(keys []string) []string {
	var newKeys []string

	for _, value := range keys {
		value = escapeMarkdownOne(value)
		if len(value) == 0 {
			continue
		}
		newKeys = append(newKeys, value)
	}

	return newKeys
}

func escapeMarkdownOne(str string) string {
	str = strings.Replace(str, `\_`, `_`, -1)
	str = strings.Replace(str, `_`, `\_`, -1)

	return str
}

func fileExist(keys []string) []string {
	var newKeys []string

	for _, value := range keys {
		if _, err := os.Stat(value); os.IsNotExist(err) {
			continue
		}
		newKeys = append(newKeys, value)
	}

	return newKeys
}

func convertLocation(value string) (Location, bool) {
	var latitude, longitude float64
	var title, address string
	var err error
	values := trimElement(strings.Split(value, ","))

	if len(values) < 2 {
		return Location{}, true
	}

	if len(values) > 2 {
		title = values[2]
	}

	if len(values) > 3 {
		title = values[2]
		address = values[3]
	}

	latitude, err = strconv.ParseFloat(values[0], 64)

	if err != nil {
		log.Println(err.Error())
		return Location{}, true
	}

	longitude, err = strconv.ParseFloat(values[1], 64)

	if err != nil {
		log.Println(err.Error())
		return Location{}, true
	}

	return Location{
		Title:     title,
		Address:   address,
		Latitude:  latitude,
		Longitude: longitude,
	}, false
}

func parseTo(to []string, authorEmail string, matchEmail bool) []int64 {
	var emails []int64
	var ids []int64
	attachEmail := true

	for _, value := range trimElement(to) {
		idArray := trimElement(strings.Split(value, ":"))

		// check id
		id, err := strconv.ParseInt(idArray[0], 10, 64)
		if err != nil {
			continue
		}

		// check match author email
		if len(idArray) > 1 {
			if email := idArray[1]; email != authorEmail {
				continue
			}

			emails = append(emails, id)
			attachEmail = false
			continue
		}

		ids = append(ids, id)
	}

	if matchEmail == true && attachEmail == false {
		return emails
	}

	for _, value := range emails {
		ids = append(ids, value)
	}

	return ids
}

// Exec executes the plugin.
func (p Plugin) Exec() error {

	if len(p.Config.Token) == 0 || len(p.Config.To) == 0 {
		log.Println("missing telegram token or user list")

		return errors.New("missing telegram token or user list")
	}

	var message []string
	if len(p.Config.Message) > 0 {
		message = p.Config.Message
	} else {
		message = p.Message(p.Repo, p.Build)
	}

	bot, err := tgbotapi.NewBotAPI(p.Config.Token)

	if err != nil {
		log.Println("Initialize New Bot Error:", err.Error())

		return err
	}

	bot.Debug = p.Config.Debug

	ids := parseTo(p.Config.To, p.Build.Email, p.Config.MatchEmail)
	photos := fileExist(trimElement(p.Config.Photo))
	documents := fileExist(trimElement(p.Config.Document))
	stickers := fileExist(trimElement(p.Config.Sticker))
	audios := fileExist(trimElement(p.Config.Audio))
	voices := fileExist(trimElement(p.Config.Voice))
	videos := fileExist(trimElement(p.Config.Video))
	locations := trimElement(p.Config.Location)
	venues := trimElement(p.Config.Venue)

	message = trimElement(message)

	if p.Config.Format == "markdown" {
		message = escapeMarkdown(message)

		p.Build.Message = escapeMarkdownOne(p.Build.Author)
		p.Build.Branch = escapeMarkdownOne(p.Build.Branch)
		p.Build.Author = escapeMarkdownOne(p.Build.Author)
		p.Build.Email = escapeMarkdownOne(p.Build.Email)
		p.Build.Link = escapeMarkdownOne(p.Build.Link)
		p.Build.PR = escapeMarkdownOne(p.Build.PR)

		p.Repo.Owner = escapeMarkdownOne(p.Repo.Owner)
		p.Repo.Name = escapeMarkdownOne(p.Repo.Name)
	}

	// send message.
	for _, user := range ids {
		for _, value := range message {
			txt, err := template.RenderTrim(value, p)
			if err != nil {
				return err
			}

			msg := tgbotapi.NewMessage(user, txt)
			msg.ParseMode = p.Config.Format
			p.Send(bot, msg)
		}

		for _, value := range photos {
			msg := tgbotapi.NewPhotoUpload(user, value)
			p.Send(bot, msg)
		}

		for _, value := range documents {
			msg := tgbotapi.NewDocumentUpload(user, value)
			p.Send(bot, msg)
		}

		for _, value := range stickers {
			msg := tgbotapi.NewStickerUpload(user, value)
			p.Send(bot, msg)
		}

		for _, value := range audios {
			msg := tgbotapi.NewAudioUpload(user, value)
			msg.Title = "Audio Message."
			p.Send(bot, msg)
		}

		for _, value := range voices {
			msg := tgbotapi.NewVoiceUpload(user, value)
			p.Send(bot, msg)
		}

		for _, value := range videos {
			msg := tgbotapi.NewVideoUpload(user, value)
			msg.Caption = "Video Message"
			p.Send(bot, msg)
		}

		for _, value := range locations {
			location, empty := convertLocation(value)

			if empty == true {
				continue
			}

			msg := tgbotapi.NewLocation(user, location.Latitude, location.Longitude)
			p.Send(bot, msg)
		}

		for _, value := range venues {
			location, empty := convertLocation(value)

			if empty == true {
				continue
			}

			msg := tgbotapi.NewVenue(user, location.Title, location.Address, location.Latitude, location.Longitude)
			p.Send(bot, msg)
		}
	}

	return nil
}

// Send bot message.
func (p Plugin) Send(bot *tgbotapi.BotAPI, msg tgbotapi.Chattable) {
	_, err := bot.Send(msg)

	if err != nil {
		log.Println(err.Error())
	}
}

// Message is plugin default message.
func (p Plugin) Message(repo Repo, build Build) []string {
	return []string{fmt.Sprintf("[%s] <%s> (%s)『%s』by %s",
		build.Status,
		build.Link,
		build.Branch,
		build.Message,
		build.Author,
	)}
}
