package main

import (
	"fmt"

	"github.com/rivo/tview"
)

func buildModal(form tview.Primitive, errorText *tview.TextView, width, height int) tview.Primitive {
	build := tview.NewTextView().
		SetText(fmt.Sprintf("%s / %s", buildDate, buildVersion)).
		SetTextAlign(tview.AlignRight)

	modal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, height, 0, true).
		AddItem(build, 1, 0, false).
		AddItem(errorText, 1, 0, false)

	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).      // отступ слева
		AddItem(modal, width, 0, true). // центр (модалка фиксированной ширины)
		AddItem(nil, 0, 1, false)       // отступ справа

	wrapper := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).          // отступ сверху
		AddItem(centered, height, 0, true). // центр по вертикали
		AddItem(nil, 0, 1, false)           // отступ снизу

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
