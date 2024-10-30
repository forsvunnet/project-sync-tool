// internal/collections/checksum.go

package collections

import (
    "crypto/sha256"
    "fmt"
    "io"
    "os"
)

// calculateChecksum calculates the SHA-256 checksum of a file at the given path.
func calculateChecksum(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", fmt.Errorf("failed to open file for checksum calculation: %w", err)
    }
    defer file.Close()

    hasher := sha256.New()
    if _, err := io.Copy(hasher, file); err != nil {
        return "", fmt.Errorf("failed to calculate checksum for file: %w", err)
    }

    return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

