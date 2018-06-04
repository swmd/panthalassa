package aes

import (
	"errors"
	"testing"

	require "github.com/stretchr/testify/require"
)

//Test the encrypt and decrypt function in one batch
func TestSuccessEncryptDecrypt(t *testing.T) {

	secret := Secret{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	value := "I am the value"

	//Encrypt
	cipherText, e := Encrypt(value, secret)
	require.Nil(t, e)

	//Decrypt
	res, err := Decrypt(cipherText, secret)
	require.Nil(t, err)

	//Decrypted value must match the given value
	require.Equal(t, value, res)

}

func TestFailedDecryption(t *testing.T) {

	secret := Secret{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	value := "I am the plain text"

	// encrypt
	cipherText, e := Encrypt(value, secret)
	require.Nil(t, e)

	// change last byte to fail on decryption
	secret[31] = 0x01

	// decrypt
	res, err := Decrypt(cipherText, secret)
	require.Equal(t, "", res)
	require.Error(t, errors.New("cipher: message authentication failed"), err)

}
