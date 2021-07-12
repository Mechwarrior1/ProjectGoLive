// this file contains the part to encrypt and decrypt data.
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

// takes a passphrase and generate a new md5 key, for use in encrpyting and decrpyting.
func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

// opens a xml file containing the passphrase, to recreate the key for encrpytion/decryption.
// loc is the location "secure/keys.xml"
func getPassphrase(loc string) string {
	type Keys struct {
		Passphrase string
	}

	xmlFile, _ := os.Open("secure/keys.xml")
	defer xmlFile.Close()

	var keys Keys
	byteValue, _ := ioutil.ReadAll(xmlFile)
	xml.Unmarshal(byteValue, &keys)
	return keys.Passphrase
}

// takes in data bytes[] and encrypt it with the key from from createHash() and getPassphrase().
// returns the encrypted data as []bytes.
func encrypt(data []byte, loc string) []byte {
	passphrase := getPassphrase(loc)
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	ciphertext := gcm.Seal(nil, nonce, data, nil)
	return ciphertext
}

// takes in data bytes[] and decrypt it with the key from createHash() and getPassphrase().
// returns the decrypted data as []bytes.
func decrypt(data []byte, loc string) []byte {
	passphrase := getPassphrase(loc)
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	if err != nil {
		panic(err.Error())
	}
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

// takes the encrypted data from encrypt() and save it to a file with the filename as name.
// filename - string, to name the new saved encrypted file.
// data - the encrypted data.
func EncryptToFile(filename string, string1 string, loc string) {
	data := []byte(string1)
	f, _ := os.Create(filename)
	defer f.Close()
	f.Write(encrypt(data, loc))
	// logger1.logTrace("TRACE", "Successfully saved mapUser to file")
}

// takes in a filename, reads it and decode it with decrypt
func DecryptFromFile(filename string, loc string) []byte {
	data1, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("error when reading file")
		return nil
	}
	// logger1.logTrace("TRACE", "Successfully loaded key from file")
	return decrypt(data1, loc)
}
