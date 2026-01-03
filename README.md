# CatBot-Go

CatBot is a lightweight IRC bot built in Go that brings a virtual cat named **Purrito** to your IRC channels. The bot tracks user interactions through a love meter system (0-100%) and provides randomized responses based on your relationship with the cat.

## Features

- Love meter per user (0-100%) that persists across sessions
- Randomized cat reactions and emotes
- Timed appearances — Purrito appears randomly and stays for 10 minutes
- Daily decay system for maintaining bonds
- Multi-channel support with separate love meters per channel
- PostgreSQL for persistent storage

## Commands

| Command | Description |
|---------|-------------|
| `!pet purrito` | Pet the cat (60% accept/40% reject, ±1 love) |
| `!love purrito` | Show affection (same as pet) |
| `!feed purrito` | Feed the cat (varies by food type) |
| `!slap purrito` | Slap the cat (warning first, then -1 love) |
| `!catnip purrito` | Give catnip (+3 love if accepted, once per day) |
| `!laser purrito` | Play with laser pointer |
| `!status purrito` | Check your current love meter and mood |
| `!toplove` | Show top 5 players by love meter |
| `!purrito` | Display help/info about the bot |
| `!invite purrito` | Invite bot to join a new channel |

## Love Meter Moods

| Love % | Mood |
|--------|------|
| 0% | Hostile |
| 1-19% | Sad |
| 20-49% | Cautious |
| 50-79% | Friendly |
| 80-99% | Loves you |
| 100% | Perfect bond |

## Command Probabilities

| Command | Accept Chance | Reject Chance | Love Change |
|---------|---------------|---------------|-------------|
| `!pet` / `!love` | 60% | 40% | +1 (accept) / -1 (reject) |
| `!feed` | 60% | 40% | +1 (accept) / -1 (reject) |
| `!laser` | 60% | 40% | +1 (accept) / -1 (reject) |
| `!catnip` | 70% | 30% | +3 (accept) / -1 (reject) |
| `!slap` | - | - | Warning first, then -1 |

Note: `!catnip` can only be used once per day per user.

## Timings

| Event | Duration |
|-------|----------|
| Purrito appearance interval | Every 30 minutes |
| Purrito presence duration | 10 minutes |
| Love decay check | Every 24 hours |
| Love decay amount | -5 (only at 100% love if no interaction) |

### Presence System

- Purrito randomly appears in the channel every 30 minutes
- Once present, Purrito stays for 10 minutes waiting for interaction
- Commands like `!pet`, `!feed`, and `!laser` require Purrito to be present
- After a user interacts, Purrito's presence is consumed (disappears)
- If no one interacts within 10 minutes, Purrito leaves with a farewell message

### Daily Decay

- Players at 100% love (perfect bond) will lose 5 love points if they don't interact within 24 hours
- A warning message is sent on the first decay
- This encourages regular interaction to maintain the bond

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Or: Go 1.23+ and PostgreSQL 15

### Running with Docker

1. Copy the example environment file and configure it:

```bash
cp .env.example .env
# Edit .env with your IRC and database settings
```

2. Start the services:

```bash
docker compose up --build
```

### Environment Variables

**Database:**
- `POSTGRES_USER` - PostgreSQL username
- `POSTGRES_PASSWORD` - PostgreSQL password
- `POSTGRES_DB` - Database name

**IRC:**
- `IRC_HOST` - IRC server hostname
- `IRC_PORT` - IRC server port
- `IRC_SSL` - Enable SSL (true/false)
- `IRC_NICK` - Bot nickname
- `IRC_USER` - Bot ident/user
- `IRC_CHANNELS` - Comma-separated channels (e.g., `#channel1,#channel2`)
- `IRC_NETWORK` - Network name
- `IRC_NICKSERV_PASSWORD` - NickServ password (optional)
- `IRC_PASSWORD` - IRC server password (optional)

### Running Locally

1. Start PostgreSQL (or use the docker-compose db service)

2. Set environment variables or create a `config/config.dev.json`

3. Run migrations:

```bash
go run ./cmd/main.go migrate up
```

4. Start the bot:

```bash
go run ./cmd/main.go serve
```

## Project Structure

```
catbot-go/
├── cmd/main.go                 # CLI entry point
├── config/                     # Configuration loading
├── internal/
│   ├── bot/                    # IRC client setup and event handlers
│   ├── db/                     # Database connection and repositories
│   ├── commands/               # CLI commands (serve, migrate)
│   └── services/
│       ├── catbot/             # Game loop and presence logic
│       ├── cat_actions/        # Action execution and responses
│       ├── lovemeter/          # Love meter calculations
│       └── commands/           # IRC command handlers
├── db/migrations/              # SQL migrations
├── docker-compose.yaml
└── Dockerfile
```

## Tech Stack

- **Go 1.23**
- **goirc** - IRC client library
- **GORM** - ORM for PostgreSQL
- **Cobra** - CLI framework
- **golang-migrate** - Database migrations
