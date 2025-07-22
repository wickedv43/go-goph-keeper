package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (g *GophKeeper) LoginCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login [email] [password]",
		Short: "Вход в GophKeeper",
		Args:  cobra.ExactArgs(2), // ждём два позиционных аргумента
		RunE: func(cmd *cobra.Command, args []string) error {
			login := args[0]
			password := args[1]

			if err := g.Login(login, password); err != nil {
				return fmt.Errorf("ошибка входа: %w", err)
			}

			g.printBanner()

			return nil
		},
	}

	return cmd
}
