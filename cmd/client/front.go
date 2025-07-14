package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (g *GophKeeper) startMenu() *tview.List {
	menu := tview.NewList().
		AddItem("Login", "", 'l', func() {
			g.showLogin()
		}).
		AddItem("Register", "", 'r', func() {
			g.showRegister()
		}).
		AddItem("Quit", "", 'q', func() {
			g.tui.Stop()
		})
	menu.SetBorder(true).SetTitle("GophKeeper TUI").SetTitleAlign(tview.AlignLeft)
	return menu
}

func (g *GophKeeper) mainMenu() *tview.List {
	menu := tview.NewList().
		AddItem("Create", "", 'c', func() {
			g.showLogin()
		}).
		AddItem("List", "", 'l', func() {
			g.showRegister()
		}).
		AddItem("Quit", "", 'q', func() {
			g.tui.Stop()
		})
	menu.SetBorder(true).SetTitle("GophKeeper").SetTitleAlign(tview.AlignLeft)
	return menu
}

func (g *GophKeeper) showLogin() {
	form := tview.NewForm()
	errorText := tview.NewTextView().
		SetTextColor(tcell.ColorRed).
		SetDynamicColors(true).
		SetWrap(true)

	form.
		AddInputField("Login", "", 30, nil, nil).
		AddPasswordField("Password", "", 30, '*', nil).
		AddButton("Submit", func() {
			login := form.GetFormItemByLabel("Login").(*tview.InputField).GetText()
			password := form.GetFormItemByLabel("Password").(*tview.InputField).GetText()

			if err := g.Login(login, password); err != nil {
				errorText.SetText("[red]" + err.Error())
			} else {

				// Переход к следующему экрану после успешного логина
				g.tui.SetRoot(g.mainMenu(), true)
			}
		}).
		AddButton("Back", func() {
			g.tui.SetRoot(g.startMenu(), true)
		})

	form.SetBorder(true).SetTitle("Login").SetTitleAlign(tview.AlignLeft)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(errorText, 2, 0, false)

	g.tui.SetRoot(layout, true)
}

func (g *GophKeeper) showRegister() {
	form := tview.NewForm()
	errorText := tview.NewTextView().
		SetTextColor(tcell.ColorRed).
		SetDynamicColors(true).
		SetWrap(true)

	form.
		AddInputField("Login", "", 30, nil, nil).
		AddPasswordField("Password", "", 30, '*', nil).
		AddButton("Submit", func() {
			login := form.GetFormItemByLabel("Login").(*tview.InputField).GetText()
			password := form.GetFormItemByLabel("Password").(*tview.InputField).GetText()

			// Пример вызова регистрации
			if err := g.Register(login, password); err != nil {
				errorText.SetText("[red]" + err.Error())
			} else {
				// Успешная регистрация — можно вернуть на главную
				g.tui.SetRoot(g.startMenu(), true)
			}
		}).
		AddButton("Back", func() {
			g.tui.SetRoot(g.startMenu(), true)
		})

	form.SetBorder(true).SetTitle("Register").SetTitleAlign(tview.AlignLeft)

	// Компоновка с полем для ошибок
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(errorText, 2, 0, false)

	g.tui.SetRoot(layout, true)
}
