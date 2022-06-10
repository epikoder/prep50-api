package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"io/ioutil"
	randMath "math/rand"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
)

const alphabet = "./ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

var (
	bcEncoding = base64.NewEncoding(alphabet)
)

func KeyGen(generatePemfile bool) (k *ecdsa.PrivateKey, err error) {
	if k, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); !logger.HandleError(err) {
		return nil, err
	}
	{
		if generatePemfile {
			pemPriv, _, err := PemKeyPair(k)
			if err != nil {
				panic(err)
			}
			__PATH__, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			if err != nil {
				panic(err)
			}
			if err := ioutil.WriteFile(__PATH__+"/jwt.key", pemPriv, os.ModePerm); err != nil {
				panic(err)
			}
		}
	}
	return
}

func PemKeyPair(key *ecdsa.PrivateKey) (privateKeyPEM []byte, publicKeyPEM []byte, err error) {
	var der []byte
	if der, err = x509.MarshalECPrivateKey(key); !logger.HandleError(err) {
		return nil, nil, err
	}

	privateKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	})

	if der, err = x509.MarshalPKIXPublicKey(key.Public()); !logger.HandleError(err) {
		return nil, nil, err
	}

	publicKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "EC PUBLIC KEY",
		Bytes: der,
	})
	return
}

func Base64Encode(src string, i interface{}) string {
	n := bcEncoding.EncodedLen(len(src))
	dst := make([]byte, n)
	bcEncoding.Encode(dst, []byte(src))
	for dst[n-1] == '=' {
		n--
	}
	if i != nil && i.(int) <= len(dst) {
		return string(dst[:i.(int)])
	}
	return string(dst[:n])
}

func Base64Decode(s string, i interface{}) (string, error) {
	src := []byte(s)
	numOfEquals := 4 - (len(src) % 4)
	for i := 0; i < numOfEquals; i++ {
		src = append(src, '=')
	}

	dst := make([]byte, bcEncoding.DecodedLen(len(src)))
	n, err := bcEncoding.Decode(dst, []byte(src))
	if !logger.HandleError(err) {
		return "", err
	}
	return string(dst[:n]), nil
}

func Base64(n int) string {
	src := make([]byte, n)
	randMath.Seed(time.Now().UnixNano())
	rand.Read(src)
	dst := make([]byte, base64.URLEncoding.EncodedLen(len(src)))
	bcEncoding.Encode(dst, src)
	return string(dst[len(dst)-n:])
}

func Ase256Encode(plaintext string) (string, error) {
	bKey := []byte(config.Conf.Aes.Key)
	bIV := []byte(config.Conf.Aes.Iv)
	bPlaintext := PKCS5Padding([]byte(plaintext), aes.BlockSize, len(plaintext))
	block, err := aes.NewCipher(bKey)
	if err != nil {
		return "", err
	}
	ciphertext := make([]byte, len(bPlaintext))
	mode := cipher.NewCBCEncrypter(block, bIV)
	mode.CryptBlocks(ciphertext, bPlaintext)
	return hex.EncodeToString(ciphertext), nil
}

func Ase256Decode(cipherText string) (decryptedString string, err error) {
	bKey := []byte(config.Conf.Aes.Key)
	bIV := []byte(config.Conf.Aes.Iv)
	cipherTextDecoded, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(bKey)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, bIV)
	mode.CryptBlocks([]byte(cipherTextDecoded), []byte(cipherTextDecoded))
	return strings.TrimFunc(string(cipherTextDecoded), func(r rune) bool {
		return !unicode.IsGraphic(r) || !unicode.IsPrint(r)
	}), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
