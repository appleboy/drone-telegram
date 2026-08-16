# drone-telegram

![logo](./images/logo.png)

[![GoDoc](https://pkg.go.dev/badge/github.com/appleboy/drone-telegram.svg)](https://pkg.go.dev/github.com/appleboy/drone-telegram)
[![Trivy Security Scan](https://github.com/appleboy/drone-telegram/actions/workflows/trivy.yml/badge.svg?branch=master)](https://github.com/appleboy/drone-telegram/actions/workflows/trivy.yml)
[![codecov](https://codecov.io/gh/appleboy/drone-telegram/branch/master/graph/badge.svg)](https://codecov.io/gh/appleboy/drone-telegram)
[![GitHub release](https://img.shields.io/github/v/release/appleboy/drone-telegram)](https://github.com/appleboy/drone-telegram/releases)

[Drone](https://github.com/harness/drone) plugin for sending telegram notifications. For the usage
information and a listing of the available options please take a look at [DOCS.md](DOCS.md) or
[the docs](https://plugins.drone.io/plugins/telegram).

Using GitHub Actions instead? See [appleboy/telegram-action](https://github.com/appleboy/telegram-action).

## Features

* Send text message in `markdown` or `html` format
* Send photo, document, audio, voice, video and sticker messages
* Send location and venue messages
* Send message to a forum topic via `message_thread_id`
* Customize the message with a [template](DOCS.md) and `template_vars` / `template_vars_file`
* Load the message from a file with `message_file`
* Filter notifications by commit author email with `only_match_email`
* Disable notification sound (`disable_notification`) or link preview (`disable_web_page_preview`)
* Connect through a SOCKS5 proxy
* Load all settings from an env file with `env_file`

## Build or Download a binary

The pre-compiled binaries can be downloaded from [release page](https://github.com/appleboy/drone-telegram/releases). Support the following OS type.

* Linux amd64/arm/arm64
* Darwin amd64/arm64
* Windows amd64
* FreeBSD amd64

With `Go` installed

```sh
go install github.com/appleboy/drone-telegram@latest
```

or build the binary with the following command:

```sh
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

go build -v -a -tags netgo -o release/linux/amd64/drone-telegram .
```

## Testing

Test the package with the following command:

```sh
make test
```

## Usage

### Drone pipeline

Add the plugin as a step in your `.drone.yml`:

```yaml
- name: send telegram notification
  image: appleboy/drone-telegram
  settings:
    token:
      from_secret: telegram_token
    to:
      from_secret: telegram_to
    message: >
      {{#success build.status}}
        build {{build.number}} succeeded. Good job.
      {{else}}
        build {{build.number}} failed. Fix me please.
      {{/success}}
```

See [DOCS.md](DOCS.md) for all available settings and template variables.

### Docker

Send a message with the minimum required settings:

```sh
docker run --rm \
  -e PLUGIN_TOKEN=xxxxxxx \
  -e PLUGIN_TO=xxxxxxx \
  -e PLUGIN_MESSAGE=test \
  appleboy/drone-telegram
```

Full example with all message types and build metadata:

```sh
docker run --rm \
  -e PLUGIN_TOKEN=xxxxxxx \
  -e PLUGIN_TO=xxxxxxx \
  -e PLUGIN_MESSAGE=test \
  -e PLUGIN_MESSAGE_FILE=testmessage.md \
  -e PLUGIN_PHOTO=tests/github.png \
  -e PLUGIN_DOCUMENT=tests/gophercolor.png \
  -e PLUGIN_STICKER=tests/github-logo.png \
  -e PLUGIN_AUDIO=tests/audio.mp3 \
  -e PLUGIN_VOICE=tests/voice.ogg \
  -e PLUGIN_LOCATION="24.9163213 121.1424972" \
  -e PLUGIN_VENUE="24.9163213 121.1424972 title address" \
  -e PLUGIN_VIDEO=tests/video.mp4 \
  -e PLUGIN_DEBUG=true \
  -e PLUGIN_ONLY_MATCH_EMAIL=false \
  -e PLUGIN_FORMAT=markdown \
  -e DRONE_REPO_OWNER=appleboy \
  -e DRONE_REPO_NAME=go-hello \
  -e DRONE_COMMIT_SHA=e5e82b5eb3737205c25955dcc3dcacc839b7be52 \
  -e DRONE_COMMIT_BRANCH=master \
  -e DRONE_COMMIT_LINK=https://github.com/appleboy/go-hello/compare/master... \
  -e DRONE_COMMIT_AUTHOR=appleboy \
  -e DRONE_COMMIT_AUTHOR_EMAIL=appleboy@gmail.com \
  -e DRONE_BUILD_NUMBER=1 \
  -e DRONE_BUILD_STATUS=success \
  -e DRONE_BUILD_LINK=http://github.com/appleboy/go-hello \
  -e DRONE_TAG=1.0.0 \
  -e DRONE_JOB_STARTED=1477550550 \
  -e DRONE_JOB_FINISHED=1477550750 \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  appleboy/drone-telegram
```

Load all environments from file.

```bash
docker run --rm \
  -e PLUGIN_ENV_FILE=your_env_file_path \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  appleboy/drone-telegram
```

## License

This project is licensed under the [MIT License](LICENSE).
