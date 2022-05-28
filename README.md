<p align="center">
  <img alt="taterubot logo" src="assets/art.svg" height="300" />
  <h1 align="center">Taterubot</h3>
  <p align="center">Record audio messages and share them with your friends without leaving Discord! Whatsapp style</p>
</p>

---

`Taterubot` is a Discord bot that allows recording voice messages and send them on your channel. 

## Requirements
- Discord bot: create yours [here](https://discord.com/developers/applications).
- [ffmepg](https://ffmpeg.org/) installed in the host machine. It is needed to convert and manipulate audio files.
- Docker and docker-compose (if running locally)

## Run `Taterubot` on your own machine (minimal setup)

- Rename .env.example to .env
- Put your discord token on BOT_TOKEN env variable.
- (Optionally) change LANGUAGE to **:gb:** or **:es:**
- Run *go mod download*
- Run *make local-infra*
- Run *go run main.go*

## Badges
![Language](https://img.shields.io/github/languages/top/hectorgabucio/taterubot-dc?style=for-the-badge)
![Build status](https://img.shields.io/github/workflow/status/hectorgabucio/taterubot-dc/Continuous%20integration?logo=github&style=for-the-badge)
[![License](https://img.shields.io/badge/license-MIT-green?logo=readthedocs&style=for-the-badge)](./LICENSE.md)
## Known bugs and limitations
- Sometimes race condition if you try to record a very short audio.
- Cant really scale horizontally; There is an internal state using channels to manage the recording, cant handle the start and end of recording in different instances.
- Not meant for unstable connections: if you are outside with the phone and trying to record an audio and you have low signal, you most likely will lose that audio.

## Thanks to
WIP