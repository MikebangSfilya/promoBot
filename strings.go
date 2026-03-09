package main

func init() {
	locpool.Resources["en"] = map[string]string{
		"commands.default.message.on.command": "Unknown command",
		"callbacks.error":                     "The state was lost 😢",
		"commands.default.message":            "It appears you haven't invoked any of the bot's support commands. Call /promo to create a promo code.",

		"commands.promo.description": "Generate promo code for dickGrowerBot",
		"commands.get.description":   "Get promo codes table",
		"commands.stats.description": "Get full promo codes statistics",

		"commands.help.description":     "Show help",
		"commands.help.fields.language": "Choose your language:",

		"commands.language.description":     "Change language",
		"commands.language.fields.language": "Choose your language:",
		"commands.promo.fields.promo":       "Enter the promo code in a new message",
		"commands.promo.fields.length":      "Enter the length increase for the promo code",
		"commands.promo.fields.capacity":    "Enter the number of activations for the promo code",
		"commands.promo.fields.setDates":    "Would you like to set custom start and end dates?",
		"commands.promo.fields.since":       "Enter the start date (number of days, DD.MM.YYYY, or YYYY-MM-DD):",
		"commands.promo.fields.until":       "Enter the end date (number of days, DD.MM.YYYY, or YYYY-MM-DD):",
		"commands.promo.fields.code":        "Enter promo code(s) you want to view in a new message on one line, separated by space or comma",

		"messages.promo.created_full": "Promo code successfully created! Promo code: %s. Length bonus: %s. Activations: %s",

		"fieldPromoCreated":   "Promo code created: ",
		"actionCreate":        "Create",
		"actionCancel":        "Cancel",
		"actionSetBothDates":  "Set both dates",
		"actionEndlessPromo":  "Endless promo",
		"actionFromNowToDate": "Temporary promo",
		"textToCreate":        "Confirm promo code creation:",
		"promoCanceled":       "Promo code creation canceled",
		"errToCreatePromo":    "Error creating promo code",
		"errNoPermission":     "Access denied. Admin rights required",

		"noPromo": "No promo codes in the database or not found for the specific query",

		"listPromoCodesTitle":       "Promo codes",
		"listPromoCodesTotalEnding": "Total: %d promo codes",

		"BadLength":   "Invalid length value",
		"BadCapacity": "Invalid capacity value",
		"BadSince":    "Invalid start date. Use a number of days, DD.MM.YYYY, or YYYY-MM-DD",
		"BadUntil":    "Invalid end date. Use a number of days, DD.MM.YYYY, or YYYY-MM-DD",

		"success": "👍🏼",
		"failure": "Something went wrong...",

		"invalid": "Invalid arguments",

		"user":   "user",
		"author": "author",
		"admin":  "admin",
	}

	locpool.Resources["ru"] = map[string]string{
		"commands.default.message.on.command": "Неизвестная команда",
		"callbacks.error":                     "Состояние формы потерялось 😢",
		"commands.default.message":            "Кажется, вы не вызвали ни одной команды, которую поддерживает бот. Вызовите /promo для создания промокода.",

		"commands.promo.description": "Создать промокод для dickGrowerBot",
		"commands.get.description":   "Получить таблицу промокодов",
		"commands.stats.description": "Получить полную статистику по промокодам",

		"commands.help.description":     "Показать помощь",
		"commands.help.fields.language": "Выберите свой язык:",

		"commands.language.description": "Сменить язык пользования",

		"commands.language.fields.language": "Выберите свой язык:",
		"commands.promo.fields.promo":       "Введите промокод новым сообщением",
		"commands.promo.fields.length":      "Введите прибавку к длине для создания промокода",
		"commands.promo.fields.capacity":    "Введите количество активаций промокода",
		"commands.promo.fields.setDates":    "Хотите задать даты начала и окончания?",
		"commands.promo.fields.since":       "Введите дату начала (количество дней, ДД.ММ.ГГГГ или ГГГГ-ММ-ДД):",
		"commands.promo.fields.until":       "Введите дату окончания (количество дней, ДД.ММ.ГГГГ или ГГГГ-ММ-ДД):",
		"commands.promo.fields.code":        "Введите промокод(ы), которые хотите просмотреть, новым сообщением в одну строчку через пробел или запятую",

		"fieldPromoCreated":   "Промокод создан: ",
		"actionCreate":        "Создать",
		"actionCancel":        "Отменить",
		"actionSetBothDates":  "Задать обе даты",
		"actionEndlessPromo":  "Бессрочный",
		"actionFromNowToDate": "Временный",
		"textToCreate":        "Подтвердите создание промокода:",
		"promoCanceled":       "Создание промокода отменено",
		"errToCreatePromo":    "Ошибка при создании промокода",
		"errNoPermission":     "Доступ запрещен. Требуются права администратора",

		"messages.promo.created_full": "Промокод успешно создан! Промокод: %s. Прибавка к длине: %s. Активаций: %s",

		"noPromo": "Нет промокодов в базе или не найдены по конкретному запросу",

		"listPromoCodesTitle":       "Промокоды",
		"listPromoCodesTotalEnding": "Всего: %d промокодов",

		"success": "👍🏼",
		"failure": "Что-то пошло не так...",

		"invalid": "Неверные аргументы",

		"BadLength":   "Неверное значение для поля \"прибавка к длине\"",
		"BadCapacity": "Неверное значение для поля \"количество активаций\"",
		"BadSince":    "Неверная дата начала. Используйте количество дней, ДД.ММ.ГГГГ или ГГГГ-ММ-ДД",
		"BadUntil":    "Неверная дата окончания. Используйте количество дней, ДД.ММ.ГГГГ или ГГГГ-ММ-ДД",

		"user":   "обычный",
		"author": "автор",
		"admin":  "админ",
	}
}
