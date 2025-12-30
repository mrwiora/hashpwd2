package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"strings"
	"time"

	"golang.org/x/crypto/argon2"

	"golang.design/x/clipboard"
	"golang.org/x/term"
)

var (
	// Version information, set via ldflags at build time
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	keyLength   uint32
}

func main() {
	// Check for version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Fprintf(os.Stderr, "hashpwd2 version %s\n", version)
		fmt.Fprintf(os.Stderr, "commit: %s\n", commit)
		fmt.Fprintf(os.Stderr, "built: %s\n", date)
		os.Exit(0)
	}
	// Establish the parameters to use for Argon2.
	p := &params{
		memory:      1 * 1024 * 1024,
		iterations:  16,
		parallelism: 4,
		keyLength:   64,
	}

	// Check if output is being piped
	stat, _ := os.Stdout.Stat()
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	// Only initialize clipboard if not piped
	if !isPiped {
		err := clipboard.Init()
		if err != nil {
			panic(err)
		}
	}

	// Write informational messages to stderr so they don't get piped
	fmt.Fprintln(os.Stderr, "Enter secret: ")
	textSecret, _ := term.ReadPassword(int(os.Stdin.Fd()))

	fmt.Fprintln(os.Stderr, "Enter salt: ")
	textSecretSalt, _ := term.ReadPassword(int(os.Stdin.Fd()))

	// Removing end of line
	textSecretCleaned := strings.Replace(string(textSecret[:]), "\n", "", -1)
	textSecretSaltCleaned := strings.Replace(string(textSecretSalt[:]), "\n", "", -1)

	fmt.Fprintln(os.Stderr, "Please wait!")

	// Pass the plaintext password and parameters to our generateFromPassword
	// helper function.
	hash, err := generateFromPassword(textSecretCleaned, textSecretSaltCleaned, p)
	if err != nil {
		log.Fatal(err)
	}

	if isPiped {
		// When piped, only output the hash to stdout
		fmt.Println(hash)
	} else {
		// When not piped, use clipboard as before
		fmt.Println(hash)
		clipboard.Write(clipboard.FmtText, []byte(hash))
		fmt.Println("OK! Hurry up - you have 30 seconds to paste :)")
		time.Sleep(10 * time.Second)
		clipboard.Write(clipboard.FmtText, []byte("---"))
	}

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
