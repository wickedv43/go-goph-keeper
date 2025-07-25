package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func (g *GophKeeper) ShellCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "shell",
		Short: "Интерактивная оболочка GophKeeper",
		RunE: g.withAuth(func(cmd *cobra.Command, args []string) error {
			g.printBanner() // banner)))
			fmt.Print("\n💻 type `exit` to quit\n")
			return g.shellLoop()
		}),
	}
}

func (g *GophKeeper) shellLoop() error {
	reader := bufio.NewScanner(os.Stdin)

	cfg, _ := g.storage.GetConfig()
	currentCtx := cfg.Current

	for {
		fmt.Printf("[%s] > ", currentCtx)
		if !reader.Scan() {
			break
		}
		line := strings.TrimSpace(reader.Text())
		args := strings.Split(line, " ")

		switch args[0] {
		case "exit", "quit", "q":
			os.Exit(0)
			return nil

		case "list":
			if err := g.VaultListCMD().RunE(g.rootCmd, nil); err != nil {
				fmt.Println("❌", err)
			}

		case "create":
			if err := g.NewVaultCMD().RunE(g.rootCmd, nil); err != nil {
				fmt.Println("❌", err)
			}

		case "get":
			if len(args) < 2 {
				fmt.Println("❌ Пример: show <id>")
				continue
			}

			if err := g.VaultShowCMD().RunE(g.rootCmd, args); err != nil {
				fmt.Println("❌", err)
			}

		case "delete":
			if len(args) < 2 {
				fmt.Println("❌ Пример: delete <id>")
				continue
			}

			if err := g.VaultDeleteCMD().RunE(g.rootCmd, args); err != nil {
				fmt.Println("❌", err)
			}
		case "contexts":
			if err := g.ContextListCMD().RunE(g.rootCmd, nil); err != nil {
				fmt.Println("❌", err)
			}

		case "use":
			if len(args) < 2 {
				fmt.Println("❌ Пример: use <name>")
				continue
			}
			if err := g.ContextUseCMD().RunE(g.rootCmd, args); err != nil {
				fmt.Println("❌", err)
			}
		case "login":
			if err := g.LoginCMD().RunE(g.rootCmd, nil); err != nil {
				fmt.Println("❌", err)
			}

		case "register":
			if err := g.RegisterCMD().RunE(g.rootCmd, nil); err != nil {
				fmt.Println("❌", err)
			}
		case "me":
			c, _ := g.storage.GetConfig()
			token, _ := g.storage.GetCurrentToken()
			key, _ := g.storage.GetCurrentKey()
			fmt.Println(c.Current)
			fmt.Println(token, key)

		case "help", "?":
			printHelp()
		default:
			fmt.Println("🤔 Неизвестная команда. Введите `help`")
		}
	}

	return reader.Err()
}

func printHelp() {
	fmt.Println(`🔧 Команды:
  login            войти в аккаунт или создать новый
  contexts         список всех контекстов
  use <name>       сменить контекст
  list            показать все записи
  get <id>        показать запись по ID
  delete <id>      удалить запись по ID
  create           создать новую запись
  logout           выйти из аккаунта
  exit             выйти из программы`)
}
