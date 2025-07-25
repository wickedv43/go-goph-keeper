package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (g *GophKeeper) withAuth(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			switch args[0] {
			case "help", "?", "login", "register":
				return fn(cmd, args)
			}
		}

		_, err := g.storage.GetConfig()
		if err != nil {
			fmt.Println("🔐 You are not authorized. Type `login` or `register`")
			return g.shellLoop()
		}

		return fn(cmd, args)
	}
}
