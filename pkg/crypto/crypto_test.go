package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateMnemonic(t *testing.T) {
	mnemo, err := GenerateMnemonic()
	require.NoError(t, err)
	require.NotEmpty(t, mnemo)
	require.Len(t, splitWords(mnemo), 12)
}

func splitWords(m string) []string {
	var res []string
	w := ""
	for _, r := range m {
		if r == ' ' {
			if w != "" {
				res = append(res, w)
				w = ""
			}
			continue
		}
		w += string(r)
	}
	if w != "" {
		res = append(res, w)
	}
	return res
}

func TestEncryptWithSeed_Errors(t *testing.T) {
	_, err := EncryptWithSeed([]byte("data"), "zzzzzz")
	require.Error(t, err)

	shortSeed := hex.EncodeToString([]byte("short"))
	_, err = EncryptWithSeed([]byte("data"), shortSeed)
	require.EqualError(t, err, "seed слишком короткий для AES-128")
}

func TestDecryptWithSeed_Errors(t *testing.T) {
	_, err := DecryptWithSeed([]byte("bad"), "00ff")
	require.Error(t, err)

	shortSeed := hex.EncodeToString([]byte("short"))
	cipher, _ := EncryptWithSeed([]byte("data"), GenerateSeed(mustMnemonic(), ""))
	_, err = DecryptWithSeed(cipher, shortSeed)
	require.EqualError(t, err, "seed слишком короткий для AES-128")

	seed := GenerateSeed(mustMnemonic(), "")
	cipher, _ = EncryptWithSeed([]byte("ok"), seed)
	cipher[13] ^= 0xFF
	_, err = DecryptWithSeed(cipher, seed)
	require.Error(t, err)
}

func mustMnemonic() string {
	m, err := GenerateMnemonic()
	if err != nil {
		panic(err)
	}
	return m
}

func TestGenerateSeed(t *testing.T) {
	mnemo := mustMnemonic()
	seed := GenerateSeed(mnemo, "password")
	require.NotEmpty(t, seed)
}

func TestEncryptAndDecryptWithSeed(t *testing.T) {
	data := []byte("secret")
	mnemo := mustMnemonic()
	seed := GenerateSeed(mnemo, "pass")

	cipher, err := EncryptWithSeed(data, seed)
	require.NoError(t, err)

	plain, err := DecryptWithSeed(cipher, seed)
	require.NoError(t, err)
	require.Equal(t, data, plain)
}

func TestDecryptWithWrongSeed(t *testing.T) {
	data := []byte("data")
	m1 := mustMnemonic()
	m2 := mustMnemonic()
	s1 := GenerateSeed(m1, "")
	s2 := GenerateSeed(m2, "")

	cipher, _ := EncryptWithSeed(data, s1)
	_, err := DecryptWithSeed(cipher, s2)
	require.Error(t, err)
}

func TestDecryptCorruptedCiphertext(t *testing.T) {
	mnemo := mustMnemonic()
	seed := GenerateSeed(mnemo, "")

	_, err := DecryptWithSeed([]byte("short"), seed)
	require.Error(t, err)

	cipher, _ := EncryptWithSeed([]byte("ok"), seed)
	cipher[13] ^= 0xFF
	_, err = DecryptWithSeed(cipher, seed)
	require.Error(t, err)
}
