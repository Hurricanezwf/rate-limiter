package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const (
	TypeXORBase64 byte = 0x01
	TypeCBC       byte = 0x02
)

func Encrypt(key, toEncrypt []byte, encType byte) ([]byte, error) {
	switch encType {
	case TypeXORBase64:
		return EncryptWithXORBase64(key, toEncrypt)
	case TypeCBC:
		return EncryptWithCBC(key, toEncrypt)
	}
	return nil, errors.New("No Encrypt method found")
}

func Decrypt(key, toDecrypt []byte) ([]byte, error) {
	t := toDecrypt[len(toDecrypt)-1]
	switch t {
	case TypeXORBase64:
		return DecryptWithXORBase64(key, toDecrypt[:len(toDecrypt)-1])
	case TypeCBC:
		return DecryptWithCBC(key, toDecrypt[:len(toDecrypt)-1])
	}
	return nil, errors.New("No Decrypt method found")
}

// CBC
func EncryptWithCBC(key, src []byte) ([]byte, error) {
	// add padding
	toEncrypt := make([]byte, 0, len(src))
	toEncrypt = append(toEncrypt, src...)
	toEncrypt = PKCS7Padding(toEncrypt, aes.BlockSize)
	if len(toEncrypt)%aes.BlockSize != 0 {
		return nil, errors.New("Content to encrypt is not a multiple of the block size")
	}

	k := make([]byte, 0, 32)
	k = append(k, key...)
	block, err := aes.NewCipher(makeKey(k, 32))
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(toEncrypt)+1)
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], toEncrypt)
	ciphertext[len(ciphertext)-1] = TypeCBC
	return ciphertext, nil
}

func DecryptWithCBC(key, src []byte) ([]byte, error) {
	if len(src) < aes.BlockSize {
		return nil, errors.New("Content to decrypt to short")
	}

	k := make([]byte, 0, 32)
	k = append(k, key...)
	block, err := aes.NewCipher(makeKey(k, 32))
	if err != nil {
		return nil, err
	}

	toDecrypt := make([]byte, 0, len(src))
	toDecrypt = append(toDecrypt, src...)
	iv := toDecrypt[:aes.BlockSize]
	toDecrypt = toDecrypt[aes.BlockSize:]
	if len(toDecrypt)%aes.BlockSize != 0 {
		return nil, errors.New("Content to decrypt is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(toDecrypt, toDecrypt)
	toDecrypt = PKCS7UnPadding(toDecrypt)
	return toDecrypt, nil
}

// XORBase64
func EncryptWithXORBase64(key, src []byte) ([]byte, error) {
	k := make([]byte, 0, len(src))
	tmpSrc := make([]byte, len(src))
	k = append(k, key...)
	k = makeKey(k, len(src))
	for i, b := range src {
		tmpSrc[i] = k[i] ^ b
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(tmpSrc))+1)
	base64.StdEncoding.Encode(dst[:len(dst)-1], tmpSrc)
	dst[len(dst)-1] = TypeXORBase64
	return dst, nil
}

func DecryptWithXORBase64(key, src []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, err := base64.StdEncoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	dst = dst[:n]
	k := make([]byte, 0, len(dst))
	k = append(k, key...)
	k = makeKey(k, len(dst))
	for i, b := range dst {
		dst[i] = k[i] ^ b
	}
	return dst, nil
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func makeKey(key []byte, size int) []byte {
	for len(key) < size {
		key = append(key, key...)
	}
	return key[0:size]
}
