package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func (g *GophKeeper) ShellCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "shell",
		Short: "Интерактивная оболочка GophKeeper",
		RunE: func(cmd *cobra.Command, args []string) error {
			g.printBanner() // banner)))
			fmt.Print("\n💻 type `exit` to quit\n")
			return g.shellLoop()
		},
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
			fmt.Println("👋 Пока!")
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
			id, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				fmt.Println("❌ Неверный ID:", err)
				continue
			}
			if err = g.VaultShowCMD(id); err != nil {
				fmt.Println("❌", err)
			}

		case "delete":
			if len(args) < 2 {
				fmt.Println("❌ Пример: delete <id>")
				continue
			}
			_, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				fmt.Println("❌ Неверный ID:", err)
				continue
			}
			if err = g.VaultDeleteCMD().RunE(g.rootCmd, []string{args[1]}); err != nil {
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
			contextName := args[1]
			if err := g.storage.UseContext(contextName); err != nil {
				fmt.Println("❌", err)
			} else {
				currentCtx = contextName
				fmt.Printf("✅ Контекст %q активирован\n", currentCtx)
			}

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
  list            показать все записи
  get <id>        показать запись по ID
  delete <id>      удалить запись по ID
  create           создать новую запись
  contexts         список всех контекстов
  use <name>       сменить контекст
  exit             выйти`)
}
