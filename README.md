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

### User configuration

Create a `users.yaml` file. Configuration example can be found in `cfg/users.yaml.example`

### Environment variables

Set environment variables, full description is available in `env.example`

## Running

```bash
make up
```

if you want, you can run only infrastructure in Docker containers and run bot locally:

```bash
make up-infra run
```

## Bot Commands

The bot works in direct messages only. In group chats every command except `/help`
is refused, and ordinary group messages are ignored.

All commands are available only to users with admin status:
- `/promo` - Create a new promo code
- `/get` - Show all promo codes
- `/stats` - Full statistics for specific promo codes
- `/edit` - Edit an existing promo code
- `/delete` - Delete an existing promo code

`/help` is available to everyone and explains what the bot does in DM and in groups.

## Weekly activation report

Once a week the bot can post a digest into an admins chat, covering the promo codes
created during the last `REPORT_PERIOD_WEEKS` weeks: what was created (with author,
bonus length, initial capacity, creation date and validity period), which of those
codes were activated and how many times, and which were never activated at all.

The window spans more than a single week on purpose: a promo code can be created at
any moment, so a wider window catches the activations of codes created shortly
before the previous report.

The feature is opt-in — without `ADMINS_CHAT_ID` it stays disabled and the bot logs a
warning at startup. See `.env.example` for `ADMINS_CHAT_ID`, `REPORT_CRON`,
`REPORT_PERIOD_WEEKS` and `REPORT_LANG`.

Who created a promo code and when comes from the audit log (`AUDIT_LOGS_DIR/audit.json`),
which is the only place that records it, so the report needs that directory to persist
between restarts. Activation timestamps come from the `activated_at` column of
`Promo_Code_Activations`, which is filled in by DickGrowerBot when a code is redeemed;
this bot only reads it, and mirrors the column in its own migration so that a database
created from scratch here matches the shared schema.

