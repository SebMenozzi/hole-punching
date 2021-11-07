package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"io"

	"golang.org/x/crypto/curve25519"
)

// Private key generation with Curve25519 Diffie-Hellman function
// http://cr.yp.to/ecdh.html
func GenKeyPair() ([32]byte, [32]byte, error) {
	var pri [32]byte
	var pub [32]byte

	_, err := rand.Read(pri[:])
	if err != nil {
		return pri, pub, err
	}
	pri[0] &= 248
	pri[31] &= 127
	pri[31] |= 64

	curve25519.ScalarBaseMult(&pub, &pri)
	if err != nil {
		return pri, pub, err
	}

	return pri, pub, nil
}

// Shared secret generation with Curve25519 Diffie-Hellman function
// http://cr.yp.to/ecdh.html
func GenSharedSecret(selfPrivate [32]byte, otherPublic [32]byte) [32]byte {
	var secret [32]byte
	curve25519.ScalarMult(&secret, &selfPrivate, &otherPublic)
	return secret
}

// HMAC + SHA2 hash function
func Hash(tag string, data []byte) []byte {
	h := hmac.New(sha512.New512_256, []byte(tag))
	h.Write(data)
	return h.Sum(nil)
}

// https://github.com/gtank/cryptopasta/blob/master/encrypt.go
func Encrypt(plaintext []byte, secret [32]byte) ([]byte, error) {
	block, err := aes.NewCipher(secret[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// https://github.com/gtank/cryptopasta/blob/master/encrypt.go
func Decrypt(ciphertext []byte, secret [32]byte) ([]byte, error) {
	block, err := aes.NewCipher(secret[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}
