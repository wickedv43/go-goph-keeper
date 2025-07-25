package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (g *GophKeeper) ContextUseCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "use [context-name]",
		Short: "Switch context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[1]
			cfg, err := g.storage.GetConfig()
			if err != nil {
				return err
			}

			if _, ok := cfg.Contexts[name]; !ok {
				return fmt.Errorf("контекст %q не найден", name)
			}

			cfg.Current = name
			if err = g.storage.SetConfig(cfg); err != nil {
				return fmt.Errorf("не удалось сохранить конфиг: %w", err)
			}

			fmt.Println("Context switched ✅")

			return g.shellLoop()
		},
	}
}

func (g *GophKeeper) ContextListCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "contexts",
		Short: "Contexts list",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := g.storage.GetConfig()
			if err != nil {
				return err
			}

			if len(cfg.Contexts) == 0 {
				fmt.Println("Contexts empty 📭")
				return nil
			}

			for name := range cfg.Contexts {
				active := ""
				if name == cfg.Current {
					active = " (in use)"
				}
				fmt.Printf("  - %s%s\n", name, active)
			}

			return g.shellLoop()
		},
	}
}

//
//func (g *GophKeeper) enterMnemo() error {
//
//	fmt.Println("Enter mnemonic: ")
//	words := make([]string, 12)
//	for i := 0; i < len(words); i++ {
//		var word string
//		fmt.Printf("[%d]: ", i+1)
//
//		if _, err = fmt.Scanln(&word); err != nil {
//			return fmt.Errorf("word reading error: %w", err)
//			os.Exit(0)
//		}
//		words[i] = word
//	}
//
//	mnemo := strings.Join(words, " ")
//	key := crypto.GenerateSeed(mnemo, password)
//
//	err = g.storage.SaveKey(login, key)
//	if err != nil {
//		return fmt.Errorf("ошибка ввода фразы: %w", err)
//	}
//}
