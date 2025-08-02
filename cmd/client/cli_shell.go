package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// ShellCMD returns the Cobra command that launches the interactive shell mode.
func (g *GophKeeper) ShellCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "shell",
		Short: "Интерактивная оболочка GophKeeper",
		RunE: g.withAuth(func(cmd *cobra.Command, args []string) error {
			fmt.Print("\n💻 type `exit` to quit\n")
			return g.shellLoop()
		}),
	}
}

// shellLoop runs the REPL (read-eval-print loop) for interactive command execution.
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

		if err := g.processShellCommand(args); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Println("❌", err)
		}
	}

	return reader.Err()
}

func (g *GophKeeper) processShellCommand(args []string) error {
	switch args[0] {
	case "exit", "quit", "q":
		os.Exit(0)
		return io.EOF

	case "login":
		if len(args) < 2 {
			return errors.New("пример: login <username> <password>")
		}
		return g.LoginCMD().RunE(g.rootCmd, args)
	case "register":
		if len(args) < 2 {
			return errors.New("пример: register <username> <password>")
		}
		return g.RegisterCMD().RunE(g.rootCmd, args)
	case "contexts":
		return g.ContextListCMD().RunE(g.rootCmd, args)
	case "use":
		if len(args) < 2 {
			return errors.New("пример: use <ctx name>")
		}
		return g.ContextUseCMD().RunE(g.rootCmd, args)
	case "list":
		return g.VaultListCMD().RunE(g.rootCmd, nil)

	case "create":
		return g.NewVaultCMD().RunE(g.rootCmd, nil)

	case "get":
		if len(args) < 2 {
			return errors.New("пример: get <id>")
		}
		return g.VaultShowCMD().RunE(g.rootCmd, args)

	case "delete":
		if len(args) < 2 {
			return errors.New("пример: delete <id>")
		}
		return g.VaultDeleteCMD().RunE(g.rootCmd, args)

	case "help", "?", "version", "v":
		g.printBanner()

		printHelp()
		return nil

	default:
		return fmt.Errorf("неизвестная команда: %s", args[0])
	}
}

// printHelp displays the list of supported commands in the interactive shell.
func printHelp() {
	fmt.Println(`🔧 Команды:
login              войти в аккаунт или создать новый
register           зарегистрировать новый аккаунт
contexts           список всех контекстов
use <name>         сменить контекст
list               показать все записи
get <id>           показать запись по ID
delete <id>        удалить запись по ID
create             создать новую запись
exit / quit / q    выйти из программы
help / version / ? список команд`)
}
