package main

// взято из бота предложки чисто
func init() {
	locpool.Resources["en"] = map[string]string{
		"commands.default.message.on.command": "Unknown command",
		"callbacks.error":                     "The state was lost 😢",

		"commands.help.description":     "show help",
		"commands.help.fields.language": "Choose your language:",

		"commands.language.description":     "change the language",
		"commands.language.fields.language": "Choose your language:",

		"handlers.suggest.fields.anonymously":      "Do you want to publish a post anonymously or publicly?",
		"handlers.suggest.fields.visibleForAdmins": "Do you accept to show your name to the admins? We'd like to know who writes to us. Do you want to join us, maybe?",
		"handlers.suggest.fields.confirmation":     "Are you sure the post is ready to be sent for approval to publish?",

		"commands.promote.fields.role":       "Choose the role below:",
		"commands.promote.fields.autoAdmins": "Do you want to promote all administrators of the chat and channel?",
		"commands.promote.status.success":    "The following users has been promoted: ",
		"commands.promote.status.nobody":     "No one to promote 😕",

		"callbacks.approve.status.duplicate": "You approved this post already 😊",
		"callbacks.approve.status.revoked":   "Unfortunately, the post was revoked 😞",
		"callbacks.approve.status.published": "The post was published after it had been approved by the following people",
		"callbacks.approve.status.no.author": "You have to be an author or admin to approve the post",

		"callbacks.revoke.status.published": "I'm sorry but revocation is not possible since the post has been already published",

		"visibility.public": "Publicly",
		"visibility.anon":   "Anonymously",

		"messages.approve":    "Is this message appropriate for publication?",
		"messages.banned":     "Unfortunately, you were banned by the administrator for some reason, and you are not allowed to use this bot.",
		"messages.admin.only": "This command is intended for use by the admin only!",
		"messages.revoke":     "If you change your mind, you can revoke your suggestion at any time.",
		"messages.refused":    "Okay! Send me a message again, and I'll help you to fill all required parameters out.",

		"success": "👍🏼👌🏼",
		"failure": "Something went wrong...",
	}

	locpool.Resources["ru"] = map[string]string{
		"commands.default.message.on.command": "Неизвестная команда",
		"callbacks.error":                     "Состояние формы потерялось 😢",

		"promo.create": "Промокод создан:",

		"commands.help.description":     "показать помощь",
		"commands.help.fields.language": "Выберите свой язык:",

		"commands.language.description":     "сменить язык",
		"commands.language.fields.language": "Выберите свой язык:",
		"commands.promo.fields.promo":       "Введите промокод новым сообщением",

		"handlers.suggest.fields.anonymously":      "Вы хотите опубликовать сообщение анонимно или публично?",
		"handlers.suggest.fields.visibleForAdmins": "Вы согласны показать своё имя админам? Нам бы хотелось знать, кто нам пишет, чтобы отбирать кандидатов в новые админы.",
		"handlers.suggest.fields.confirmation":     "Вы уверены, что хотите отправить пост на одобрение к публикации?",

		"commands.promote.fields.role":       "Выберите роль из списка ниже:",
		"commands.promote.fields.autoAdmins": "Вы хотите изменить роль для всех администраторов чата и канала?",
		"commands.promote.status.success":    "Следующие пользователи были назначены: ",
		"commands.promote.status.nobody":     "Некого назначать 😕",

		"callbacks.approve.status.duplicate": "Вы уже одобряли этот пост 😊",
		"callbacks.approve.status.revoked":   "К сожалению, пост был отозван 😞",
		"callbacks.approve.status.published": "Пост был опубликован после одобрения следующими людьми",
		"callbacks.approve.status.no.author": "Одобрять посты к публикации могут только авторы или администраторы",

		"callbacks.revoke.status.published": "Сожалею, но отозвать данный пост невозможно, так как он уже был опубликован",

		"visibility.public": "Публично",
		"visibility.anon":   "Анонимно",

		"messages.approve":    "Данное сообщение пригодно для публикации?",
		"messages.banned":     "К сожалению, по какой-то причине Вы были заблокированы администратором и не сможете воспользоваться ботом.",
		"messages.admin.only": "Данной командой может пользоваться только администратор!",
		"messages.revoke":     "Если передумаете, можете отозвать пост в любое время.",
		"messages.refused":    "Окей! Отправь мне сообщение снова, и я проведу тебя через установку необходимых параметров!",

		"failure": "Что-то пошло не так...",

		"Approve":  "Одобрить",
		"Approved": "Одобрено",
		"Revoke":   "Отозвать",
		"Ban":      "Заблокировать",
		"Unban":    "Разблокировать",

		"user":   "обычный",
		"author": "автор",
		"admin":  "админ",
	}
}
