package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	cryptorand "crypto/rand"
	"crypto/sha512"
	"io"

	"github.com/blend/go-sdk/exception"
)

// Crypto is a namespace for crypto related functions.
var Crypto = cryptoUtil{}

type cryptoUtil struct{}

// GCMEncryptionResult is a struct for a gcm encryption result
type GCMEncryptionResult struct {
	CipherText []byte
	Nonce      []byte
}

// CreateKey creates a key of a given size by reading that much data off the crypto/rand reader.
func (cu cryptoUtil) CreateKey(keySize int) ([]byte, error) {
	key := make([]byte, keySize)
	_, err := cryptorand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// MustCreateKey creates a key, if an error is returned, it panics.
func (cu cryptoUtil) MustCreateKey(keySize int) []byte {
	key, err := cu.CreateKey(keySize)
	if err != nil {
		panic(err)
	}
	return key
}

// Encrypt encrypts data with the given key.
func (cu cryptoUtil) Encrypt(key, plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plainText))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(cryptorand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], plainText)
	return ciphertext, nil
}

// GCMEncryptionInPlace will encrypt and authenticate the plaintext with the given key. GCMEncryptionResult.CipherText will be backed in memory by the inputted plaintext []byte. Use GCMEncryption.CipherText, never plaintext.
func (cu cryptoUtil) GCMEncryptInPlace(key, plainText []byte) (*GCMEncryptionResult, error) {
	return cu.GCMEncrypt(key, plainText, plainText[:0])
}

// GCMEncrypt encrypts and authenticates the plaintext with the given key. dst is the destination slice for the encrypted data
func (cu cryptoUtil) GCMEncrypt(key, plainText, dst []byte) (*GCMEncryptionResult, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, exception.New(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, exception.New(err)
	}
	nonce := make([]byte, aead.NonceSize())
	_, err = cryptorand.Read(nonce)
	if err != nil {
		return nil, exception.New(err)
	}
	dst = aead.Seal(dst, nonce, plainText, nil)
	return &GCMEncryptionResult{CipherText: dst, Nonce: nonce}, nil
}

// GCMDecryptInPlace decrypts the ciphertext in place. GCMEncryptionResult.CipherText will be used as the backing memory, but only use the returned slice
func (cu cryptoUtil) GCMDecryptInPlace(key []byte, gcm *GCMEncryptionResult) ([]byte, error) {
	return cu.GCMDecrypt(key, gcm, gcm.CipherText[:0])
}

// GCMDecrypt decrypts and authenticates the cipherText with the given key
func (cu cryptoUtil) GCMDecrypt(key []byte, gcm *GCMEncryptionResult, dst []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, exception.New(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, exception.New(err)
	}
	dst, err = aead.Open(dst, gcm.Nonce, gcm.CipherText, nil)
	return dst, exception.New(err)
}

// Decrypt decrypts data with the given key.
func (cu cryptoUtil) Decrypt(key, cipherText []byte) ([]byte, error) {
	if len(cipherText) < aes.BlockSize {
		return nil, exception.New("cannot decrypt string: `cipherText` is smaller than AES block size").WithMessagef("block size: %v", aes.BlockSize)
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)
	return cipherText, nil
}

// Hash hashes data with the given key.
func (cu cryptoUtil) Hash(key, plainText []byte) []byte {
	mac := hmac.New(sha512.New, key)
	mac.Write([]byte(plainText))
	return mac.Sum(nil)
}

// SecureRandomBytes generates a fixed length of random bytes.
func (cu cryptoUtil) SecureRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := cryptorand.Read(b)
	if err != nil {
		return nil, exception.New(err)
	}

	return b, nil
}

// MustSecureRandomBytes generates a random byte sequence and panics if it fails.
func (cu cryptoUtil) MustSecureRandomBytes(length int) []byte {
	b, err := cu.SecureRandomBytes(length)
	if err != nil {
		panic(err)
	}
	return b
}
