# PromoBot

[![CI Build](https://github.com/MikebangSfilya/promoBot/actions/workflows/ci-build.yml/badge.svg)](https://github.com/MikebangSfilya/promoBot/actions/workflows/ci-build.yml)

Telegram bot for managing promo codes, built on top of [goSadTgBot](https://github.com/kozalosev/goSadTgBot). Works with Go 1.24

[Русская версия](docs/README.ru.md) | English

## Description

PromoBot allows administrators to create promo codes for [DickGrowerBot](https://github.com/kozalosev/DickGrowerBot).

## Features

- Creating promo codes with customizable parameters
- Viewing table of all promo codes
- Access control based on user configuration

## Installation

### Clone the repository
```bash
git clone https://github.com/MikebangSfilya/promoBot
```

### User configuration

Create a `users.yaml` file. Configuration example can be found in `cfg/users.yaml.example`

### Environment variables

Set environment variables, full description is available in `env.example`

## Running

```bash
make compose-build
make up
```

## Bot Commands

All commands are available only to users with admin status:
- `/promo` - Create a new promo code
- `/get` - Show all promo codes
