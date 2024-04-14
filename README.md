# Overview

Telegram bot that downloads photo or video from Twitter and sends it to the user. Supports multiple media. Sample https://t.me/my_twitter_downloader_bot

Built with awesome https://github.com/gotd/td

## Usage

```bash
export APP_ID=111 APP_HASH=abcdef BOT_TOKEN=12345:abcdef
go run main.go bot start -d /data_folder -s /data_folder/session.json

  -a, --admin-id int             admin id (optional)
  -r, --restrict-to-admin-id     restrict usage to admin id
  -D, --debug-telegram           enable debug log
  -d, --download-folder string   download folder
  -f, --forward-to int           forward media that was sent to a user to a channel (optional)
  -B, --include-bot-name         post will include bot name 
  -T, --include-text             post will include text
  -U, --include-url              post will include tweet url
  -p, --limit-pending int        limit pending requests from a user (admin has no limit) (default 1)
  -L, --limit-per-day int        limit requests per day (admin has no limit). Will reset after restart or next day. (default 30)
  -s, --session-file string      session file (default "twitter-downloader-session.json")
  -l, --use-limiter              use rate limiter for telegram api calls (default true)

```
