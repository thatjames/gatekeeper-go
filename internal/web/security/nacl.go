package security

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/nacl/box"
)

var (
	privKey, pubKey *[32]byte
)

type AuthToken struct {
	Expires  time.Time
	Username string
}

func init() {
	var err error
	if pubKey, privKey, err = box.GenerateKey(rand.Reader); err != nil {
		panic(err)
	}
}

func CreateAuthToken(username string) (string, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return "", err
	}

	msg, _ := json.Marshal(AuthToken{time.Now().Add(time.Hour), username})
	recipientPublicKey, recipientPrivateKey, err := box.GenerateKey(rand.Reader) //create a unique keypair for this request
	if err != nil {
		return "", err
	}

	encrypted := box.Seal(nonce[:], msg, &nonce, recipientPublicKey, privKey)
	return fmt.Sprintf("%x", append(recipientPrivateKey[:], encrypted...)), nil //we will need the private key to decrypt the request when it returns
}

func ParseClaim(claim string) (*AuthToken, error) {
	encrypted, err := hex.DecodeString(claim)
	if err != nil {
		return nil, err
	}

	//First pull the private key, the first 32 bytes of the message
	var recipientPrivateKey [32]byte
	copy(recipientPrivateKey[:], encrypted[:32])
	encrypted = encrypted[32:]

	//Then pull the nonce, the next 24 bytes
	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])

	//Whatever remains in the buffer is our encrypted payload, decrypt it now
	decrypted, ok := box.Open(nil, encrypted[24:], &decryptNonce, pubKey, &recipientPrivateKey)
	if !ok {
		return nil, errors.New("failed")
	}

	var token AuthToken
	if err := json.Unmarshal(decrypted, &token); err != nil {
		return nil, err
	}

	return &token, nil
}
