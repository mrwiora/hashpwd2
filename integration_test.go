package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// Integration tests for hashpwd2
//
// These tests ONLY test the piped mode of the application:
// - Input is provided via stdin pipe (echo -e 'password\nsalt' | ./hashpwd2)
// - Output is captured via stdout pipe (./hashpwd2 | cat)
// - Clipboard functionality is NEVER triggered because isStdoutPiped=true
// - Tests can run in headless/CI environments without X11 display
//
// The tests verify:
// 1. Correct hash calculation for various inputs
// 2. Proper separation of informational messages (stderr) vs hash output (stdout)
// 3. Consistent hashing for same inputs
// 4. Performance benchmarks

// TestDebugInputReading tests what the application actually receives as input
func TestDebugInputReading(t *testing.T) {
	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "hashpwd2_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer exec.Command("rm", "-f", "hashpwd2_test").Run()

	tests := []struct {
		name     string
		password string
		salt     string
	}{
		{"abc with abc", "abc", "abc"},
		{"abc with empty", "abc", ""},
		{"testpassword with testsalt", "testpassword", "testsalt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputCmd := "printf '" + tt.password + "\\n" + tt.salt + "\\n' | ./hashpwd2_test --debug"
			cmd := exec.Command("sh", "-c", inputCmd)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to run application: %v\nStderr: %s", err, stderr.String())
			}

			t.Logf("Input: password='%s', salt='%s'", tt.password, tt.salt)
			t.Logf("Stderr output:\n%s", stderr.String())
			t.Logf("Stdout (hash):\n%s", stdout.String())
		})
	}
}

// TestHashCalculation tests that the application produces the correct hash
// when given specific inputs via piped stdin/stdout (no clipboard involved)
func TestHashCalculation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		salt     string
		expected string
	}{
		{
			name:     "password with empty salt",
			password: "abc",
			salt:     "",
			expected: "N+WFT6oRgFxMB3voTrB/n+2O7ebp7jz621UZT2EWIOYdfMedRNPoULcb1F5MKs5HkBvY0Mo7Kt26NdRuIRjWSQ",
		},
		{
			name:     "password with salt",
			password: "abc",
			salt:     "abc",
			expected: "nZjkO5tW6Vz05mTlng0c8RYomnLLUE1CZrGIgwJlTI1UdUpZCJb+Ai1kAj6NVuNpjmZWXT3iK4i6Up0M8K72Pw",
		},
		{
			name:     "complex password with salt",
			password: "testpassword",
			salt:     "testsalt",
			expected: "HoTBwJKbaiIUMcK7xd2igzpjh1QFsAwCSA3jQb9Ai3iSPDTGM8yIk4UrYvFewW9+Y+GMU4rRqs/u/yUXCdzo1A",
		},
		{
			name:     "special characters in password and salt",
			password: "p@ssw0rd!#$%",
			salt:     "s@lt&*()",
			expected: "eim3H9sv+vlPdU0RsNIU/VEXm9emN8C/1VXMuLqkjDhnAfHmd2GAPZw4Qly6oB+aNT+/aZ4LGtTnqi06HMKigQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the application - we only test piped mode, so clipboard is never initialized
			buildCmd := exec.Command("go", "build", "-o", "hashpwd2_test", ".")
			if err := buildCmd.Run(); err != nil {
				t.Fatalf("Failed to build application: %v", err)
			}
			defer exec.Command("rm", "-f", "hashpwd2_test").Run()

			// Prepare input using printf for proper newline handling
			// We need to use sh -c because Go's exec doesn't support pipes directly
			inputCmd := "printf '" + tt.password + "\\n" + tt.salt + "\\n' | ./hashpwd2_test"

			cmd := exec.Command("sh", "-c", inputCmd)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to run application: %v\nStderr: %s", err, stderr.String())
			}

			// Get the output and trim whitespace
			output := strings.TrimSpace(stdout.String())

			// Verify the hash matches expected
			if output != tt.expected {
				t.Errorf("Hash mismatch\nExpected: %s\nGot:      %s", tt.expected, output)
			}

			// Verify that informational messages went to stderr, not stdout
			stderrStr := stderr.String()
			if !strings.Contains(stderrStr, "Enter secret:") {
				t.Error("Expected 'Enter secret:' in stderr")
			}
			if !strings.Contains(stderrStr, "Enter salt:") {
				t.Error("Expected 'Enter salt:' in stderr")
			}
			if !strings.Contains(stderrStr, "Please wait!") {
				t.Error("Expected 'Please wait!' in stderr")
			}

			// Verify that only the hash is in stdout (no prompts)
			if strings.Contains(output, "Enter secret:") {
				t.Error("Stdout should not contain prompts when piped")
			}
			if strings.Contains(output, "Enter salt:") {
				t.Error("Stdout should not contain prompts when piped")
			}
			if strings.Contains(output, "Please wait!") {
				t.Error("Stdout should not contain 'Please wait!' when piped")
			}
			if strings.Contains(output, "OK! Hurry up") {
				t.Error("Stdout should not contain clipboard message when piped")
			}
		})
	}
}

// TestPipedOutput specifically tests the piped behavior with cat (no clipboard)
func TestPipedOutput(t *testing.T) {
	// Build the application - we only test piped mode, so clipboard is never initialized
	buildCmd := exec.Command("go", "build", "-o", "hashpwd2_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer exec.Command("rm", "-f", "hashpwd2_test").Run()

	// Test piped scenario: printf | ./hashpwd2_test | cat
	// Using sh -c to properly handle the pipe chain
	cmd := exec.Command("sh", "-c", "printf 'abc\\nabc\\n' | ./hashpwd2_test | cat")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run piped command: %v\nStderr: %s", err, stderr.String())
	}

	// Verify output
	result := strings.TrimSpace(stdout.String())
	expected := "nZjkO5tW6Vz05mTlng0c8RYomnLLUE1CZrGIgwJlTI1UdUpZCJb+Ai1kAj6NVuNpjmZWXT3iK4i6Up0M8K72Pw"

	if result != expected {
		t.Errorf("Piped output mismatch\nExpected: %s\nGot:      %s", expected, result)
	}

	// Verify stderr contains prompts
	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Enter secret:") {
		t.Error("Expected 'Enter secret:' in stderr")
	}
	if !strings.Contains(stderrStr, "Enter salt:") {
		t.Error("Expected 'Enter salt:' in stderr")
	}
}

// TestStderrStdoutSeparation verifies that prompts go to stderr and hash to stdout (piped mode)
func TestStderrStdoutSeparation(t *testing.T) {
	// Build the application - we only test piped mode, so clipboard is never initialized
	buildCmd := exec.Command("go", "build", "-o", "hashpwd2_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer exec.Command("rm", "-f", "hashpwd2_test").Run()

	// Run with pipe to ensure piped mode is detected
	cmd := exec.Command("sh", "-c", "printf 'abc\\n\\n' | ./hashpwd2_test | cat")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run application: %v", err)
	}

	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	// Stdout should ONLY contain the hash (and maybe trailing newline)
	stdoutLines := strings.Split(strings.TrimSpace(stdoutStr), "\n")
	if len(stdoutLines) != 1 {
		t.Errorf("Expected exactly 1 line in stdout (the hash), got %d lines:\n%s", len(stdoutLines), stdoutStr)
	}

	// Stderr should contain all the prompts
	requiredPrompts := []string{"Enter secret:", "Enter salt:", "Please wait!"}
	for _, prompt := range requiredPrompts {
		if !strings.Contains(stderrStr, prompt) {
			t.Errorf("Expected '%s' in stderr, but it was not found", prompt)
		}
	}

	// Stdout should NOT contain any prompts
	for _, prompt := range requiredPrompts {
		if strings.Contains(stdoutStr, prompt) {
			t.Errorf("Stdout should not contain '%s'", prompt)
		}
	}

	// Verify the hash format (base64, no padding)
	hash := strings.TrimSpace(stdoutStr)
	if len(hash) != 86 { // 64 bytes base64-encoded without padding = 86 chars
		t.Errorf("Expected hash length of 86 characters, got %d", len(hash))
	}
}

// TestConsistentHashing verifies that same inputs produce same outputs (piped mode)
func TestConsistentHashing(t *testing.T) {
	// Build the application - we only test piped mode, so clipboard is never initialized
	buildCmd := exec.Command("go", "build", "-o", "hashpwd2_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer exec.Command("rm", "-f", "hashpwd2_test").Run()

	// Run the same input twice
	password := "consistent"
	salt := "test"
	input := "printf '" + password + "\\n" + salt + "\\n' | ./hashpwd2_test"

	var hash1, hash2 string

	for i := 0; i < 2; i++ {
		cmd := exec.Command("sh", "-c", input)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		if err := cmd.Run(); err != nil {
			t.Fatalf("Run %d failed: %v", i+1, err)
		}

		hash := strings.TrimSpace(stdout.String())
		if i == 0 {
			hash1 = hash
		} else {
			hash2 = hash
		}
	}

	if hash1 != hash2 {
		t.Errorf("Hashes are not consistent!\nFirst:  %s\nSecond: %s", hash1, hash2)
	}

	t.Logf("Consistent hash for password='%s', salt='%s': %s", password, salt, hash1)
}

// BenchmarkHashGeneration benchmarks the hash generation performance
func BenchmarkHashGeneration(b *testing.B) {
	p := &params{
		memory:      1 * 1024 * 1024,
		iterations:  16,
		parallelism: 4,
		keyLength:   64,
	}

	password := "testpassword"
	salt := "testsalt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generateFromPassword(password, salt, p)
		if err != nil {
			b.Fatalf("Hash generation failed: %v", err)
		}
	}
}
