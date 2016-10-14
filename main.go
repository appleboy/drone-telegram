package main

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"
)

// Version for command line
var Version string

func main() {
	app := cli.NewApp()
	app.Name = "telegram plugin"
	app.Usage = "telegram plugin"
	app.Action = run
	app.Version = Version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "token",
			Usage:  "telegram token",
			EnvVar: "PLUGIN_TOKEN,TELEGRAM_TOKEN",
		},
		cli.StringSliceFlag{
			Name:   "to",
			Usage:  "telegram user",
			EnvVar: "PLUGIN_TO",
		},
		cli.StringSliceFlag{
			Name:   "message",
			Usage:  "send telegram message",
			EnvVar: "PLUGIN_MESSAGE",
		},
		cli.StringSliceFlag{
			Name:   "photo",
			Usage:  "send photo message",
			EnvVar: "PLUGIN_PHOTO",
		},
		cli.StringSliceFlag{
			Name:   "document",
			Usage:  "send document message",
			EnvVar: "PLUGIN_DOCUMENT",
		},
		cli.StringSliceFlag{
			Name:   "sticker",
			Usage:  "send sticker message",
			EnvVar: "PLUGIN_STICKER",
		},
		cli.StringSliceFlag{
			Name:   "audio",
			Usage:  "send audio message",
			EnvVar: "PLUGIN_AUDIO",
		},
		cli.StringSliceFlag{
			Name:   "voice",
			Usage:  "send voice message",
			EnvVar: "PLUGIN_VOICE",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "enable debug message",
			EnvVar: "PLUGIN_DEBUG",
		},
		cli.StringFlag{
			Name:   "format",
			Value:  "markdown",
			Usage:  "telegram message format",
			EnvVar: "PLUGIN_FORMAT",
		},
		cli.StringFlag{
			Name:   "repo.owner",
			Usage:  "repository owner",
			EnvVar: "DRONE_REPO_OWNER",
		},
		cli.StringFlag{
			Name:   "repo.name",
			Usage:  "repository name",
			EnvVar: "DRONE_REPO_NAME",
		},
		cli.StringFlag{
			Name:   "commit.sha",
			Usage:  "git commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
		},
		cli.StringFlag{
			Name:   "commit.branch",
			Value:  "master",
			Usage:  "git commit branch",
			EnvVar: "DRONE_COMMIT_BRANCH",
		},
		cli.StringFlag{
			Name:   "commit.author",
			Usage:  "git author name",
			EnvVar: "DRONE_COMMIT_AUTHOR",
		},
		cli.StringFlag{
			Name:   "commit.message",
			Usage:  "commit message",
			EnvVar: "DRONE_COMMIT_MESSAGE",
		},
		cli.StringFlag{
			Name:   "build.event",
			Value:  "push",
			Usage:  "build event",
			EnvVar: "DRONE_BUILD_EVENT",
		},
		cli.IntFlag{
			Name:   "build.number",
			Usage:  "build number",
			EnvVar: "DRONE_BUILD_NUMBER",
		},
		cli.StringFlag{
			Name:   "build.status",
			Usage:  "build status",
			Value:  "success",
			EnvVar: "DRONE_BUILD_STATUS",
		},
		cli.StringFlag{
			Name:   "build.link",
			Usage:  "build link",
			EnvVar: "DRONE_BUILD_LINK",
		},
	}
	app.Run(os.Args)
}

func run(c *cli.Context) error {
	plugin := Plugin{
		Repo: Repo{
			Owner: c.String("repo.owner"),
			Name:  c.String("repo.name"),
		},
		Build: Build{
			Number:  c.Int("build.number"),
			Event:   c.String("build.event"),
			Status:  c.String("build.status"),
			Commit:  c.String("commit.sha"),
			Branch:  c.String("commit.branch"),
			Author:  c.String("commit.author"),
			Message: c.String("commit.message"),
			Link:    c.String("build.link"),
		},
		Config: Config{
			Token:    c.String("token"),
			Debug:    c.Bool("debug"),
			To:       c.StringSlice("to"),
			Message:  c.StringSlice("message"),
			Photo:    c.StringSlice("photo"),
			Document: c.StringSlice("document"),
			Sticker:  c.StringSlice("sticker"),
			Audio:    c.StringSlice("audio"),
			Voice:    c.StringSlice("voice"),
			Format:   c.String("format"),
		},
	}

	return plugin.Exec()
}
