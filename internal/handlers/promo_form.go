package handlers

import (
	"strings"

	"github.com/kozalosev/goSadTgBot/wizard"
)

func addPromoFieldsFromArgs(form wizard.Wizard, args []string) {
	addPromoFieldFromArg(form, fieldPromo, args, 0)
	addPromoFieldFromArg(form, fieldLength, args, 1)
	addPromoFieldFromArg(form, fieldCapacity, args, 2)

	switch {
	case len(args) >= 5:
		form.AddPrefilledField(fieldSetDates, actionSetBothDates)
		form.AddPrefilledField(fieldSince, args[3])
		form.AddPrefilledField(fieldUntil, args[4])
	case len(args) == 4 && isEndlessArg(args[3]):
		form.AddPrefilledField(fieldSetDates, actionEndlessPromo)
		form.AddEmptyField(fieldSince, wizard.Text)
		form.AddEmptyField(fieldUntil, wizard.Text)
	case len(args) == 4:
		form.AddPrefilledField(fieldSetDates, actionFromNowToDate)
		form.AddEmptyField(fieldSince, wizard.Text)
		form.AddPrefilledField(fieldUntil, args[3])
	default:
		form.AddEmptyField(fieldSetDates, wizard.Text)
		form.AddEmptyField(fieldSince, wizard.Text)
		form.AddEmptyField(fieldUntil, wizard.Text)
	}
}

func addPromoFieldFromArg(form wizard.Wizard, field string, args []string, index int) {
	if len(args) > index {
		form.AddPrefilledField(field, args[index])
		return
	}
	form.AddEmptyField(field, wizard.Text)
}

func isEndlessArg(arg string) bool {
	switch strings.ToLower(arg) {
	case "endless", "forever", "no", "none", "-", "бессрочно", "бессрочный", "нет":
		return true
	default:
		return false
	}
}
