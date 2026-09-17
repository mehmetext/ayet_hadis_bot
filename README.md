# Ayet & Hadith Bot

An open-source Go and SQLite bot that fetches verses and hadiths from QuranEnc and HadeethEnc and sends them to Telegram subscribers during configured hours.

Telegram bot: [@hadis_ayet_bot](https://t.me/hadis_ayet_bot)

## Features

- Uses Telegram polling; no webhook, public HTTPS endpoint, or reverse proxy is required.
- Users subscribe with `/start` and choose their language.
- Supports Arabic, English, Turkish, and German.
- Alternates between verse and hadith deliveries.
- Subscribers using the same language share one global content stream.
- Sends evenly distributed notifications inside the configured daily window.
- Does not cache content locally; SQLite stores only delivery state, duplicate history, and user subscriptions.
- Retries temporary QuranEnc, HadeethEnc, and Telegram failures.
- Builds and deploys automatically through GitHub Actions and GitHub Container Registry (GHCR).

## Telegram commands

```text
/start     Subscribe and choose a notification language
/stop      Stop notifications
/language  Change the notification language
/status    Show language and subscription status
/help      Show the command list
```

After `/start`, the bot explains its schedule and verse/hadith rotation, then shows the language buttons. Subscription does not send an immediate content message; the user joins the next planned delivery.

## How it works

At each scheduled slot, the bot:

1. Finds languages with active Telegram subscribers in SQLite.
2. Reads the next content type (`verse` or `hadith`) for each language.
3. Fetches one item from QuranEnc or HadeethEnc.
4. Sends it to all active subscribers of that language.
5. Atomically records the successful delivery and advances the rotation.

Verse and hadith text is never written to disk. Messages longer than Telegram’s 4096-character limit are split into consecutive parts without truncating content.

## Requirements

- Go 1.26 or newer
- Docker and Docker Compose
- A Telegram BotFather token

## Run locally

Create the local configuration file:

```sh
cp .env.example .env
```

Fill in the Telegram token and other values in `.env`. To fetch and print one item without sending it to Telegram:

```sh
go run ./cmd/ayet-hadis-bot once
```

To run the scheduler and Telegram polling together:

```sh
go run ./cmd/ayet-hadis-bot run
```

## Run with Docker

```sh
cp .env.example .env
# Edit .env
docker compose up -d --build
docker compose logs -f
```

SQLite is stored in the persistent `bot-data` Docker volume. Users, rotation state, and duplicate history survive container restarts.

## Configuration

`.env` is the single source of runtime configuration:

```env
TIMEZONE=Europe/Istanbul
CONSOLE_LANGUAGE=ara
SEND_WINDOW_START=06:30
SEND_WINDOW_END=22:30
DAILY_NOTIFICATION_COUNT=4
HTTP_TIMEOUT_SECONDS=30
DATA_DIR=data
TELEGRAM_BOT_TOKEN=
```

With `06:30–22:30` and `DAILY_NOTIFICATION_COUNT=4`, the inclusive slots are `06:30`, `11:50`, `17:10`, and `22:30`.

Never commit `.env` or a Telegram token. `.env` is excluded by both `.gitignore` and `.dockerignore`.

## CI/CD and server deployment

Every commit to `main` is tested by the [GitHub Actions workflow](.github/workflows/deploy.yml), built as a Docker image, pushed to GHCR, and deployed to the server over SSH.

The server does not clone the repository or build Docker images. The pipeline copies the current `docker-compose.yml`, generates `.env` from GitHub configuration, logs in to GHCR, pulls the image, and runs `docker compose up -d`.

### GitHub Actions secrets

```text
DEPLOY_HOST
DEPLOY_USER
DEPLOY_PATH
DEPLOY_SSH_KEY_B64
GHCR_USERNAME
GHCR_TOKEN
TELEGRAM_BOT_TOKEN
```

### GitHub Actions variables

```text
TIMEZONE
CONSOLE_LANGUAGE
SEND_WINDOW_START
SEND_WINDOW_END
DAILY_NOTIFICATION_COUNT
HTTP_TIMEOUT_SECONDS
DATA_DIR
```

`DEPLOY_SSH_KEY_B64` must be a Base64-encoded private SSH key to avoid multiline secret formatting issues. `GHCR_TOKEN` must be a classic PAT with at least `read:packages` permission for pulling the private image.

## Sources

- [QuranEnc API](https://quranenc.com/nqo/home/api)
- [HadeethEnc](https://hadeethenc.com)

## Contributing

Contributions are welcome:

1. Fork the repository.
2. Create a feature branch:

   ```sh
   git checkout -b feature/your-change
   ```

3. Make your changes and run the tests:

   ```sh
   go test ./...
   ```

4. Commit and push your branch to your fork.
5. Open a Pull Request against `main`.

Please describe the change, test results, and any behavior changes in the Pull Request. New features should preserve the API-based content flow, SQLite duplicate prevention, Telegram commands, and Docker deployment model.

## License

No license has been specified for this repository yet. Check with the repository owner before redistributing the project or submitting substantial contributions.
