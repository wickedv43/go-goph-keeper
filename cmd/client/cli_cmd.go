package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/sqweek/dialog"
	"github.com/wickedv43/go-goph-keeper/cmd/client/kv"

	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"golang.org/x/term"
)

func (g *GophKeeper) LoginCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Вход в GophKeeper",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("🔐 Login: ")
			var login string
			if _, err := fmt.Scanln(&login); err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}

			fmt.Print("🔐 Password: ")
			passBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("ошибка чтения пароля: %w", err)
			}
			fmt.Println()

			password := string(passBytes)

			if err = g.Login(login, password); err != nil {
				return fmt.Errorf("ошибка входа: %w", err)
			}

			return g.shellLoop()
		},
	}

	return cmd
}

func (g *GophKeeper) RegisterCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Регистрация в GophKeeper",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("🔐 Login: ")
			var login string
			if _, err := fmt.Scanln(&login); err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}

			fmt.Print("🔐 Password: ")
			passBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("ошибка чтения пароля: %w", err)
			}
			fmt.Println()

			login = strings.TrimSpace(login)
			password := string(passBytes)

			if err = g.Register(login, password); err != nil {
				return fmt.Errorf("ошибка регистрации: %w", err)
			}

			return g.shellLoop()
		},
	}

	return cmd
}

func (g *GophKeeper) NewVaultCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create new record in GophKeeper",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := &pb.VaultRecord{}

			fmt.Print("Title: ")
			if _, err := fmt.Scanln(&v.Title); err != nil {
				return fmt.Errorf("ошибка чтения названия: %w", err)
			}

			fmt.Print("Types  \"login\", \"note\", \"card\" or \"binary\"  ")
			fmt.Print("Type: ")
			if _, err := fmt.Scanln(&v.Type); err != nil {
				return fmt.Errorf("ошибка чтения логина: %w", err)
			}

			switch v.Type {
			case "login":
				var err error

				v, err = vaultLoginPass(v)
				if err != nil {
					return err
				}
			case "note":
				var err error

				v, err = vaultNote(v)
				if err != nil {
					return err
				}
			case "card":
				var err error

				v, err = vaultCard(v)
				if err != nil {
					return err
				}
			case "binary":
				var err error

				v, err = vaultBinary(v)
				if err != nil {
					return err
				}
			}

			//fmt.Println("Enter some tag: ")
			//if _, err := fmt.Scanln(&v.Metadata); err != nil {
			//	return err
			//}

			//TODO: input metadata)))
			v.Metadata = "{}"

			_, err := g.VaultCreate(v)
			if err != nil {
				return err
			}

			return g.shellLoop()
		},
	}

	return cmd
}

func vaultLoginPass(v *pb.VaultRecord) (*pb.VaultRecord, error) {
	var (
		d   kv.LoginPass
		err error
	)

	fmt.Print("Login: ")
	if _, err = fmt.Scanln(&d.Login); err != nil {
		return v, fmt.Errorf("ошибка чтения названия: %w", err)
	}
	fmt.Print("Password: ")
	if _, err = fmt.Scanln(&d.Password); err != nil {
		return v, fmt.Errorf("ошибка чтения названия: %w", err)
	}

	//TODO: CRYPTO???
	v.EncryptedData, err = json.Marshal(d)
	if err != nil {
		return v, err
	}

	return v, nil
}

func vaultNote(v *pb.VaultRecord) (*pb.VaultRecord, error) {
	var (
		d   kv.Note
		err error
	)

	fmt.Print("Note: ")
	if _, err = fmt.Scanln(&d.Text); err != nil {
		return v, fmt.Errorf("ошибка чтения названия: %w", err)
	}

	//TODO: CRYPTO???
	v.EncryptedData, err = json.Marshal(d)
	if err != nil {
		return v, err
	}

	return v, nil
}

func vaultCard(v *pb.VaultRecord) (*pb.VaultRecord, error) {
	var (
		d   kv.Card
		err error
	)

	fmt.Print("Number: ")
	if _, err = fmt.Scanln(&d.Number); err != nil {
		return v, fmt.Errorf("ошибка чтения названия: %w", err)
	}

	fmt.Print("Date: ")
	if _, err = fmt.Scanln(&d.Date); err != nil {
		return v, fmt.Errorf("ошибка чтения названия: %w", err)
	}

	fmt.Print("CVV: ")
	if _, err = fmt.Scanln(&d.CVV); err != nil {
		return v, fmt.Errorf("ошибка чтения названия: %w", err)
	}

	//TODO: CRYPTO???
	v.EncryptedData, err = json.Marshal(d)
	if err != nil {
		return v, err
	}

	return v, nil
}

func vaultBinary(v *pb.VaultRecord) (*pb.VaultRecord, error) {
	path, err := dialog.File().Title("Выберите файл").Load()
	if err != nil {
		return v, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return v, fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	// 🔐 TODO: encrypt(data) если нужно
	v.EncryptedData = data

	// например, сохранить имя файла в metadata
	meta := map[string]string{"filename": filepath.Base(path)}
	metaJSON, _ := json.Marshal(meta)
	v.Metadata = string(metaJSON)

	return v, nil
}

func (g *GophKeeper) VaultListCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "vault list",
		Short: "Показать все записи в хранилище",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := g.VaultList()
			if err != nil {
				return fmt.Errorf("ошибка получения списка записей: %w", err)
			}

			if len(resp.Vaults) == 0 {
				fmt.Println("🔒 Хранилище пусто.")
				return g.shellLoop()
			}

			for _, v := range resp.Vaults {
				fmt.Printf("📄 [%s] %s (ID: %d)\n", v.Type, v.Title, v.Id)
			}

			return g.shellLoop()
		},
	}
}

func (g *GophKeeper) VaultShowCMD(id uint64) error {
	v, err := g.VaultGet(id)
	if err != nil {
		return fmt.Errorf("не удалось получить запись: %w", err)
	}

	fmt.Printf("📄 ID: %d\n", v.Id)
	fmt.Printf("📌 Type: %s\n", v.Type)
	fmt.Printf("📝 Title: %s\n", v.Title)

	switch v.Type {
	case "login":
		var d kv.LoginPass
		if err := json.Unmarshal(v.EncryptedData, &d); err != nil {
			fmt.Println("🔐 Ошибка чтения логина:", err)
		} else {
			fmt.Printf("👤 Login: %s\n", d.Login)
			fmt.Printf("🔑 Password: %s\n", d.Password)
		}

	case "note":
		var d kv.Note
		if err := json.Unmarshal(v.EncryptedData, &d); err != nil {
			fmt.Println("📝 Ошибка чтения заметки:", err)
		} else {
			fmt.Printf("📝 Note: %s\n", d.Text)
		}

	case "card":
		var d kv.Card
		if err := json.Unmarshal(v.EncryptedData, &d); err != nil {
			fmt.Println("💳 Ошибка чтения карты:", err)
		} else {
			fmt.Printf("💳 Number: %s\n", d.Number)
			fmt.Printf("📆 Date: %s\n", d.Date)
			fmt.Printf("🔒 CVV: %s\n", d.CVV)
		}

	case "binary":
		var meta map[string]string
		filename := "file.bin"

		if err = json.Unmarshal([]byte(v.Metadata), &meta); err == nil {
			filename = meta["filename"]
			fmt.Printf("📎 File: %s (%d байт)\n", filename, len(v.EncryptedData))
		} else {
			fmt.Printf("📎 Binary file (%d байт)\n", len(v.EncryptedData))
		}

		// 👉 Спросить пользователя
		fmt.Print("💾 Хотите сохранить файл? (y/n): ")
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) != "y" {
			break
		}

		// 👉 Выбрать место сохранения
		savePath, err := dialog.File().Title("Сохранить файл как...").Save()
		if err != nil {
			fmt.Println("❌ Не удалось выбрать путь:", err)
			break
		}

		err = os.WriteFile(savePath, v.EncryptedData, 0644)
		if err != nil {
			fmt.Println("❌ Ошибка сохранения:", err)
		} else {
			fmt.Println("✅ Файл сохранён в", savePath)
		}

	default:
		fmt.Println("🤷 Неизвестный тип данных")
	}

	return nil
}

func (g *GophKeeper) VaultDeleteCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [id]",
		Short: "Удалить запись по ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("неверный ID: %w", err)
			}
			_, err = g.VaultDelete(id)
			if err != nil {
				return fmt.Errorf("ошибка удаления: %w", err)
			}
			fmt.Println("✅ Запись удалена.")
			return g.shellLoop()
		},
	}
}
