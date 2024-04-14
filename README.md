# Overview

Telegram bot that downloads photo or video from Twitter and sends it to the user. Supports multiple media. Sample https://t.me/my_twitter_downloader_bot

Built with awesome https://github.com/gotd/td

## Usage

```bash
export APP_ID=111 APP_HASH=abcdef BOT_TOKEN=12345:abcdef
go run main.go bot start -d /data_folder -s -d /data_folder/session.json

  -a, --admin-id int             Restrict to a user
  -D, --debug-telegram           debug log
  -d, --download-folder string   download folder
  -f, --forward-to int           forward media that was sent to a user to a channel
  -B, --include-bot-name         include bot name
  -T, --include-text             include text
  -U, --include-url              include url
  -s, --session-file string      session file (default "twitter-downloader-session.json")
  -l, --use-limiter              use rate limiter (default true)

```