package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/sqweek/dialog"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/pkg/crypto"
)

func (g *GophKeeper) NewVaultCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create new record in GophKeeper",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			v := &pb.VaultRecord{}

			v, err := g.createVaultRecord(out)
			if err != nil {
				return err
			}

			switch v.Type {
			case "login":
				v, err = vaultLoginPass(v)
				if err != nil {
					return err
				}
			case "note":
				v, err = vaultNote(v)
				if err != nil {
					return err
				}
			case "card":
				v, err = vaultCard(v)
				if err != nil {
					return err
				}
			case "binary":
				v, err = vaultBinary(v)
				if err != nil {
					return err
				}
			}

			//crypto
			key, err := g.storage.GetCurrentKey()
			if err != nil {
				return err
			}

			v.EncryptedData, err = crypto.EncryptWithSeed(v.EncryptedData, key)
			if err != nil {
				return err
			}

			_, err = g.VaultCreate(v)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func (g *GophKeeper) createVaultRecord(out io.Writer) (*pb.VaultRecord, error) {
	v := &pb.VaultRecord{Metadata: "{}"}

	fmt.Fprintln(out, "Title: ")
	if _, err := fmt.Scanln(&v.Title); err != nil {
		return nil, fmt.Errorf("ошибка чтения названия: %w", err)
	}

	fmt.Fprintln(out, "Types  \"login\", \"note\", \"card\" or \"binary\"  ")
	fmt.Fprintln(out, "Type: ")
	if _, err := fmt.Scanln(&v.Type); err != nil {
		return nil, fmt.Errorf("ошибка чтения типа: %w", err)
	}

	return v, nil
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
			out := cmd.OutOrStdout()
			resp, err := g.VaultList()
			if err != nil {
				return fmt.Errorf("ошибка получения списка записей: %w", err)
			}

			if len(resp.Vaults) == 0 {
				fmt.Fprintln(out, "🔒 Хранилище пусто.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
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

			return nil
		},
	}
}

func (g *GophKeeper) VaultShowCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Показать запись в хранилище",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			id, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {

				return fmt.Errorf("❌ Неверный ID: %w", err)
			}

			v, err := g.VaultGet(id)
			if err != nil {
				return fmt.Errorf("не удалось получить запись: %w", err)
			}

			key, err := g.storage.GetCurrentKey()
			if err != nil {
				return err
			}
			v.EncryptedData, err = crypto.DecryptWithSeed(v.EncryptedData, key)
			if err != nil {
				return err
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
			fmt.Fprintln(out, "═══════════════════════════════════════════════")
			fmt.Fprintf(out, " %-14s : %v\n", "ID", v.Id)
			fmt.Fprintf(out, " %-14s : %v\n", "Тип", v.Type)
			fmt.Fprintf(out, " %-14s : %v\n", "Заголовок", v.Title)
			fmt.Fprintf(out, " %-14s : %v\n", "Обновлено", updated)
			if len(meta) > 0 {
				for k, val := range meta {
					fmt.Fprintf(out, " %-14s : %v\n", k, val)
				}
			}
			fmt.Fprintln(out, "═══════════════════════════════════════════════")

			// Данные
			fmt.Fprintln(out, "🔐 Данные:")
			switch v.Type {
			case "login":
				var d kv.LoginPass
				if err = json.Unmarshal(v.EncryptedData, &d); err == nil {
					fmt.Fprintf(out, " 👤 Login     : %s\n", d.Login)
					fmt.Fprintf(out, " 🔑 Password  : %s\n", d.Password)
				} else {
					fmt.Fprintln(out, "❌ Ошибка чтения login/pass:", err)
				}

			case "note":
				var d kv.Note
				if err := json.Unmarshal(v.EncryptedData, &d); err == nil {
					fmt.Fprintln(out, " 📝 Note:")
					fmt.Fprintln(out, " ---------------------------------------------")
					fmt.Fprintln(out, d.Text)
					fmt.Fprintln(out, " ---------------------------------------------")
				} else {
					fmt.Fprintln(out, "❌ Ошибка чтения заметки:", err)
				}

			case "card":
				var d kv.Card
				if err = json.Unmarshal(v.EncryptedData, &d); err == nil {
					fmt.Fprintf(out, " 💳 Number    : %s\n", d.Number)
					fmt.Fprintf(out, " 📆 Date      : %s\n", d.Date)
					fmt.Fprintf(out, " 🔒 CVV       : %s\n", d.CVV)
				} else {
					fmt.Fprintln(out, "❌ Ошибка чтения карты:", err)
				}

			case "binary":

				filename := "file.bin"
				if meta != nil && meta["filename"] != "" {
					filename = meta["filename"]
				}
				fmt.Fprintf(out, " 📎 File      : %s (%d байт)\n", filename, len(v.EncryptedData))

				fmt.Fprintln(out, "💾 Download? (y/n): ")
				var answer string
				fmt.Scanln(&answer)
				if strings.ToLower(answer) != "y" {
					return nil
				}

				var savePath string
				savePath, err = dialog.File().Title("Сохранить файл как...").Save()
				if err != nil {
					fmt.Fprintln(out, "❌ Не удалось выбрать путь:", err)
					break
				}

				// Если пользователь не указал расширение, добавим его
				if filepath.Ext(savePath) == "" {
					savePath += filepath.Ext(filename)
				}

				if err = os.WriteFile(savePath, v.EncryptedData, 0644); err != nil {
					fmt.Fprintln(out, "❌ Ошибка сохранения:", err)
				} else {
					fmt.Fprintln(out, "✅ Файл сохранён в", savePath)
				}

			default:
				fmt.Fprintln(out, "🤷 Неизвестный тип данных")
			}

			return nil
		},
	}
}

func (g *GophKeeper) VaultDeleteCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [id]",
		Short: "Удалить запись по ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			id, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("неверный ID: %w", err)
			}
			_, err = g.VaultDelete(id)
			if err != nil {
				return fmt.Errorf("ошибка удаления: %w", err)
			}
			_, _ = fmt.Fprintln(out, "✅ Запись удалена.")
			return nil
		},
	}
}
