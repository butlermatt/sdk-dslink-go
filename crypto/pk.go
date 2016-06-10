package crypto

import (
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
)

// ECDH manages creating and providing keys.
type ECDH interface {
	GenerateKey(io.Reader) (PrivateKey, error)
	Marshal(PrivateKey) (string, error)
	Unmarshal(string) (PrivateKey, error)
	UnmarshalPublic(string) (PublicKey, error)
	GenerateSharedSecret(PrivateKey, PublicKey) []byte
	HashSalt(string, []byte) string
}

// NewECDH returns a new Elliptic ECDH
func NewECDH() ECDH {
	return &ellipticECDH{Curve: elliptic.P256(), base: base64.RawURLEncoding}
}

type ellipticECDH struct {
	ECDH
	Curve elliptic.Curve
	base  *base64.Encoding
}

type PublicKey struct {
	Curve elliptic.Curve
	X, Y  *big.Int
}

func (p PublicKey) marshal() []byte {
	return elliptic.Marshal(p.Curve, p.X, p.Y)
}

// Base64 returns a Base64 Raw Url Encoded (no padding) string of the bytes for the
// public key.
func (p PublicKey) Base64() string {
	return base64.RawURLEncoding.EncodeToString(p.marshal())
}

// Hash64 returns the SHA256 check sum of the bytes for the public key encoded
// as Base64 Raw Url encoded (no padding) string.
func (p PublicKey) Hash64() string {
	s := sha256.Sum256(p.marshal())
	return base64.RawURLEncoding.EncodeToString(s[:])
}

// DsId generates the dsId for this Public Key based on the prefix supplied.
// DsId returned should be the prefix
func (p PublicKey) DsId(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, p.Hash64())
}

// VerifyDsId confirms that the provided dsid matches the expected Hash64 for this
// public key.
func (p PublicKey) VerifyDsId(dsid string) bool {
	return strings.HasSuffix(dsid, p.Hash64())
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

	return PrivateKey{PublicKey: PublicKey{Curve: e.Curve, X: x, Y: y}, D: d}, nil
}

// Marshal converts a Private/Public key pair into a Base64.RawUrlEncoded string.
// Returned string separates the pairs with a space, where private key is first.
// Returns an error if unable to convert values to a string.
func (e *ellipticECDH) Marshal(priv PrivateKey) (string, error) {
	pd := e.base.EncodeToString(priv.D)
	pm := priv.PublicKey.Base64()
	return fmt.Sprintf("%s %s", pd, pm), nil
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
		priv = PrivateKey{PublicKey: PublicKey{Curve: pub.Curve, X: pub.X, Y: pub.Y}, D: d}
		return priv, nil
	case 1:
		x, y := e.Curve.ScalarBaseMult(d)
		priv = PrivateKey{PublicKey: PublicKey{Curve: e.Curve, X: x, Y: y}, D: d}
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

// GenerateSharedSecret Creates a shared secret based on the private of one and
// public key of the other. Returns a byte slice.
func (e *ellipticECDH) GenerateSharedSecret(priv PrivateKey, pub PublicKey) []byte {

	x, _ := e.Curve.ScalarMult(pub.X, pub.Y, priv.D)
	return x.Bytes() // RFC5903 states we should only return X.
}

// HashSalt adds the provided string salt to the SharedSecret byte slice sec.
// It returns a Base64 RawUrl encoded string of the SHA256 Sum of bytes.
func (e *ellipticECDH) HashSalt(salt string, sec []byte) string {
	raw := append([]byte(salt), sec...)
	s := sha256.Sum256(raw)
	return e.base.EncodeToString(s[:])
}
