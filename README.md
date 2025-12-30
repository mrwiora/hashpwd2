# hashpwd2

**A secure password-to-hash tool for cryptsetup and disk encryption**

`hashpwd2` converts passwords into strong, deterministic base64-encoded hashes using Argon2id, perfect for use with cryptsetup LUKS encryption. It's designed to transform memorable passwords into cryptographically strong passphrases.

## Features

- 🔐 **Argon2id hashing** - Industry-standard memory-hard key derivation
- 🔄 **Deterministic output** - Same password + salt = same hash every time
- 🚰 **Pipe-friendly** - Seamlessly integrates with cryptsetup via stdin/stdout
- 🎯 **Base64 encoding** - Clean, filesystem-safe 86-character output
- 🛡️ **No clipboard dependencies** - Secure piped operation only
- 🐛 **Debug mode** - Inspect password/salt input for troubleshooting
- ⚡ **Optimized parameters** - 1GB memory, 16 iterations, 4 threads

## Installation

### From Arch Linux AUR
```bash
yay -S hashpwd2
# or
paru -S hashpwd2
```

### From Binary Release
```bash
wget https://github.com/mrwiora/hashpwd2/releases/latest/download/hashpwd2
chmod +x hashpwd2
sudo mv hashpwd2 /usr/local/bin/
```

### From Source
```bash
git clone https://github.com/mrwiora/hashpwd2.git
cd hashpwd2
make build
sudo make install-system
```

## Usage

### Basic Usage with cryptsetup

The primary use case is piping the hash directly to cryptsetup:

```bash
# Open LUKS volume
./hashpwd2 | sudo cryptsetup luksOpen /dev/sda1 cryptsda1 --key-file=-

# Format new LUKS volume
./hashpwd2 | sudo cryptsetup luksFormat /dev/sda1 --key-file=-
```

When you run the command, you'll be prompted:
```
Enter secret: [type your password]
Enter salt: [type your salt]
Please wait!
[hash is piped to cryptsetup]
```

### Interactive Usage

```bash
# Generate hash and see output
./hashpwd2
Enter secret: mypassword
Enter salt: mysalt
Please wait!
HoTBwJKbaiIUMcK7xd2igzpjh1QFsAwCSA3jQb9Ai3iSPDTGM8yIk4UrYvFewW9+Y+GMU4rRqs/u/yUXCdzo1A
```

### Piped Input

```bash
# Provide password and salt via stdin
echo -e "mypassword\nmysalt" | ./hashpwd2 | cat

# Or use printf for better compatibility
printf 'mypassword\nmysalt\n' | ./hashpwd2 | cat
```

### Debug Mode

```bash
# See what the application is receiving
echo -e "test\nsalt" | ./hashpwd2 --debug
Enter secret: 
Enter salt: 
DEBUG: Password length: 4
DEBUG: Password (hex): 74657374
DEBUG: Salt length: 4
DEBUG: Salt (hex): 73616c74
Please wait!
[hash output]
```

### Version Information

```bash
./hashpwd2 --version
hashpwd2 version 0.0.1
commit: abc1234
built: 2025-12-30T03:45:00Z
```

## How It Works

1. **Input**: Reads password and salt from stdin (terminal or pipe)
2. **Processing**: Applies Argon2id with these parameters:
   - Memory: 1 GB (1,048,576 KB)
   - Iterations: 16
   - Parallelism: 4 threads
   - Output length: 64 bytes
3. **Output**: 86-character base64-encoded hash (RawStdEncoding, no padding)

### Password vs Salt

- **Password**: Your memorable secret (e.g., "MySecurePassword123")
- **Salt**: Additional entropy, can be empty or another value (e.g., "server01" or "backup-disk")
- **Result**: Deterministic hash unique to the password+salt combination

Using different salts with the same password produces completely different hashes, allowing you to:
- Derive multiple encryption keys from one password
- Use the same password for different volumes securely
- Add context-specific entropy (hostname, disk label, etc.)

## Security Considerations

### ✅ Good Practices

- Use strong, random passwords with the salt for maximum entropy
- Store salts separately from encrypted volumes (e.g., password manager, paper backup)
- Use different salts for different volumes/systems
- The hash is suitable for cryptsetup (86 characters of base64 = ~512 bits of entropy input)

### ⚠️ Important Notes

- Same password + salt always produces the same hash (by design)
- The strength of your encryption depends on both password AND salt strength
- Argon2 is memory-hard but not immune to dedicated hardware attacks
- This tool doesn't replace good password management practices

### 🔒 Why No Clipboard?

Previous versions supported clipboard operations, but this introduced:
- X11 library dependencies (platform-specific)
- Security risks (clipboard history, other apps reading clipboard)
- Complexity for headless/CI environments

The current pipe-only design is simpler, more secure, and more portable.

## Examples

### Example 1: Unlocking System Disk at Boot

```bash
# Create encrypted volume
printf 'MyStrongPassword\nserver-hostname\n' | sudo cryptsetup luksFormat /dev/sda1 --key-file=-

# Mount at boot (add to /etc/crypttab)
# cryptroot /dev/sda1 none luks

# Then use hashpwd2 in initramfs or boot script
```

### Example 2: Multiple Volumes, One Password

```bash
# Volume 1 with salt "backup"
printf 'MyPassword\nbackup\n' | sudo cryptsetup luksFormat /dev/sdb1 --key-file=-

# Volume 2 with salt "data"
printf 'MyPassword\ndata\n' | sudo cryptsetup luksFormat /dev/sdc1 --key-file=-

# Same password, different salts = different encryption keys
```

### Example 3: Scripted Backup Encryption

```bash
#!/bin/bash
HOSTNAME=$(hostname)
printf 'BackupPassword\n%s\n' "$HOSTNAME" | \
    sudo cryptsetup luksOpen /dev/sdd1 backup-$(date +%Y%m%d) --key-file=-
```

## Building

### Prerequisites

- Go 1.19 or higher
- Standard build tools (make, gcc)

### Build Commands

```bash
# Build binary
make build

# Build optimized (smaller binary)
make build-optimized

# Run tests
make test

# Install to /usr/local/bin
sudo make install-system

# Clean build artifacts
make clean

# Show version information
make version
```

## Testing

The project includes comprehensive integration tests:

```bash
# Run all tests
go test -v -timeout 10m ./...

# Run specific test
go test -v -run TestHashCalculation

# Run with debug output
go test -v -run TestDebugInputReading
```

Tests verify:
- ✅ Correct hash generation for various inputs
- ✅ Proper stdin/stdout/stderr separation
- ✅ Special character handling
- ✅ Piped input/output behavior
- ✅ Consistent hashing (deterministic output)

## Architecture

```
┌─────────────┐
│   stdin     │ ← Password + Salt (terminal or pipe)
└─────┬───────┘
      │
      v
┌─────────────────────────────┐
│  Password Reading           │
│  - Terminal: term.ReadPassword (hidden)
│  - Piped: bufio.Reader (visible)
└─────┬───────────────────────┘
      │
      v
┌─────────────────────────────┐
│  Argon2id Key Derivation    │
│  - Memory: 1 GB             │
│  - Iterations: 16           │
│  - Parallelism: 4           │
│  - Output: 64 bytes         │
└─────┬───────────────────────┘
      │
      v
┌─────────────────────────────┐
│  Base64 Encoding            │
│  - RawStdEncoding (no pad)  │
│  - Output: 86 chars         │
└─────┬───────────────────────┘
      │
      v
┌─────────────┐
│   stdout    │ → Hash (86-character base64 string)
└─────────────┘
┌─────────────┐
│   stderr    │ → Prompts and status messages
└─────────────┘
```

## Troubleshooting

### Issue: Different hash in CI/testing

**Solution**: Use debug mode to verify input:
```bash
printf 'test\nsalt\n' | ./hashpwd2 --debug
```

Check that password/salt lengths and hex values match expectations.

### Issue: cryptsetup says "No key available"

**Possible causes**:
1. Different password or salt used during format vs open
2. Extra newlines in input
3. Shell escaping issues with special characters

**Solution**: Use printf with explicit newlines and test with debug mode first.

### Issue: Special characters not working

**Solution**: Use printf instead of echo, or quote properly:
```bash
# Good
printf 'p@ssw0rd!#$\ns@lt\n' | ./hashpwd2

# Also good
./hashpwd2  # Type interactively
```

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `go test -v ./...`
5. Submit a pull request

## License

BSD-3-Clause License - See [LICENSE](LICENSE) file for details.

## Author

Matthias R. Wiora <matthias@wiora.io>

## See Also

- [cryptsetup](https://gitlab.com/cryptsetup/cryptsetup) - Linux disk encryption
- [Argon2](https://github.com/P-H-C/phc-winner-argon2) - Password hashing competition winner
- [LUKS](https://gitlab.com/cryptsetup/cryptsetup/-/wikis/LUKS-standard) - Linux Unified Key Setup

---

**Note**: This tool is designed for advanced users who understand disk encryption. Always keep secure backups of your encryption keys/passwords!