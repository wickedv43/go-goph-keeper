package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/dongri/go-mnemonic"
	"github.com/pkg/errors"
)

func GenerateMnemonic() (string, error) {
	words, err := mnemonic.GenerateMnemonic(128, mnemonic.LanguageEnglish)
	if err != nil {
		return "", err
	}

	return words, nil
}

func GenerateSeed(words, password string) string {
	seed := mnemonic.ToSeedHex(words, password)

	return seed
}

// EncryptAES128 шифрует данные с 128-битным ключом (16 байт)
func EncryptAES128(data []byte, key []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("ключ должен быть 16 байт (128 бит)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12) // 12 байт — рекомендуемый размер для GCM
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, data, nil)
	return append(nonce, ciphertext...), nil
}

// DecryptAES128 расшифровывает данные, зашифрованные EncryptAES128
func DecryptAES128(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("ключ должен быть 16 байт (128 бит)")
	}
	if len(ciphertext) < 12 {
		return nil, errors.New("слишком короткий ciphertext")
	}

	nonce := ciphertext[:12]
	encrypted := ciphertext[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Open(nil, nonce, encrypted, nil)
}
