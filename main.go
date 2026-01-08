package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"strings"

	"golang.org/x/crypto/argon2"
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
	useAskpass := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--version", "-v":
			fmt.Fprintf(os.Stderr, "hashpwd2 version %s\n", version)
			fmt.Fprintf(os.Stderr, "commit: %s\n", commit)
			fmt.Fprintf(os.Stderr, "built: %s\n", date)
			os.Exit(0)
		case "--debug", "-d":
			debugMode = true
		case "--askpass", "-a":
			useAskpass = true
		}
	}
	// Establish the parameters to use for Argon2.
	p := &params{
		memory:      1 * 1024 * 1024,
		iterations:  16,
		parallelism: 4,
		keyLength:   64,
	}

	// Check if stdin is a terminal or pipe
	stdinStat, _ := os.Stdin.Stat()
	isStdinPiped := (stdinStat.Mode() & os.ModeCharDevice) == 0

	var textSecret, textSecretSalt []byte

	if useAskpass {
		// Use ksshaskpass or similar SSH_ASKPASS program
		askpassCmd := os.Getenv("SSH_ASKPASS")
		if askpassCmd == "" {
			askpassCmd = "ksshaskpass"
		}

		// Prompt for secret
		cmd := exec.Command(askpassCmd, "Enter secret:")
		output, err := cmd.Output()
		if err != nil {
			log.Fatalf("Failed to run askpass program: %v", err)
		}
		textSecret = output

		// Prompt for salt
		cmd = exec.Command(askpassCmd, "Enter salt:")
		output, err = cmd.Output()
		if err != nil {
			log.Fatalf("Failed to run askpass program: %v", err)
		}
		textSecretSalt = output
	} else if isStdinPiped {
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

	// Output the hash to stdout
	fmt.Println(hash)
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
