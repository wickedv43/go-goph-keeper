package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	"github.com/wickedv43/go-goph-keeper/pkg/crypto"
)

func (g *GophKeeper) LoginCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Вход в GophKeeper",
		RunE: func(cmd *cobra.Command, args []string) error {
			var login, password string
			out := cmd.OutOrStdout()

			_, _ = fmt.Fprint(out, "🔐 Login: ")
			if _, err := fmt.Scanln(&login); err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}

			_, _ = fmt.Fprint(out, "🔐 Password: ")
			_, err := fmt.Scanln(&password) //term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("ошибка чтения пароля: %w", err)
			}
			_, _ = fmt.Fprintln(out, "")

			if _, err = g.Login(login, password); err != nil {
				return err
			}

			key, err := g.storage.GetCurrentKey()
			if err != nil && errors.Is(err, kv.ErrEmptyKey) {
				_, _ = fmt.Fprintln(out, "Введите мнемоническую фразу:")
				words := make([]string, 12)
				for i := range words {
					_, _ = fmt.Fprintf(out, "[%d]: ", i+1)
					if _, err = fmt.Scanln(&words[i]); err != nil {
						return fmt.Errorf("ошибка чтения слова: %w", err)
					}
				}
				mnemo := strings.Join(words, " ")
				key = crypto.GenerateSeed(mnemo, password)
				if err = g.storage.SaveKey(login, key); err != nil {
					return fmt.Errorf("ошибка сохранения ключа: %w", err)
				}
			}

			return nil
		},
	}
	return cmd
}

func (g *GophKeeper) RegisterCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Регистрация в GophKeeper",
		RunE: func(cmd *cobra.Command, args []string) error {
			var login, password string
			out := cmd.OutOrStdout()

			_, _ = fmt.Fprintln(out, "🔐 Login: ")
			if _, err := fmt.Scanln(&login); err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}

			_, _ = fmt.Fprint(out, "🔐 Password: ")
			_, err := fmt.Scanln(&password) //term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("ошибка чтения пароля: %w", err)
			}
			_, _ = fmt.Fprintln(out, "")

			login = strings.TrimSpace(login)

			words, err := g.Register(login, password)
			if err != nil {
				return fmt.Errorf("ошибка регистрации: %w", err)
			}

			_, _ = fmt.Fprintln(out, "💾 Save this phrase:")
			for row := 0; row < 4; row++ {
				for col := 0; col < 3; col++ {
					index := row + col*4
					_, _ = fmt.Fprintf(out, "%2d. %-8s  ", index+1, words[index])
				}
				_, _ = fmt.Fprintln(out, "")
			}

			return nil
		},
	}

	return cmd
}
