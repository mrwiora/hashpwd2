package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"strings"
	"time"

	"golang.org/x/crypto/argon2"

	"golang.design/x/clipboard"
	"golang.org/x/term"
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	keyLength   uint32
}

func main() {
	// Establish the parameters to use for Argon2.
	p := &params{
		memory:      1 * 1024 * 1024,
		iterations:  16,
		parallelism: 4,
		keyLength:   64,
	}

	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	fmt.Println("Enter secret: ")
	textSecret, _ := term.ReadPassword(0)

	fmt.Println("Enter salt: ")
	textSecretSalt, _ := term.ReadPassword(0)

	// Removing end of line
	textSecretCleaned := strings.Replace(string(textSecret[:]), "\n", "", -1)
	textSecretSaltCleaned := strings.Replace(string(textSecretSalt[:]), "\n", "", -1)

	fmt.Println("Please wait!")

	// Pass the plaintext password and parameters to our generateFromPassword
	// helper function.
	hash, err := generateFromPassword(textSecretCleaned, textSecretSaltCleaned, p)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(hash)

	clipboard.Write(clipboard.FmtText, []byte(hash))
	fmt.Println("OK! Hurry up - you have 30 seconds to paste :)")
	time.Sleep(10 * time.Second)
	clipboard.Write(clipboard.FmtText, []byte("---"))

}

func generateFromPassword(password string, salt string, p *params) (encodedHash string, err error) {

	hash := argon2.IDKey([]byte(password), []byte(salt), p.iterations, p.memory, p.parallelism, p.keyLength)

	// Base64 encode the salt and hashed password.
	//b64Salt := base64.RawStdEncoding.EncodeToString([]byte(salt))
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	encodedHash = fmt.Sprintf("%s", b64Hash)

	return encodedHash, nil
}
