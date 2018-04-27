package keyManager

import (
	keyStore "github.com/Bit-Nation/panthalassa/keyStore"
	"github.com/stretchr/testify/require"
	"testing"
)

//Test the Create from function
func TestCreateFromKeyStore(t *testing.T) {
	//create keyStore
	ks, err := keyStore.NewKeyStoreFactory()
	require.Nil(t, err)

	km := CreateFromKeyStore(ks)

	require.Equal(t, km.keyStore, ks)
}

func TestExportFunction(t *testing.T) {

	//create key storage
	jsonKeyStore := `{"mnemonic":"abandon amount liar amount expire adjust cage candy arch gather drum buyer","keys":{"eth_private_key":"dedbc9eb2b7eea18727f4b2e2d440b93e597cb283f00a3245943481785944d75"},"version":1}`
	ks, err := keyStore.FromJson(jsonKeyStore)
	require.Nil(t, err)

	//create key manager
	km := CreateFromKeyStore(ks)

	//Export the key storage via the key manager
	//The export should be encrypted
	cipherText, err := km.Export("my_password", "my_password")
	require.Nil(t, err)

	//Decrypt the exported encrypted key storage
	km, err = OpenWithPassword(cipherText, "my_password")
	require.Nil(t, err)

	jsonKs, err := km.keyStore.Marshal()
	require.Nil(t, err)

	require.Equal(t, jsonKeyStore, string(jsonKs))
}

func TestOpenWithMnemonic(t *testing.T) {

	//create key storage
	jsonKeyStore := `{"mnemonic":"abandon amount liar amount expire adjust cage candy arch gather drum buyer","keys":{"eth_private_key":"dedbc9eb2b7eea18727f4b2e2d440b93e597cb283f00a3245943481785944d75"},"version":1}`
	ks, err := keyStore.FromJson(jsonKeyStore)
	require.Nil(t, err)

	//create key manager
	km := CreateFromKeyStore(ks)

	//Export the key storage via the key manager
	//The export should be encrypted
	cipherText, err := km.Export("my_password", "my_password")
	require.Nil(t, err)

	//Decrypt the exported encrypted key storage
	km, err = OpenWithMnemonic(cipherText, "abandon amount liar amount expire adjust cage candy arch gather drum buyer")
	require.Nil(t, err)

	jsonKs, err := km.keyStore.Marshal()
	require.Nil(t, err)

	require.Equal(t, jsonKeyStore, string(jsonKs))

}