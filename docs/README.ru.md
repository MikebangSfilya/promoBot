# PromoBot

[![CI Build](https://github.com/MikebangSfilya/promoBot/actions/workflows/ci-build.yml/badge.svg)](https://github.com/MikebangSfilya/promoBot/actions/workflows/ci-build.yml)

Telegram-бот для управления промокодами, построенный на базе [goSadTgBot](https://github.com/kozalosev/goSadTgBot). Работает на версии Go 1.24

Русская версия | [English](../README.md)

## Описание

PromoBot позволяет администраторам создавать промокоды для [DickGrowerBot](https://github.com/kozalosev/DickGrowerBot).

## Возможности

- Создание промокодов с настраиваемыми параметрами
- Просмотр таблицы всех промокодов
- Контроль доступа на основе конфигурации пользователей

## Установка

### Конфигурация пользователей

Создайте файл `users.yaml`. Пример создания конфигурации находится в `cfg/users.yaml.example`

### Переменные окружения

Задайте переменные окружения, полное описание находится в `env.example`

## Запуск

```bash
make up
```

Если хотите запустить только инфраструктуру в Docker контейнерах и запустить бота локально:

```bash
make up-infra
make run
```

## Команды бота

Все команды доступны только пользователям со статусом admin:
- `/promo` - Создать новый промокод
- `/get` - Показать все промокоды