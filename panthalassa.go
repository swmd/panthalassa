package panthalassa

import (
	"encoding/hex"

	"fmt"
	api "github.com/Bit-Nation/panthalassa/api/device"
	deviceApi "github.com/Bit-Nation/panthalassa/api/device"
	chat "github.com/Bit-Nation/panthalassa/chat"
	keyManager "github.com/Bit-Nation/panthalassa/keyManager"
	mesh "github.com/Bit-Nation/panthalassa/mesh"
	lp2pCrypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

type Panthalassa struct {
	km        *keyManager.KeyManager
	upStream  api.UpStream
	deviceApi *deviceApi.Api
	mesh      *mesh.Network
	chat      *chat.Chat
}

//Stop the panthalassa instance
//this becomes interesting when we start
//to use the mesh network
func (p *Panthalassa) Stop() error {
	return p.mesh.Close()
}

//Export account with the given password
func (p *Panthalassa) Export(pw, pwConfirm string) (string, error) {

	// export
	store, err := p.km.Export(pw, pwConfirm)
	if err != nil {
		return "", err
	}

	// marshal key store
	rawStore, err := store.Marshal()
	if err != nil {
		return "", err
	}

	return string(rawStore), nil

}

// add friend to peer store
func (p *Panthalassa) AddContact(pubKey string) error {

	// decode public key
	rawPubKey, err := hex.DecodeString(pubKey)
	if err != nil {
		return err
	}

	// create lp2p public key
	lp2pPubKey, err := lp2pCrypto.UnmarshalEd25519PublicKey(rawPubKey)
	if err != nil {
		return err
	}

	// create ID from friend public key
	id, err := peer.IDFromPublicKey(lp2pPubKey)
	if err != nil {
		return err
	}

	// add public key to peer store
	err = p.mesh.Host.Peerstore().AddPubKey(id, lp2pPubKey)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("added contact: %s", pubKey))

	return p.mesh.Host.Peerstore().Put(id, "contact", true)

}
