# TRCMUSIC

A high-performance Telegram bot that streams music into group voice chats, written in Go. 

## Features
- Fast and performant (written in Go)
- Supports voice chat streaming via Pyrogram-like architecture (gogram)
- Powered by `yt-dlp` and `ffmpeg`

## Deployment

Deploying is easy! You can deploy directly to Render using the button below. 

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/bindplant123/TRCMUSIC)

### Environment Variables
Before deploying, make sure you have the following variables ready:
- `API_ID`: Your Telegram API ID from my.telegram.org
- `API_HASH`: Your Telegram API Hash from my.telegram.org
- `TOKEN`: Your Bot Token from @BotFather
- `MONGO_DB_URI`: MongoDB connection string
- `STRING_SESSION`: A Pyrogram string session for the assistant account
- `OWNER_ID`: Your Telegram User ID
