package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc/status"
)

func (g *GophKeeper) LoginPage() tview.Primitive {
	errorText := tview.NewTextView().
		SetTextColor(tcell.ColorRed).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	form := tview.NewForm()

	form.
		AddInputField("Login", "", 20, nil, nil).
		AddPasswordField("Password", "", 20, '*', nil).
		AddButton("SignIn", func() {
			login := form.GetFormItemByLabel("Login").(*tview.InputField).GetText()
			password := form.GetFormItemByLabel("Password").(*tview.InputField).GetText()

			if err := g.Login(login, password); err != nil {
				st, ok := status.FromError(err)
				if ok {
					errorText.SetText("[red]Ошибка: " + st.Message())
				} else {
					errorText.SetText("[red]Ошибка: " + err.Error())
				}
				return
			}
		}).
		AddButton("SignUp", func() {
			g.pages.SwitchToPage("Register")
		}).
		SetBorder(true).
		SetTitle("GophKeeper").
		SetTitleAlign(tview.AlignCenter)

	// «Модалка» с формой и текстом ошибки
	modal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 10, 1, true).     // высота формы (примерно 10)
		AddItem(errorText, 2, 0, false) // текст ошибки высотой 2 строки

	// Центрируем модалку по горизонтали
	leftOffset := 5
	centered := tview.NewFlex().
		AddItem(nil, leftOffset, 0, false). // левый отступ
		AddItem(modal, 60, 0, true).        // ширина модалки
		AddItem(nil, 0, 1, false)           // остальное место справа

	// Центрируем по вертикали
	wrapper := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).      // отступ сверху
		AddItem(centered, 15, 0, true). // высота модалки 15
		AddItem(nil, 0, 1, false)       // отступ снизу

	return wrapper
}

func (g *GophKeeper) RegisterPage() tview.Primitive {
	errorText := tview.NewTextView().
		SetTextColor(tcell.ColorRed).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	form := tview.NewForm()

	form.
		AddInputField("Login", "", 20, nil, nil).
		AddPasswordField("Password", "", 20, '*', nil).
		AddPasswordField("Repeat Password", "", 20, '*', nil).
		AddButton("Register", func() {
			login := form.GetFormItemByLabel("Login").(*tview.InputField).GetText()
			fPass := form.GetFormItemByLabel("Password").(*tview.InputField).GetText()
			sPass := form.GetFormItemByLabel("Repeat Password").(*tview.InputField).GetText()

			if fPass != sPass {
				errorText.SetText("[red]Ошибка: разные пароли")
				return
			}

			if err := g.Register(login, fPass); err != nil {
				st, ok := status.FromError(err)
				if ok {
					errorText.SetText("[red]Ошибка: " + st.Message())
				} else {
					errorText.SetText("[red]Ошибка: " + err.Error())
				}
				return
			}
		}).
		AddButton("Back", func() {
			g.pages.SwitchToPage("Login")
		}).
		SetBorder(true).
		SetTitle("Register").
		SetTitleAlign(tview.AlignLeft)

	// «Модалка» с формой и текстом ошибки
	modal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 13, 1, true).     // высота формы (примерно 10)
		AddItem(errorText, 2, 0, false) // текст ошибки высотой 2 строки

	// Центрируем модалку по горизонтали
	leftOffset := 5
	centered := tview.NewFlex().
		AddItem(nil, leftOffset, 0, false). // левый отступ
		AddItem(modal, 60, 0, true).        // ширина модалки
		AddItem(nil, 0, 1, false)           // остальное место справа

	// Центрируем по вертикали
	wrapper := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).      // отступ сверху
		AddItem(centered, 15, 0, true). // высота модалки 15
		AddItem(nil, 0, 1, false)       // отступ снизу

	return wrapper
}
