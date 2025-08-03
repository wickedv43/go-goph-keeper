package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ContextUseCMD returns a Cobra command that switches the current context to the specified name.
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

			fmt.Fprintf(cmd.OutOrStdout(), "Context switched ✅")

			return nil
		},
	}
}

// ContextListCMD returns a Cobra command that lists all available contexts and highlights the active one.
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
				fmt.Fprintln(cmd.OutOrStdout(), "Contexts empty 📭")
				return nil
			}

			for name := range cfg.Contexts {
				active := ""
				if name == cfg.Current {
					active = " (in use)"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s%s\n", name, active)
			}

			return nil
		},
	}
}
