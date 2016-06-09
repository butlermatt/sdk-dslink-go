package crypto

import (
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"math/big"
	"strings"
	"fmt"
)

// ECDH manages creating and providing keys.
type ECDH interface {
	GenerateKey(io.Reader) (PrivateKey, error)
	Marshal(PrivateKey) (string, error)
	Unmarshal(string) (PrivateKey, error)
	UnmarshalPublic(string) (PublicKey, error)
	GenerateSharedSecret(PrivateKey, PublicKey) ([]byte, error)
}

// NewECDH returns a new Elliptic ECDH
func NewECDH() ECDH {
	return &ellipticECDH{elliptic.P256(), base64.RawURLEncoding}
}

type ellipticECDH struct {
	ECDH
	Curve elliptic.Curve
	base  base64.Encoding
}

type PublicKey struct {
	Curve elliptic.Curve
	X, Y  *big.Int
}

func (p PublicKey) marshal() []byte {
	return elliptic.Marshal(p.Curve, p.X, p.Y)
}

func (p PublicKey) Base64() string {
	return base64.RawURLEncoding.EncodeToString(p.marshal())
}

func (p PublicKey) Hash64() string {
	return base64.RawURLEncoding.EncodeToString(sha256.Sum256(p.marshal()))
}

type PrivateKey struct {
	PublicKey
	D []byte
}

// GenerateKey will create a new Private/Public key pair based on random numbers from io.Reader.
// Returns error if it was unable to create keys.
func (e *ellipticECDH) GenerateKey(rand io.Reader) (PrivateKey, error) {
	var priv PrivateKey

	d, x, y, err := elliptic.GenerateKey(e.Curve, rand)
	if err != nil {
		return priv, err
	}

	return PrivateKey{PublicKey{Curve: e.Curve, X: x, Y: y}, D: d}, nil
}

// Marshal converts a Private/Public key pair into a Base64.RawUrlEncoded string.
// Returned string separates the pairs with a space, where private key is first.
// Returns an error if unable to convert values to a string.
func (e *ellipticECDH) Marshal(priv PrivateKey) (string, error) {
	pd := e.base.EncodeToString(priv.D)
	pm := priv.PublicKey.Base64()
	return fmt.Sprintf("%s %s", pd, pm)
}

// Unmarshal will decode a Base64.RawUrlEncoded string into a Private/Public key pair.
// String may be Private / Public keys separated by a space or alternatively a private
// key and the public will be generated automatically.
// Returns an error if string cannot be decoded.
func (e *ellipticECDH) Unmarshal(str string) (PrivateKey, error) {
	keys := strings.Split(str, " ")

	var priv PrivateKey

	d, err := e.base.DecodeString(keys[0])
	if err != nil {
		return priv, err
	}

	switch len(keys) {
	case 2:
		pub, err := e.UnmarshalPublic(keys[1])
		if err != nil {
			return priv, err
		}
		priv = PrivateKey{PublicKey{Curve: pub.Curve, X: pub.X, Y: pub.Y}, D: d}
		return priv, nil
	case 1:
		x, y := e.Curve.ScalarBaseMult(d)
		priv = PrivateKey{PublicKey{Curve: e.Curve, X: x, Y: y}, D: d}
		return priv, nil
	default:
		return priv, errors.New("too many sections to unmarshal.")
	}

}

// UnmarshalPublic will decode a Base64.RawUrlEncoded string into a Public key.
// Returns an error if string cannot be decoded.
func (e *ellipticECDH) UnmarshalPublic(str string) (PublicKey, error) {
	var pub PublicKey

	data, err := e.base.DecodeString(str)
	if err != nil {
		return pub, err
	}

	x, y := elliptic.Unmarshal(e.Curve, data)
	if x == nil || y == nil {
		return pub, errors.New("unmashaled values are nil")
	}

	pub = PublicKey{Curve: e.Curve, X: x, Y: y}
	return pub, nil
}
