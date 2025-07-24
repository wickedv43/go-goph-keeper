package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

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
		Use:   "list",
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

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTYPE\tTITLE\tUPDATED AT\tTAGS")

			for _, v := range resp.Vaults {
				// Парсим дату
				var formatted string
				if t, err := time.Parse(time.RFC3339, v.UpdatedAt); err == nil {
					formatted = t.Format("01-02-2006")
				} else {
					formatted = "-"
				}

				// Парсим метаданные
				var meta map[string]string
				var tags string
				if err = json.Unmarshal([]byte(v.Metadata), &meta); err == nil {
					// Пример: берём все значения и соединяем
					for k, v := range meta {
						tags += fmt.Sprintf("%s=%s,", k, v)
					}
					tags = strings.TrimSuffix(tags, ",")
				} else {
					tags = "-"
				}

				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", v.Id, v.Type, v.Title, formatted, tags)
			}

			w.Flush()

			return g.shellLoop()
		},
	}
}

func (g *GophKeeper) VaultShowCMD(id uint64) error {
	v, err := g.VaultGet(id)
	if err != nil {
		return fmt.Errorf("не удалось получить запись: %w", err)
	}

	// Метаданные
	var meta map[string]string
	_ = json.Unmarshal([]byte(v.Metadata), &meta)

	// Дата
	updated := v.UpdatedAt
	if t, err := time.Parse(time.RFC3339, updated); err == nil {
		updated = t.Format("2006-01-02 15:04:05")
	}

	// Шапка
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Printf(" %-14s : %v\n", "ID", v.Id)
	fmt.Printf(" %-14s : %v\n", "Тип", v.Type)
	fmt.Printf(" %-14s : %v\n", "Заголовок", v.Title)
	fmt.Printf(" %-14s : %v\n", "Обновлено", updated)
	if len(meta) > 0 {
		for k, val := range meta {
			fmt.Printf(" %-14s : %v\n", k, val)
		}
	}
	fmt.Println("═══════════════════════════════════════════════")

	// Данные
	fmt.Println("🔐 Данные:")
	switch v.Type {
	case "login":
		var d kv.LoginPass
		if err := json.Unmarshal(v.EncryptedData, &d); err == nil {
			fmt.Printf(" 👤 Login     : %s\n", d.Login)
			fmt.Printf(" 🔑 Password  : %s\n", d.Password)
		} else {
			fmt.Println("❌ Ошибка чтения login/pass:", err)
		}

	case "note":
		var d kv.Note
		if err := json.Unmarshal(v.EncryptedData, &d); err == nil {
			fmt.Println(" 📝 Note:")
			fmt.Println(" ---------------------------------------------")
			fmt.Println(d.Text)
			fmt.Println(" ---------------------------------------------")
		} else {
			fmt.Println("❌ Ошибка чтения заметки:", err)
		}

	case "card":
		var d kv.Card
		if err := json.Unmarshal(v.EncryptedData, &d); err == nil {
			fmt.Printf(" 💳 Number    : %s\n", d.Number)
			fmt.Printf(" 📆 Date      : %s\n", d.Date)
			fmt.Printf(" 🔒 CVV       : %s\n", d.CVV)
		} else {
			fmt.Println("❌ Ошибка чтения карты:", err)
		}

	case "binary":
		filename := "file.bin"
		if meta != nil && meta["filename"] != "" {
			filename = meta["filename"]
		}
		fmt.Printf(" 📎 File      : %s (%d байт)\n", filename, len(v.EncryptedData))

		fmt.Print("💾 Хотите сохранить файл? (y/n): ")
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) != "y" {
			break
		}

		savePath, err := dialog.File().Title("Сохранить файл как...").Save()
		if err != nil {
			fmt.Println("❌ Не удалось выбрать путь:", err)
			break
		}

		if err := os.WriteFile(savePath, v.EncryptedData, 0644); err != nil {
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
