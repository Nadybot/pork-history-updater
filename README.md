# pork-history-updater

A Go CLI that keeps a historical record of Anarchy Online characters by polling the official people API and persisting changes to a MySQL database.

## What it does

- Streams all rows from a `player` table.
- For each player, fetches the current character info from `https://people.anarchy-online.com`.
- Compares the remote state with the local state.
- If something changed, it inserts the new state into `player_history` as a snapshot and updates the `player` row.
- If the character no longer exists remotely, it marks the player as deleted.
- Supports concurrent workers and connection-error retries.

## Architecture

The project uses a ports-and-adapters (aka hexagonal architecture) layout:

- `cmd/updater` — entry point and dependency wiring.
- `internal/domain` — `Player` and `GuildMembership` entities with equality and diff logic.
- `internal/application` — `Updater` orchestration and repository/fetcher interfaces.
- `internal/adapters/external/pork` — HTTP client for the Anarchy Online people API.
- `internal/adapters/persistence/mysql` — MySQL repository implementation with paging, retries, and DTO mapping.
- `internal/adapters/persistence/dryrun` — decorator that logs writes instead of persisting them.
- `pkg/config` — CLI flag and environment variable parsing.
- `platform/db` — database connection setup.

## Requirements

- Go 1.26+
- MySQL (or MariaDB) with a `player` table and a `player_history` table.

## Configuration

Configuration can be passed as CLI flags or environment variables. Flags take precedence.

| Flag           | Environment variable | Default     | Description                                        |
|----------------|----------------------|-------------|----------------------------------------------------|
| `--dry-run`    | `DRY_RUN`            | `false`     | Log writes but do not persist them.                |
| `--max-workers`| `MAX_WORKERS`        | `5`         | Number of simultaneous API requests.               |
| `--db-host`    | `DB_HOST`            | `localhost` | Database host.                                     |
| `--db-port`    | `DB_PORT`            | `3306`      | Database port.                                     |
| `--db-user`    | `DB_USER`            | `porkhist`  | Database user.                                     |
| `--db-pass`    | `DB_PASSWORD`        | empty       | Database password.                                 |
| `--db-name`    | `DB_NAME`            | `porkhist`  | Database name.                                     |
| `--db-type`    | `DB_TYPE`            | `mysql`     | Database driver type.                              |

## Database expectations

The MySQL repository expects at least these tables and columns:

**player**

- `nickname`, `char_id`, `first_name`, `last_name`, `guild_rank`, `guild_rank_name`
- `level`, `faction`, `profession`, `profession_title`, `gender`, `breed`
- `defender_rank`, `defender_rank_name`, `guild_id`, `guild_name`, `server`
- `last_checked`, `last_changed`, `deleted`

**player_history**

Same columns as `player`, used to store snapshots before an update or deletion.

## Usage

Build:

```bash
go build -o updater ./cmd/updater
```

Run with flags:

```bash
./updater --db-pass=secret --max-workers=10
```

Run with environment variables:

```bash
DB_PASSWORD=secret MAX_WORKERS=10 ./updater
```

Dry run:

```bash
./updater --dry-run
```

## Logging

The application logs structured JSON to stdout using the standard `log/slog` package.
