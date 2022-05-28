<p align="center">
  <img alt="taterubot logo" src="assets/art.svg" height="300" />
  <h3 align="center">Taterubot</h3>
  <p align="center">Record audio messages and share them with your friends without leaving Discord! Whatsapp style</p>
</p>

---

`golangci-lint` is a fast Go linters runner. It runs linters in parallel, uses caching, supports `yaml` config, has integrations
with all major IDE and has dozens of linters included.

## Install `golangci-lint`

- [On my machine](https://golangci-lint.run/usage/install/#local-installation);
- [On CI/CD systems](https://golangci-lint.run/usage/install/#ci-installation).


## Badges
![Language](https://img.shields.io/github/languages/top/hectorgabucio/taterubot-dc?style=for-the-badge)
![Build status](https://img.shields.io/github/workflow/status/hectorgabucio/taterubot-dc/Continuous%20integration?logo=github&style=for-the-badge)
[![License](https://img.shields.io/badge/license-MIT-green?logo=readthedocs&style=for-the-badge)](./LICENSE.md)
## Known bugs and limitations
- Sometimes race condition if you try to record a very short audio.
- Cant really scale horizontally; There is an internal state using channels to manage the recording, cant handle the start and end of recording in different instances.
- Not meant for unstable connections: if you are outside with the phone and trying to record an audio and you have low signal, you most likely will lose that audio.