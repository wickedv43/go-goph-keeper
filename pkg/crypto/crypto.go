// Package crypto provides encryption, decryption, and mnemonic-based key generation for secure local storage.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/dongri/go-mnemonic"
	"github.com/pkg/errors"
)

// GenerateMnemonic returns a new 12-word English mnemonic phrase for seed generation.
func GenerateMnemonic() (string, error) {
	words, err := mnemonic.GenerateMnemonic(128, mnemonic.LanguageEnglish)
	if err != nil {
		return "", err
	}

	return words, nil
}

// GenerateSeed generates a hexadecimal seed string from a mnemonic phrase and password.
func GenerateSeed(words, password string) string {
	seed := mnemonic.ToSeedHex(words, password)

	return seed
}

// EncryptWithSeed encrypts the given data using AES-GCM with a key derived from the hex seed.
func EncryptWithSeed(data []byte, seedHex string) ([]byte, error) {
	seedBytes, err := hex.DecodeString(seedHex)
	if err != nil {
		return nil, err
	}
	if len(seedBytes) < 16 {
		return nil, errors.New("seed слишком короткий для AES-128")
	}
	key := seedBytes[:16]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12) // 12 байт для GCM
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

// DecryptWithSeed decrypts the given ciphertext using AES-GCM with a key derived from the hex seed.
func DecryptWithSeed(ciphertext []byte, seedHex string) ([]byte, error) {
	seedBytes, err := hex.DecodeString(seedHex)
	if err != nil {
		return nil, err
	}
	if len(seedBytes) < 16 {
		return nil, errors.New("seed слишком короткий для AES-128")
	}
	key := seedBytes[:16]

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
