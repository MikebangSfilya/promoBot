package main

func init() {
	locpool.Resources["en"] = map[string]string{
		"commands.default.message.on.command": "Unknown command",
		"callbacks.error":                     "The state was lost 😢",
		"commands.default.message":            "It appears you haven't invoked any of the bot's support commands. Call /help for assistance.",

		"commands.promo.fields.promo": "Vvdedite promic, pendosLanguage",
		"commands.promo.description":  "Generate promoCode for dickGrowerBot",

		"commands.help.description":     "show help",
		"commands.help.fields.language": "Choose your language:",

		"commands.language.description":     "change the language, реально",
		"commands.language.fields.language": "Choose your language:",

		"success": "👍🏼👌🏼",
		"failure": "Something went wrong...",
	}

	locpool.Resources["ru"] = map[string]string{
		"commands.default.message.on.command": "Неизвестная команда",
		"callbacks.error":                     "Состояние формы потерялось 😢",

		"commands.default.message": "Кажется вы не вызвали ни одной команды которую поддерживает бот. Вызовите /help для помощи.",

		"commands.help.description":     "показать помощь",
		"commands.help.fields.language": "Выберите свой язык:",

		"commands.language.description": "сменить язык пользования",
		"commands.promo.description":    "Generate promoCode for dickGrowerBot",

		"commands.language.fields.language": "Выберите свой язык:",
		"commands.promo.fields.promo":       "Введите промокод новым сообщением",

		"failure": "Что-то пошло не так...",

		"user":   "обычный",
		"author": "автор",
		"admin":  "админ",
	}
}
