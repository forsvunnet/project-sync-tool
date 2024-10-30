// internal/collections/files.go

package collections

import (
    "os"
    "path/filepath"
    "fmt"
)

// GetCollectionFiles returns a list of file paths in the central collection directory for a collection.
func GetCollectionFiles(collectionName string) ([]string, error) {
    collectionDir := GetCollectionPath(collectionName)
    files := []string{}

    err := filepath.Walk(collectionDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            files = append(files, path)
        }
        return nil
    })
    if err != nil {
        return nil, fmt.Errorf("failed to list files in collection %s: %w", collectionName, err)
    }

    return files, nil
}

