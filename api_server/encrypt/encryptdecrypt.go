// this file contains the part to encrypt and decrypt data.
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
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
		// fmt.Println("error when reading file: ", err.Error())
		return nil
	}
	// logger1.logTrace("TRACE", "Successfully loaded key from file")
	return decrypt(data1, loc)
}

func InsertSort(arr []float32, arrSort []int) ([]float32, []int) {
	len1 := len(arr)
	for i := 1; i < len1; i++ {
		temp1 := arr[i]
		tempSort := arrSort[i]
		i2 := i
		for i2 > 0 && arr[i2-1] > temp1 {
			arr[i2] = arr[i2-1]
			arrSort[i2] = arrSort[i2-1]
			i2--
		}
		arr[i2] = temp1
		arrSort[i2] = tempSort
	}
	// fmt.Println(arr, arrSort)
	return arr, arrSort
}

func MergeSort(arr []float32, arrSort []int) ([]float32, []int) {
	len1 := int(len(arr))
	len2 := int(len1 / 2)
	if len1 <= 5 {
		return InsertSort(arr, arrSort)
	} else {
		arr1, arrSort1 := MergeSort(arr[len2:], arrSort[len2:])
		arr2, arrSort2 := MergeSort(arr[:len2], arrSort[:len2])
		tempArr := make([]float32, len1)
		tempArrSort := make([]int, len1)
		i := 0
		for len(arr1) > 0 && len(arr2) > 0 {
			if arr1[0] < arr2[0] {
				tempArr[i] = arr1[0]
				tempArrSort[i] = arrSort1[0]
				arr1 = arr1[1:]
				arrSort1 = arrSort1[1:]
			} else {
				tempArr[i] = arr2[0]
				tempArrSort[i] = arrSort2[0]
				arr2 = arr2[1:]
				arrSort2 = arrSort2[1:]
			}
			i++
		}
		if len(arr1) == 0 {
			for j := 0; j < len(arr2); j++ {
				// fmt.Println(j, len(arr2), arr2, arr1, tempArr)
				tempArr[i] = arr2[j]
				tempArrSort[i] = arrSort2[j]
				i++
			}
		} else {
			for j := 0; j < len(arr1); j++ {
				tempArr[i] = arr1[j]
				tempArrSort[i] = arrSort1[j]
				i++
			}
		}
		return tempArr, tempArrSort
	}
}
