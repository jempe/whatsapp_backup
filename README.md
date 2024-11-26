# WhatsApp Backup

This repository provides a backup solution for WhatsApp messages using a Go-based API server. The application is designed to handle various configurations, rate limiting, and email notifications.

## Features

- API server for managing WhatsApp backups
- Configurable database connections
- Rate limiting to control API request rates
- SMTP settings for email notifications
- Embedding settings for text processing
- Cron job support for scheduled tasks

## Installation

To install and run the API server, follow these steps:

1. Clone the repository:
```sh
git clone https://github.com/jempe/whatsapp_backup.git
cd whatsapp_backup
```

2. Configure the environment variables (e.g., database connection, SMTP settings).
   
3. Build and run the application:

```sh
go build -o whatsapp_backup cmd/api/main.go
./whatsapp_backup
```

## Configuration

The application can be configured using command-line flags. The following are the available flags:

- `--env`:  Environment (development|staging|production)
- `--port`:  API server port
- `--db-dsn`:  PostgreSQL DSN
- `--db-max-open-conns`:  PostgreSQL max open connections
- `--db-max-idle-conns`:  PostgreSQL max idle connections
- `--db-max-idle-time`:  PostgreSQL max connection idle time
- `--limiter-enabled`:  Enable rate limiter
- `--limiter-rps`:  Rate limiter maximum requests per second
- `--limiter-burst`:  Rate limiter maximum burst
- `--smtp-host`:  SMTP host
- `--smtp-port`:  SMTP port
- `--smtp-username`:  SMTP username
- `--smtp-password`:  SMTP password
- `--smtp-sender`:  SMTP sender
- `--embeddings-per-batch`:  Embeddings per batch
- `--max-tokens`:  Max tokens per document
- `--sentence-transformers-server-url`:  Sentence Transformers Server URL
- `--default-embeddings-provider`:  Default embeddings provider (google|openai|sentence-transformers)
- `--do-cron-job`:  Run cron job
- `--site-base-url`:  Base URL for the site
- `--version`:  Display version and exit

## Usage

To start the API server, use the following command:
```sh
./whatsapp_backup --env=production --port=8001
```

## License

This project is licensed under the Apache 2.0 License. See the LICENSE file for details.
