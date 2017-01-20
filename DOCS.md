---
date: 2017-01-08T00:00:00+00:00
title: Telegram
author: appleboy
tags: [ notifications, chat ]
repo: appleboy/drone-telegram
logo: telegram.svg
image: appleboy/drone-telegram
---

The Telegram plugin posts build status messages to your account. The below pipeline configuration demonstrates simple usage:

```yaml
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
```

Example configuration with photo message:

```diff
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
+   photo:
+     - tests/1.png
+     - tests/2.png
```

Example configuration with document message:

```diff
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
+   document:
+     - tests/1.pdf
+     - tests/2.pdf
```

Example configuration with sticker message:

```diff
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
+   sticker:
+     - tests/3.png
+     - tests/4.png
```

Example configuration with audio message:

```diff
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
+   audio:
+     - tests/audio1.mp3
+     - tests/audio2.mp3
```

Example configuration with voice message:

```diff
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
+   voice:
+     - tests/voice1.ogg
+     - tests/voice2.ogg
```

Example configuration with location message:

```diff
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
+   location:
+     - 24.9163213,121.1424972
+     - 24.9263213,121.1224972
```

Example configuration with venue message:

```diff
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
+   venue:
+     - 24.9163213,121.1424972,title,address
+     - 24.3163213,121.1824972,title,address
```

Example configuration with video message:

```diff
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
+   video:
+     - tests/video1.mp4
+     - tests/video2.mp4
```

Example configuration with message format:

```diff
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
+   format: markdown
```

Example configuration for success and failure messages:

```diff
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
+   when:
+     status: [ success, failure ]
```

Example configuration with a custom message template:

```diff
pipeline:
  telegram:
    image: appleboy/drone-telegram
    token: xxxxxxxxxx
    to: telegram_user_id
+   message: |
+     {{ #success build.status }}
+       build {{ build.number }} succeeded. Good job.
+     {{ else }}
+       build {{ build.number }} failed. Fix me please.
+     {{ /success }}
```

# Parameter Reference

token
: telegram token from [telegram developer center](https://core.telegram.org/bots/api)

to
: telegram user id

message
: overwrite the default message template

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

# Template Reference

repo.owner
: repository owner

repo.name
: repository name

build.status
: build status type enumeration, either `success` or `failure`

build.event
: build event type enumeration, one of `push`, `pull_request`, `tag`, `deployment`

build.number
: build number

build.commit
: git sha for current commit

build.branch
: git branch for current commit

build.tag
: git tag for current commit

build.ref
: git ref for current commit

build.author
: git author for current commit

build.link
: link the the build results in drone

build.started
: unix timestamp for build started

build.finished
: unix timestamp for build finished

# Template Function Reference

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
