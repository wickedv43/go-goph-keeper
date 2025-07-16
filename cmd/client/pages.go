package main

import (
	"encoding/hex"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
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

			modal := tview.NewModal().
				SetText("Регистрация прошла успешно!").
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					g.pages.SwitchToPage("Login")
					g.pages.RemovePage("SuccessModal") // удалить модалку после показа
				})

			// Обёртка: сдвигаем модалку влево, например, оставляя справа больше места
			centered := tview.NewFlex().
				AddItem(nil, 5, 1, false).  // Остальное место справа
				AddItem(modal, 60, 0, true) // Ширина модалки

			g.pages.AddPage("SuccessModal", centered, true, true)
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

	btns := tview.NewList().AddItem("Add", "", 'a', func() {
		g.pages.AddAndSwitchToPage("VaultCreate", g.NewVaultPage(), true)

	})

	// Grid: 2 столбца — список и детальный просмотр
	grid := tview.NewGrid().
		SetRows(0).        // одна строка
		SetColumns(30, 0). // список: 30 ширина, остальное — деталь
		AddItem(list, 0, 0, 1, 1, 0, 0, false).
		AddItem(details, 0, 1, 1, 1, 0, 0, false).
		AddItem(btns, 1, 0, 1, 1, 0, 0, true)

	return grid
}

func (g *GophKeeper) NewVaultPage() tview.Primitive {
	form := tview.NewForm()
	errorText := tview.NewTextView().
		SetTextColor(tcell.ColorRed).
		SetDynamicColors(true)

	var recordType string
	var fields map[string]*tview.InputField

	var saveFunc func()
	var cancelFunc func()

	// Динамически перестраиваемая форма
	updateForm := func(rt string) {
		recordType = rt
		form.Clear(false)
		fields = map[string]*tview.InputField{}

		// Title
		titleField := tview.NewInputField().SetLabel("Title")
		fields["Title"] = titleField
		form.AddFormItem(titleField)

		// Type
		form.AddDropDown("Type", []string{"login", "note", "card", "binary"}, dropdownIndex(rt), func(opt string, _ int) {
		})

		// Поля по типу
		switch rt {
		case "login":
			fields["Login"] = tview.NewInputField().SetLabel("Login")
			fields["Password"] = tview.NewInputField().SetLabel("Password")
			form.AddFormItem(fields["Login"])
			form.AddFormItem(fields["Password"])
		case "note":
			fields["Note"] = tview.NewInputField().SetLabel("Note")
			form.AddFormItem(fields["Note"])
		case "card":
			fields["Card Number"] = tview.NewInputField().SetLabel("Card Number")
			fields["Expiry"] = tview.NewInputField().SetLabel("Expiry")
			form.AddFormItem(fields["Card Number"])
			form.AddFormItem(fields["Expiry"])
		case "binary":
			fields["File Name"] = tview.NewInputField().SetLabel("File Name")
			form.AddFormItem(fields["File Name"])
		}

		// Общие поля
		fields["Metadata"] = tview.NewInputField().SetLabel("Metadata")
		fields["EncryptedData"] = tview.NewInputField().SetLabel("EncryptedData (hex)")
		form.AddFormItem(fields["Metadata"])
		form.AddFormItem(fields["EncryptedData"])

		form.
			AddButton("Save", saveFunc).
			AddButton("Cancel", cancelFunc).
			SetBorder(true).
			SetTitle(" New Vault (" + rt + ") ").
			SetTitleAlign(tview.AlignLeft)
	}

	// Обработчики
	saveFunc = func() {
		title := fields["Title"].GetText()
		metadata := fields["Metadata"].GetText()
		dataHex := fields["EncryptedData"].GetText()

		data, err := hex.DecodeString(dataHex)
		if err != nil {
			errorText.SetText("[red]Ошибка: EncryptedData не в hex-формате")
			return
		}
		_, err = g.client.CreateVault(g.authCtx(), &pb.CreateVaultRequest{Record: &pb.VaultRecord{
			Title:         title,
			Type:          recordType,
			Metadata:      metadata,
			EncryptedData: data,
		},
		})
		if err != nil {
			if st, ok := status.FromError(err); ok {
				errorText.SetText("[red]Ошибка: " + st.Message())
			} else {
				errorText.SetText("[red]Ошибка: " + err.Error())
			}
			return
		}

		g.pages.RemovePage("VaultList")
		g.pages.AddAndSwitchToPage("VaultList", g.VaultPage(), true)
	}

	cancelFunc = func() {
		g.pages.SwitchToPage("VaultList")
	}

	// Начальное построение формы
	updateForm("login")

	// Layout
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(errorText, 2, 0, false)

	return layout
}
