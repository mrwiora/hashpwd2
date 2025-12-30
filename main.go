package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
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
	// Check for flags
	debugMode := false
	if len(os.Args) > 1 {
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			fmt.Fprintf(os.Stderr, "hashpwd2 version %s\n", version)
			fmt.Fprintf(os.Stderr, "commit: %s\n", commit)
			fmt.Fprintf(os.Stderr, "built: %s\n", date)
			os.Exit(0)
		}
		if os.Args[1] == "--debug" || os.Args[1] == "-d" {
			debugMode = true
		}
	}
	// Establish the parameters to use for Argon2.
	p := &params{
		memory:      1 * 1024 * 1024,
		iterations:  16,
		parallelism: 4,
		keyLength:   64,
	}

	// Check if stdout is being piped (for clipboard vs pipe output)
	stdoutStat, _ := os.Stdout.Stat()
	isStdoutPiped := (stdoutStat.Mode() & os.ModeCharDevice) == 0

	// Only initialize clipboard if stdout is not piped
	if !isStdoutPiped {
		err := clipboard.Init()
		if err != nil {
			panic(err)
		}
	}

	// Check if stdin is a terminal or pipe
	stdinStat, _ := os.Stdin.Stat()
	isStdinPiped := (stdinStat.Mode() & os.ModeCharDevice) == 0

	var textSecret, textSecretSalt []byte

	if isStdinPiped {
		// When stdin is piped, read directly from stdin
		fmt.Fprintln(os.Stderr, "Enter secret: ")
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		textSecret = []byte(password)

		fmt.Fprintln(os.Stderr, "Enter salt: ")
		salt, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		textSecretSalt = []byte(salt)
	} else {
		// When stdin is a terminal, use ReadPassword for security
		fmt.Fprintln(os.Stderr, "Enter secret: ")
		textSecret, _ = term.ReadPassword(int(os.Stdin.Fd()))

		fmt.Fprintln(os.Stderr, "Enter salt: ")
		textSecretSalt, _ = term.ReadPassword(int(os.Stdin.Fd()))
	}

	// Removing end of line
	textSecretCleaned := strings.Replace(string(textSecret[:]), "\n", "", -1)
	textSecretSaltCleaned := strings.Replace(string(textSecretSalt[:]), "\n", "", -1)

	// Debug output
	if debugMode {
		fmt.Fprintf(os.Stderr, "DEBUG: Password length: %d\n", len(textSecretCleaned))
		fmt.Fprintf(os.Stderr, "DEBUG: Password (hex): %x\n", textSecretCleaned)
		fmt.Fprintf(os.Stderr, "DEBUG: Salt length: %d\n", len(textSecretSaltCleaned))
		fmt.Fprintf(os.Stderr, "DEBUG: Salt (hex): %x\n", textSecretSaltCleaned)
	}

	fmt.Fprintln(os.Stderr, "Please wait!")

	// Pass the plaintext password and parameters to our generateFromPassword
	// helper function.
	hash, err := generateFromPassword(textSecretCleaned, textSecretSaltCleaned, p)
	if err != nil {
		log.Fatal(err)
	}

	if isStdoutPiped {
		// When stdout is piped, only output the hash to stdout (no clipboard)
		fmt.Println(hash)
	} else {
		// When stdout is not piped, use clipboard as before
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
