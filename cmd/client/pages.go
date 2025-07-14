package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc/status"
)

func (g *GophKeeper) LoginPage() tview.Primitive {
	errorText := tview.NewTextView().
		SetTextColor(tcell.ColorRed).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

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

			g.pages.AddAndSwitchToPage("Vault", g.VaultPage(), true)
		}).
		AddButton("SignUp", func() {
			g.pages.SwitchToPage("Register")
		}).
		SetBorder(true).
		SetTitle(" GophKeeper ").
		SetTitleAlign(tview.AlignCenter)

	return buildModal(form, errorText, 60, 10, 5)
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

			}
			g.pages.SwitchToPage("Login")
		}).
		AddButton("Back", func() {
			g.pages.SwitchToPage("Login")
		}).
		SetBorder(true).
		SetTitle(" Register ").
		SetTitleAlign(tview.AlignLeft)

	return buildModal(form, errorText, 60, 12, 5)
}

func (g *GophKeeper) VaultPage() tview.Primitive {
	// Левый список
	list := tview.NewList()

	// Правая панель — подробности
	details := tview.NewTextView()

	// Загружаем записи
	vaults, err := g.ListVaults()
	if err != nil {
		details.SetText("[red]Ошибка загрузки: " + err.Error())
	} else {
		for _, v := range vaults.Vaults {
			list.AddItem(v.Title, "", 0, func() {
				// Показать расшифрованную/отформатированную запись
				details.SetTitle(fmt.Sprintf(
					"[yellow]Тип: [white]%s\n[yellow]Метаданные: [white]%s\n[yellow]Encrypted: [white]%x",
					v.Type, v.Metadata, v.EncryptedData,
				))
			})
		}
	}

	details.
		SetDynamicColors(true).
		SetWrap(true).
		SetBorder(true).
		SetTitle(" Details ")

	list.
		ShowSecondaryText(false).
		SetBorder(true).
		SetTitle(" Vaults ")

	// Grid: 2 столбца — список и детальный просмотр
	grid := tview.NewGrid().
		SetRows(0).        // одна строка
		SetColumns(30, 0). // список: 30 ширина, остальное — деталь
		AddItem(list, 0, 0, 1, 1, 0, 0, true).
		AddItem(details, 0, 1, 1, 1, 0, 0, false)

	return grid
}
