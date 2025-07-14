package main

import (
	"fmt"

	"github.com/rivo/tview"
)

func buildModal(form tview.Primitive, errorText *tview.TextView, width, height, leftOffset int) tview.Primitive {
	build := tview.NewTextView().SetText(fmt.Sprintf("%s / %s", buildDate, buildVersion)).SetTextAlign(tview.AlignRight)
	modal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, height, 1, true).
		AddItem(build, 1, 0, false).    // футер (версия, сборка и т.п.)
		AddItem(errorText, 1, 0, false) // строка ошибок

	centered := tview.NewFlex().
		AddItem(nil, leftOffset, 0, false). // отступ слева
		AddItem(modal, width, 0, true).     // модалка фиксированной ширины
		AddItem(nil, 0, 1, false)           // всё остальное справа

	wrapper := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false). // сверху
		AddItem(centered, height, 0, true).
		AddItem(nil, 0, 1, false) // снизу

	return wrapper
}

func dropdownIndex(rt string) int {
	switch rt {
	case "login":
		return 0
	case "note":
		return 1
	case "card":
		return 2
	case "binary":
		return 3
	default:
		return 0
	}
}
