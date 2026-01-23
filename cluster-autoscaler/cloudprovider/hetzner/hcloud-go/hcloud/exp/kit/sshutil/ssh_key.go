package sshutil

import (
	"crypto"
	"crypto/ed25519"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// GenerateKeyPair generates a new ed25519 ssh key pair, and returns the private key and
// the public key respectively.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func GenerateKeyPair() ([]byte, []byte, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate key pair: %w", err)
	}

	privBytes, err := encodePrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("could not encode private key: %w", err)
	}

	pubBytes, err := encodePublicKey(pub)
	if err != nil {
		return nil, nil, fmt.Errorf("could not encode public key: %w", err)
	}

	return privBytes, pubBytes, nil
}

func encodePrivateKey(priv crypto.PrivateKey) ([]byte, error) {
	privPem, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(privPem), nil
}

func encodePublicKey(pub crypto.PublicKey) ([]byte, error) {
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, err
	}

	return ssh.MarshalAuthorizedKey(sshPub), nil
}

type privateKeyWithPublicKey interface {
	crypto.PrivateKey
	Public() crypto.PublicKey
}

// GeneratePublicKey generate a public key from the provided private key.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func GeneratePublicKey(privBytes []byte) ([]byte, error) {
	priv, err := ssh.ParseRawPrivateKey(privBytes)
	if err != nil {
		return nil, fmt.Errorf("could not decode private key: %w", err)
	}

	key, ok := priv.(privateKeyWithPublicKey)
	if !ok {
		return nil, fmt.Errorf("private key doesn't export Public() crypto.PublicKey")
	}

	pubBytes, err := encodePublicKey(key.Public())
	if err != nil {
		return nil, fmt.Errorf("could not encode public key: %w", err)
	}

	return pubBytes, nil
}

// GetPublicKeyFingerprint generate the finger print for the provided public key.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func GetPublicKeyFingerprint(pubBytes []byte) (string, error) {
	pub, _, _, _, err := ssh.ParseAuthorizedKey(pubBytes)
	if err != nil {
		return "", fmt.Errorf("could not decode public key: %w", err)
	}

	fingerprint := ssh.FingerprintLegacyMD5(pub)

	return fingerprint, nil
}
