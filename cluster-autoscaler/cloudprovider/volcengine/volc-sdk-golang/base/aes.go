/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package base

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
)

// AES CBC
func aesEncryptCBC(origData, key []byte) (crypted []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			crypted = nil
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	blockSize := block.BlockSize()
	origData = zeroPadding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted = make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return
}

// AES CBC Do a Base64 encryption after encryption
func aesEncryptCBCWithBase64(origData, key []byte) (string, error) {
	cbc, err := aesEncryptCBC(origData, key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(cbc), nil
}

func zeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	if padding == 0 {
		return ciphertext
	}

	padtext := bytes.Repeat([]byte{byte(0)}, padding)
	return append(ciphertext, padtext...)
}
