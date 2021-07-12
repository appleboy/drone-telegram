package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/appleboy/drone-template-lib/template"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	formatMarkdown = "Markdown"
	formatHTML     = "HTML"
)

type (
	// GitHub information.
	GitHub struct {
		Workflow  string
		Workspace string
		Action    string
		EventName string
		EventPath string
	}

	// Repo information.
	Repo struct {
		FullName  string
		Namespace string
		Name      string
	}

	// Commit information.
	Commit struct {
		Sha     string
		Ref     string
		Branch  string
		Link    string
		Author  string
		Avatar  string
		Email   string
		Message string
	}

	// Build information.
	Build struct {
		Tag      string
		Event    string
		Number   int
		Status   string
		Link     string
		Started  int64
		Finished int64
		PR       string
		DeployTo string
	}

	// Config for the plugin.
	Config struct {
		Token            string
		Debug            bool
		MatchEmail       bool
		To               []string
		Message          string
		MessageFile      string
		TemplateVarsFile string
		TemplateVars     string
		Photo            []string
		Document         []string
		Sticker          []string
		Audio            []string
		Voice            []string
		Location         []string
		Video            []string
		Venue            []string
		Format           string
		GitHub           bool
		Socks5           string

		DisableWebPagePreview bool
		DisableNotification   bool
	}

	// Plugin values.
	Plugin struct {
		GitHub GitHub
		Repo   Repo
		Commit Commit
		Build  Build
		Config Config
		Tpl    map[string]string
	}

	// Location format
	Location struct {
		Title     string
		Address   string
		Latitude  float64
		Longitude float64
	}
)

var icons = map[string]string{
	"failure":   "❌",
	"cancelled": "❕",
	"success":   "✅",
}

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

func globList(keys []string) []string {
	var newKeys []string

	for _, pattern := range keys {
		pattern = strings.Trim(pattern, " ")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("Glob error for %q: %s\n", pattern, err)
			continue
		}
		newKeys = append(newKeys, matches...)
	}

	return newKeys
}

func convertLocation(value string) (Location, bool) {
	var latitude, longitude float64
	var title, address string
	var err error
	values := trimElement(strings.Split(value, " "))

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

func loadTextFromFile(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return []string{string(content)}, nil
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

	if matchEmail && !attachEmail {
		return emails
	}

	ids = append(ids, emails...)

	return ids
}

func templateMessage(t string, plugin Plugin) (string, error) {
	return template.RenderTrim(t, plugin)
}

// Exec executes the plugin.
func (p Plugin) Exec() (err error) {
	if len(p.Config.Token) == 0 || len(p.Config.To) == 0 {
		return errors.New("missing telegram token or user list")
	}

	var message []string
	switch {
	case len(p.Config.MessageFile) > 0:
		message, err = loadTextFromFile(p.Config.MessageFile)
		if err != nil {
			return fmt.Errorf("error loading message file '%s': %v", p.Config.MessageFile, err)
		}
	case len(p.Config.Message) > 0:
		message = []string{p.Config.Message}
	default:
		p.Config.Format = formatMarkdown
		message = p.Message()
	}

	if p.Config.TemplateVars != "" {
		p.Tpl = make(map[string]string)
		if err = json.Unmarshal([]byte(p.Config.TemplateVars), &p.Tpl); err != nil {
			return fmt.Errorf("unable to unmarshall template vars from JSON string '%s': %v", p.Config.TemplateVars, err)
		}
	}

	if p.Config.TemplateVarsFile != "" {
		content, err := ioutil.ReadFile(p.Config.TemplateVarsFile)
		if err != nil {
			return fmt.Errorf("unable to read file with template vars '%s': %v", p.Config.TemplateVarsFile, err)
		}
		vars := make(map[string]string)
		if err = json.Unmarshal(content, &vars); err != nil {
			return fmt.Errorf("unable to unmarshall template vars from JSON file '%s': %v", p.Config.TemplateVarsFile, err)
		}
		// Merging templates variables from file to the variables form plugin settings (variables from file takes precedence)
		if p.Tpl == nil {
			p.Tpl = vars
		} else {
			for k, v := range vars {
				p.Tpl[k] = v
			}
		}
	}

	var proxyURL *url.URL
	if proxyURL, err = url.Parse(p.Config.Socks5); err != nil {
		return fmt.Errorf("unable to unmarshall socks5 proxy url from string '%s': %v", p.Config.Socks5, err)
	}

	var bot *tgbotapi.BotAPI
	if len(p.Config.Socks5) > 0 {
		proxyClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
		bot, err = tgbotapi.NewBotAPIWithClient(p.Config.Token, proxyClient)
	} else {
		bot, err = tgbotapi.NewBotAPI(p.Config.Token)
	}

	if err != nil {
		return err
	}

	bot.Debug = p.Config.Debug

	ids := parseTo(p.Config.To, p.Commit.Email, p.Config.MatchEmail)
	photos := globList(trimElement(p.Config.Photo))
	documents := globList(trimElement(p.Config.Document))
	stickers := globList(trimElement(p.Config.Sticker))
	audios := globList(trimElement(p.Config.Audio))
	voices := globList(trimElement(p.Config.Voice))
	videos := globList(trimElement(p.Config.Video))
	locations := trimElement(p.Config.Location)
	venues := trimElement(p.Config.Venue)

	message = trimElement(message)

	if p.Config.Format == formatMarkdown {
		message = escapeMarkdown(message)

		p.Commit.Message = escapeMarkdownOne(p.Commit.Message)
		p.Commit.Branch = escapeMarkdownOne(p.Commit.Branch)
		p.Commit.Link = escapeMarkdownOne(p.Commit.Link)
		p.Commit.Author = escapeMarkdownOne(p.Commit.Author)
		p.Commit.Email = escapeMarkdownOne(p.Commit.Email)

		p.Build.Tag = escapeMarkdownOne(p.Build.Tag)
		p.Build.Link = escapeMarkdownOne(p.Build.Link)
		p.Build.PR = escapeMarkdownOne(p.Build.PR)

		p.Repo.Namespace = escapeMarkdownOne(p.Repo.Namespace)
		p.Repo.Name = escapeMarkdownOne(p.Repo.Name)
	}

	// send message.
	for _, user := range ids {
		for _, value := range message {
			txt, err := templateMessage(value, p)
			if err != nil {
				return err
			}

			txt = html.UnescapeString(txt)
			log.Println(txt)
			log.Println(p.Config.Format)

			msg := tgbotapi.NewMessage(user, txt)
			msg.ParseMode = p.Config.Format
			msg.DisableWebPagePreview = p.Config.DisableWebPagePreview
			msg.DisableNotification = p.Config.DisableNotification
			if err := p.Send(bot, msg); err != nil {
				return err
			}
		}

		for _, value := range photos {
			msg := tgbotapi.NewPhotoUpload(user, value)
			if err := p.Send(bot, msg); err != nil {
				return err
			}
		}

		for _, value := range documents {
			msg := tgbotapi.NewDocumentUpload(user, value)
			if err := p.Send(bot, msg); err != nil {
				return err
			}
		}

		for _, value := range stickers {
			msg := tgbotapi.NewStickerUpload(user, value)
			if err := p.Send(bot, msg); err != nil {
				return err
			}
		}

		for _, value := range audios {
			msg := tgbotapi.NewAudioUpload(user, value)
			msg.Title = "Audio Message."
			if err := p.Send(bot, msg); err != nil {
				return err
			}
		}

		for _, value := range voices {
			msg := tgbotapi.NewVoiceUpload(user, value)
			if err := p.Send(bot, msg); err != nil {
				return err
			}
		}

		for _, value := range videos {
			msg := tgbotapi.NewVideoUpload(user, value)
			msg.Caption = "Video Message"
			if err := p.Send(bot, msg); err != nil {
				return err
			}
		}

		for _, value := range locations {
			location, empty := convertLocation(value)

			if empty {
				continue
			}

			msg := tgbotapi.NewLocation(user, location.Latitude, location.Longitude)
			if err := p.Send(bot, msg); err != nil {
				return err
			}
		}

		for _, value := range venues {
			location, empty := convertLocation(value)

			if empty {
				continue
			}

			msg := tgbotapi.NewVenue(user, location.Title, location.Address, location.Latitude, location.Longitude)
			if err := p.Send(bot, msg); err != nil {
				return err
			}
		}
	}

	return nil
}

// Send bot message.
func (p Plugin) Send(bot *tgbotapi.BotAPI, msg tgbotapi.Chattable) error {
	message, err := bot.Send(msg)

	if p.Config.Debug {
		log.Println("=====================")
		log.Printf("Response Message: %#v\n", message)
		log.Println("=====================")
	}

	if err == nil {
		return nil
	}

	return errors.New(strings.Replace(err.Error(), p.Config.Token, "<token>", -1))
}

// Message is plugin default message.
func (p Plugin) Message() []string {
	icon := icons[strings.ToLower(p.Build.Status)]

	if p.Config.GitHub {
		return []string{fmt.Sprintf("%s/%s triggered by %s (%s)",
			p.Repo.FullName,
			p.GitHub.Workflow,
			p.Repo.Namespace,
			p.GitHub.EventName,
		)}
	}

	// ✅  Build #106 of drone-telegram succeeded.
	//
	// 📝 Commit by appleboy on master:
	//  chore: update default template
	//
	// 🌐 https://cloud.drone.io/appleboy/drone-telegram/106
	return []string{fmt.Sprintf("%s Build #%d of `%s` %s.\n\n📝 Commit by %s on `%s`:\n``` %s ```\n\n🌐 %s",
		icon,
		p.Build.Number,
		p.Repo.FullName,
		p.Build.Status,
		p.Commit.Author,
		p.Commit.Branch,
		p.Commit.Message,
		p.Build.Link,
	)}
}
