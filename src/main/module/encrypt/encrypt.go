package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"log"
)
//加密
func AesEncryptSimple(origData []byte, key string, iv string) (string, error) {
	return AesEncryptPkcs5(origData, []byte(key), []byte(iv))
}

func AesEncryptPkcs5(origData []byte, key []byte, iv []byte ) (string, error) {
	return AesEncrypt(origData, key, iv, PKCS5Padding)
}

func AesEncrypt(origData []byte, key []byte, iv []byte, paddingFunc func([]byte, int) []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return "", err
	}
	blockSize := block.BlockSize()
	origData = paddingFunc(origData, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

//解密
func AesDecryptSimple(crypted string, key string, iv string) ([]byte, error) {
	return AesDecryptPkcs5(crypted, []byte(key), []byte(iv))
}

func AesDecryptPkcs5(crypted string, key []byte, iv []byte) ([]byte, error) {
	return AesDecrypt(crypted, key, iv, PKCS5UnPadding)
}

func AesDecrypt(crypted string, key []byte, iv []byte, unPaddingFunc func([]byte) []byte) ([]byte, error) {
	crypted1,err := base64.StdEncoding.DecodeString(crypted)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(crypted1))
	blockMode.CryptBlocks(origData, crypted1)
	origData = unPaddingFunc(origData)
	return origData, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	if length < unpadding {
		return []byte("unpadding error")
	}
	return origData[:(length - unpadding)]
}

