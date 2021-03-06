package chat

import (
	"errors"

	x3dh "github.com/Bit-Nation/x3dh"
	doubleratchet "github.com/tiabc/doubleratchet"
)

// local double ratchet key pair
// as a local helper
type doubleratchetKeyPair struct {
	kp x3dh.KeyPair
}

func (p doubleratchetKeyPair) PrivateKey() doubleratchet.Key {
	var byt [32]byte = p.kp.PrivateKey
	return byt
}
func (p doubleratchetKeyPair) PublicKey() doubleratchet.Key {
	var byt [32]byte = p.kp.PublicKey
	return byt
}

// encrypt a double rachet message
func (c *Chat) encryptMessage(secret x3dh.SharedSecret, data []byte) (doubleratchet.Message, error) {

	// fetch chat ID key pair from key manager
	chatIdKey, err := c.km.ChatIdKeyPair()
	if err != nil {
		return doubleratchet.Message{}, err
	}

	// decoded secret
	var secBytes [32]byte = secret

	// create double rachet session
	s, err := doubleratchet.New(secBytes, doubleratchetKeyPair{
		kp: chatIdKey,
	}, doubleratchet.WithKeysStorage(c.doubleRachetKeyStore))
	if err != nil {
		return doubleratchet.Message{}, err
	}

	// encrypt message
	return s.RatchetEncrypt(data, []byte{}), nil

}

// decrypt a message
func (c *Chat) DecryptMessage(secret x3dh.SharedSecret, msg Message) (string, error) {

	valid, err := msg.VerifySignature()
	if err != nil {
		return "", err
	}
	if !valid {
		return "", errors.New("failed to verify message signature")
	}

	// chat partner chat id public key
	chatIdKey := msg.DoubleratchetMessage.Header.DH

	var secBytes [32]byte = secret
	var remotePub [32]byte = chatIdKey

	// create double rachet instance
	s, err := doubleratchet.NewWithRemoteKey(
		secBytes,
		remotePub,
		doubleratchet.WithKeysStorage(c.doubleRachetKeyStore),
	)
	if err != nil {
		return "", err
	}

	// decrypt
	dec, err := s.RatchetDecrypt(msg.DoubleratchetMessage, nil)

	return string(dec), err
}
