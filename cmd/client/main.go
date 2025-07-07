package main

import (
	"fmt"
	"log"

	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/wickedv43/go-goph-keeper/internal/api"
)

var (
	app *tview.Application
)

func main() {
	app = tview.NewApplication()

	cc, err := grpc.NewClient(
		"localhost:8080",
		// Используем insecure-коннект для тестов
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	// gRPC-клиент сервера Auth
	_ = pb.NewGophKeeperClient(cc)

	// главное меню
	menu := tview.NewList().
		AddItem("Register", "Зарегистрировать нового пользователя", 'r', func() {
			showRegister()
		}).
		AddItem("Quit", "Выход", 'q', func() {
			app.Stop()
		})

	menu.SetBorder(true).SetTitle("GophKeeper TUI")

	if err := app.SetRoot(menu, true).Run(); err != nil {
		panic(err)
	}
}

func showRegister() {
	var form *tview.Form

	form = tview.NewForm().
		AddInputField("Email", "", 30, nil, nil).
		AddPasswordField("Password", "", 30, '*', nil).
		AddButton("Submit", func() {
			email := form.GetFormItemByLabel("Email").(*tview.InputField).GetText()
			password := form.GetFormItemByLabel("Password").(*tview.InputField).GetText()
			//	doRegister(email, password)
			fmt.Println(email, password)
		}).
		AddButton("Back", func() {
			app.SetRoot(showMenu(), true)
		})
	form.SetBorder(true).SetTitle("Register").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(form, true)
}

func showMenu() *tview.List {
	menu := tview.NewList().
		AddItem("Register", "Зарегистрировать нового пользователя", 'r', func() {
			showRegister()
		}).
		AddItem("Quit", "Выход", 'q', func() {
			app.Stop()
		})
	menu.SetBorder(true).SetTitle("GophKeeper TUI").SetTitleAlign(tview.AlignLeft)
	return menu
}
