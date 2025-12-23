package main

func init() {
	locpool.Resources["en"] = map[string]string{
		"commands.default.message.on.command": "Unknown command",
		"callbacks.error":                     "The state was lost 😢",
		"commands.default.message":            "It appears you haven't invoked any of the bot's support commands. Call /promo for create promocode.",

		"commands.promo.description": "Generate promoCode for dickGrowerBot",

		"commands.help.description":     "show help",
		"commands.help.fields.language": "Choose your language:",

		"commands.language.description":     "change the language, реально",
		"commands.language.fields.language": "Choose your language:",
		"commands.promo.fields.promo":       "Enter the promo code in a new message",
		"commands.promo.fields.length":      "Enter the bonus length for the promo code",
		"commands.promo.fields.capacity":    "Enter the number of activations for the promo code",

		"fieldPromoCreated": "Promo code created: ",
		"actionCreate":      "Create",
		"actionCancel":      "Cancel",
		"textToCreate":      "Confirm promo code creation:",
		"promoCanceled":     "Promo code creation canceled",
		"errToCreatePromo":  "Error creating promo code",
		"errNoPermission":   "Access denied. Admin rights required",

		"listPromoCodesTitle":       "Promo codes",
		"listPromoCodesTotalEnding": "Total: %d promo codes",

		"success": "👍🏼👌🏼",
		"failure": "Something went wrong...",
	}

	locpool.Resources["ru"] = map[string]string{
		"commands.default.message.on.command": "Неизвестная команда",
		"callbacks.error":                     "Состояние формы потерялось 😢",
		"commands.default.message":            "Кажется вы не вызвали ни одной команды которую поддерживает бот. Вызовите /promo для создания промокода.",

		"commands.promo.description": "Создать промокод для dickGrowerBot",

		"commands.help.description":     "показать помощь",
		"commands.help.fields.language": "Выберите свой язык:",

		"commands.language.description": "сменить язык пользования",

		"commands.language.fields.language": "Выберите свой язык:",
		"commands.promo.fields.promo":       "Введите промокод новым сообщением",
		"commands.promo.fields.length":      "Введите длину писи для создания промокода",
		"commands.promo.fields.capacity":    "Введите кол-во активаций промокода",

		"fieldPromoCreated": "Промокод создан: ",
		"actionCreate":      "Создать",
		"actionCancel":      "Отменить",
		"textToCreate":      "Подтвердите создание промокода:",
		"promoCanceled":     "Создание промокода отменено",
		"errToCreatePromo":  "Ошибка при создание промокода",
		"errNoPermission":   "Доступ запрещен. Требуются права администратора",

		"noPromo": "Нет промокодов в базе",

		"listPromoCodesTitle":       "Промокоды",
		"listPromoCodesTotalEnding": "Всего: %d промокодов",

		"success": "👍🏼",
		"failure": "Что-то пошло не так...",

		"user":   "обычный",
		"author": "автор",
		"admin":  "админ",
	}
}
