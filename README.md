# Les Copaings Bot

Bot for the private server Discord "Les Copaings"

## Features

- Levels & XP
- Roles management
- Purge command

### CopaingXPs

Functions:
- $xp-message(x;y) = 0.025 x^{1.25}\sqrt{y}+1$ where $x$ is the length of the message and $y$ is the diversity of the 
message (number of different rune)
- $xp-vocal(x)=0.01 x^{1.3}+1$ where $x$ is the time spent in vocal (in second)
- $level(x)=0.2 \sqrt{x}$ where $x$ is the xp
- $level^{-1}(x)=25x^2$ where $x$ is the level

## Installation

There are two ways to install the bot: docker and build.

### Docker

1. Clone the repository
```bash
$ git clone https://github.com/anhgelus/les-copaings-bot.git
```
2. Go into the repository, rename `.env.example` into `.env` and customize it: add your token, change the user and the 
password of the database
3. Start the compose file
```bash
$ docker compose up -d --build
```

Now you have to edit `config/config.toml`.
You can understand how this config file works below.
After editing this file, you have to start again the bot.
Every time you edit this file, you must restart the bot.

You can stop the compose file with `docker compose down`

### Build

1. Clone the repository
```bash
$ git clone https://github.com/anhgelus/les-copaings-bot.git
```
2. Install Go 1.22+
3. Go into the repository and build the program
```bash
$ go build . 
```
4. Run the application through bash (or PowerShell if you are on Windows)

Now you have to edit `config/config.toml`.
You can understand how this config file works below.
After editing this file, you have to start again the bot.
Every time you edit this file, you must restart the bot.

## Config

The main config file is `config/config.toml`.

The default configuration is
```toml
debug = false
author = "anhgelus"

[redis]
address = "localhost:6379"
password = ""
db = 0

[database]
host = "localhost"
user = ""
password = ""
db_name = ""
port = 5432
```

- `debug` is true if the bot is in debug mode (don't turn it on unless you are modifying the source code)
- `author` is the author's name
- `[redis].address` is the address of redis (using docker, it's `redis:6379`)
- `[redis].password` is the redis's password
- `[redis].db` is the db to use
- `[database].host` is the host of postgres (using docker, it's `postgres`)
- `[database].user` is the user of postgres to use (using docker, it must be the same value as `POSTGRES_USER` in `.env`)
- `[database].password` is the user's password of postgres to use (using docker, it must be the same value as
`POSTGRES_PASSWORD` in `.env`)` 
- `[database].db_name` is the postgres's database name to use (using docker, it must be the same value as `POSTGRES_DB`
in `.env`)
- `[database].port` is the port of postgres to use (using docker, it's `5432`)

## Technologies

- Go 1.24
- anhgelus/gokord
