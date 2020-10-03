---
date: 2019-10-19T00:00:00+00:00
title: Telegram
author: appleboy
tags: [ notifications, chat ]
repo: appleboy/drone-telegram
logo: telegram.svg
image: appleboy/drone-telegram
---

The Telegram plugin posts build status messages to your account. The below pipeline configuration demonstrates simple usage:

```yaml
- name: send telegram notification
  image: appleboy/drone-telegram
  settings:
    token: xxxxxxxxxx
    to: telegram_user_id
```

Example configuration with photo message:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     photo:
+       - tests/1.png
+       - tests/2.png
```

Example configuration with document message:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     document:
+       - tests/1.pdf
+       - tests/2.pdf
```

Example configuration with sticker message:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     sticker:
+       - tests/3.png
+       - tests/4.png
```

Example configuration with audio message:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     audio:
+       - tests/audio1.mp3
+       - tests/audio2.mp3
```

Example configuration with voice message:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     voice:
+       - tests/voice1.ogg
+       - tests/voice2.ogg
```

Example configuration with location message:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     location:
+       - 24.9163213,121.1424972
+       - 24.9263213,121.1224972
```

Example configuration with venue message:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     venue:
+       - 24.9163213,121.1424972,title,address
+       - 24.3163213,121.1824972,title,address
```

Example configuration with video message:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     video:
+       - tests/video1.mp4
+       - tests/video2.mp4
```

Example configuration with message format:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     format: markdown
```

Example configuration with a custom message template:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     message: >
+       {{#success build.status}}
+         build {{build.number}} succeeded. Good job.
+       {{else}}
+         build {{build.number}} failed. Fix me please.
+       {{/success}}
```

Example configuration with a custom message template loaded from file:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     message_file: message_file.tpl
```

Example configuration with a generic message template loaded from file, with additional extra vars:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     message_file: message_file.tpl
+     template_vars:
+       env: testing
+       app: MyApp
```

Where `message_file.tpl` is:

```bash
Build finished for *{{tpl.app}}* - *{{tpl.env}}*

{{#success build.status}}
  build {{build.number}} succeeded. Good job.
{{else}}
  build {{build.number}} failed. Fix me please.
{{/success}}
```

Example configuration with a custom message template, with extra vars loaded from file (e.g. from previous build steps):

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
+     template_vars_file: build_report.json
+     message: >
+       {{#success build.status}}
+         build {{build.number}} succeeded, artefact version = {{tpl.artefact_version}}.
+       {{else}}
+         build {{build.number}} failed. Fix me please.
+       {{/success}}
```

Where `build_report.json` is:

```
{
  ...
  "artefact_version": "0.2.3452"
  ...
}
```

Example configuration with a custom socks5 URL:

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
      message: send message using custom socks5 URL
+     socks5: socks5://67.204.21.1:64312
```

Disables link previews for links in this message

```diff
  - name: send telegram notification
    image: appleboy/drone-telegram
    settings:
      token: xxxxxxxxxx
      to: telegram_user_id
      message: send message using custom socks5 URL
+     disable_web_page_preview: true
```

## Parameter Reference

token
: telegram token from [telegram developer center](https://core.telegram.org/bots/api)

to
: telegram user id (can be requested from the @userinfobot inside Telegram)

message
: overwrite the default message template

message_file
: overwrite the default message template with the contents of the specified file

template_vars
: define additional template vars. Example: `var1: hello` can be used within the template as `tpl.var1`

template_vars_file
: load additional template vars from json file. Example: given file content `{"var1":"hello"}`, variable can be used within the template as `tpl.var1`

photo
: local file path

document
: local file path

sticker
: local file path

audio
: local file path

voice
: local file path

location
: local file path

video
: local file path

venue
: local file path

format
: `markdown` or `html` format

## Template Reference

repo.owner
: repository owner

repo.name
: repository name

commit.sha
: git sha for current commit

commit.branch
: git branch for current commit

commit.link
: git commit link in remote

commit.author
: git author for current commit

commit.email
: git author email for current commit

commit.message
: git current commit message

build.status
: build status type enumeration, either `success` or `failure`

build.event
: build event type enumeration, one of `push`, `pull_request`, `tag`, `deployment`

build.number
: build number

build.tag
: git tag for current commit

build.link
: link the the build results in drone

build.started
: unix timestamp for build started

build.finished
: unix timestamp for build finished

## Template Function Reference

uppercasefirst
: converts the first letter of a string to uppercase

uppercase
: converts a string to uppercase

lowercase
: converts a string to lowercase. Example `{{lowercase build.author}}`

datetime
: converts a unix timestamp to a date time string. Example `{{datetime build.started}}`

success
: returns true if the build is successful

failure
: returns true if the build is failed

truncate
: returns a truncated string to n characters. Example `{{truncate build.sha 8}}`

urlencode
: returns a url encoded string

since
: returns a duration string between now and the given timestamp. Example `{{since build.started}}`
