package profile

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	km "github.com/Bit-Nation/panthalassa/keyManager"
	x3dh "github.com/Bit-Nation/x3dh"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	lp2pCrypto "github.com/libp2p/go-libp2p-crypto"
	ed25519 "golang.org/x/crypto/ed25519"
)

const profileVersion = 1

type Information struct {
	Name           string         `json:"name"`
	Location       string         `json:"location"`
	Image          string         `json:"image"`
	IdentityPubKey string         `json:"identity_pub_key"`
	EthereumPubKey string         `json:"ethereum_pub_Key"`
	ChatIDKey      x3dh.PublicKey `json:"chat_id_key"`
	Timestamp      time.Time      `json:"timestamp"`
	Version        uint16         `json:"version"`
}

type Signatures struct {
	IdentityKey string `json:"identity_key"`
	EthereumKey string `json:"ethereum_key"`
}

type Profile struct {
	Information Information `json:"information"`
	Signatures  Signatures  `json:"signatures"`
}

func (p *Profile) Marshal() ([]byte, error) {

	return json.Marshal(p)

}

func (p *Profile) GetIdentityKey() (ed25519.PublicKey, error) {

	pubStr := p.Information.IdentityPubKey
	rawPub, err := hex.DecodeString(pubStr)

	if err != nil {
		return ed25519.PublicKey{}, err
	}

	if len(rawPub) != 32 {
		return ed25519.PublicKey{}, errors.New("public key must have 32 bytes")
	}

	return rawPub, nil

}

// fetch chat public key
func (p Profile) GetChatIDPublicKey() x3dh.PublicKey {
	return p.Information.ChatIDKey
}

// check if signatures of profile are correct
func (p Profile) SignaturesValid() (bool, error) {

	// unmarshal id pub key
	rawIdPubKey, err := hex.DecodeString(p.Information.IdentityPubKey)
	if err != nil {
		return false, err
	}

	// unmarshal eth pub key
	rawEthPubKey, err := hex.DecodeString(p.Information.EthereumPubKey)
	if err != nil {
		return false, err
	}

	// concat data to verify (all information data)
	dataToVerify := []byte(p.Information.Name)
	dataToVerify = append(dataToVerify, []byte(p.Information.Location)...)
	dataToVerify = append(dataToVerify, []byte(p.Information.Image)...)
	dataToVerify = append(dataToVerify, []byte(p.Information.Timestamp.String())...)
	dataToVerify = append(dataToVerify, rawIdPubKey...)
	dataToVerify = append(dataToVerify, rawEthPubKey...)
	dataToVerify = append(dataToVerify, p.Information.ChatIDKey[:]...)
	version := make([]byte, 2)
	binary.LittleEndian.PutUint16(version, p.Information.Version)
	dataToVerify = append(dataToVerify, version...)

	// hash to verify
	dataHash := sha256.Sum256(dataToVerify)

	// check data signature of id key
	idPubKey, err := lp2pCrypto.UnmarshalEd25519PublicKey(rawIdPubKey)
	if err != nil {
		return false, err
	}
	idSig, err := hex.DecodeString(p.Signatures.IdentityKey)
	if err != nil {
		return false, err
	}
	valid, err := idPubKey.Verify(dataHash[:], idSig)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, errors.New("identity public key signature is invalid")
	}

	// check data signature of eth key
	ethSig, err := hex.DecodeString(p.Signatures.EthereumKey)
	if err != nil {
		return false, err
	}
	// an the signature contain's [R || S || V] - but we only need R and S
	valid = ethCrypto.VerifySignature(rawEthPubKey, dataHash[:], ethSig[:64])
	if !valid {
		return false, errors.New("ethereum public key signature is invalid")
	}

	return true, nil

}

// parse json profile to Profile
func Unmarshal(jsonProf string) (Profile, error) {

	var profile Profile

	if err := json.Unmarshal([]byte(jsonProf), &profile); err != nil {
		return Profile{}, err
	}

	//@todo check if profile is valid (at least the field's)

	return profile, nil

}

// sign the metadata with identity and ethereum key
func SignProfile(name, location, image string, km km.KeyManager) (Profile, error) {

	// id public key
	idPubKeyStr, err := km.IdentityPublicKey()
	if err != nil {
		return Profile{}, err
	}
	idPubKey, err := hex.DecodeString(idPubKeyStr)
	if err != nil {
		return Profile{}, err
	}

	// ethereum public key
	ethPubKeyStr, err := km.GetEthereumPublicKey()
	if err != nil {
		return Profile{}, err
	}
	ethPubKey, err := hex.DecodeString(ethPubKeyStr)
	if err != nil {
		return Profile{}, err
	}

	// now
	now := time.Now().UTC()

	// chat id public key
	pair, err := km.ChatIdKeyPair()
	if err != nil {
		return Profile{}, err
	}
	chatPub := pair.PublicKey

	// concat profile information
	dataToSign := []byte(name)
	dataToSign = append(dataToSign, []byte(location)...)
	dataToSign = append(dataToSign, []byte(image)...)
	dataToSign = append(dataToSign, []byte(now.String())...)
	dataToSign = append(dataToSign, idPubKey...)
	dataToSign = append(dataToSign, ethPubKey...)
	dataToSign = append(dataToSign, chatPub[:]...)
	version := make([]byte, 2)
	binary.LittleEndian.PutUint16(version, profileVersion)
	dataToSign = append(dataToSign, version...)

	// hash of profile data
	dataHash := sha256.Sum256(dataToSign)

	// hash the profile data
	if err != nil {
		return Profile{}, err
	}

	// sign hash of data with id key
	idKeySignature, err := km.IdentitySign(dataHash[:])
	if err != nil {
		return Profile{}, err
	}

	// sign hash of data with eth key
	ethKeySignature, err := km.EthereumSign(dataHash)
	if err != nil {
		return Profile{}, err
	}

	// profile
	prof := Profile{
		Information: Information{
			Name:           name,
			Location:       location,
			Image:          image,
			IdentityPubKey: idPubKeyStr,
			EthereumPubKey: ethPubKeyStr,
			Timestamp:      now,
			ChatIDKey:      chatPub,
			Version:        profileVersion,
		},
		Signatures: Signatures{
			IdentityKey: hex.EncodeToString(idKeySignature),
			EthereumKey: hex.EncodeToString(ethKeySignature),
		},
	}

	return prof, nil

}

// sign's a profile without the need to start panthalassa
func SignWithKeyManagerStore(name, location, image string, keyManagerStore km.Store, password string) (Profile, error) {

	keyManager, err := km.OpenWithPassword(keyManagerStore, password)
	if err != nil {
		return Profile{}, err
	}

	return SignProfile(name, location, image, *keyManager)

}
