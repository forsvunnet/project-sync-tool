// internal/collections/scan.go

package collections

import (
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
    "fmt"
)

// scanForCollections searches the metadata directory for collections associated with the specified path.
func ScanForCollections(dir string) ([]string, error) {
    metaDir := filepath.Join(os.Getenv("HOME"), ".config", "project-sync-tool", "meta")
    collections := []string{}

    // Read the meta directory for each collection's YAML metadata file
    entries, err := os.ReadDir(metaDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read metadata directory: %w", err)
    }

    // Check each metadata file to see if the directory matches any listed project path
    for _, entry := range entries {
        if filepath.Ext(entry.Name()) != ".yml" {
            continue
        }

        collectionName := entry.Name()[:len(entry.Name())-4] // Remove ".yml" extension
        metaFilePath := filepath.Join(metaDir, entry.Name())

        meta, err := loadCollectionMeta(metaFilePath)
        if err != nil {
            return nil, fmt.Errorf("failed to load metadata for %s: %w", collectionName, err)
        }

        for _, path := range meta.Paths {
            absPath, _ := filepath.Abs(path)
            targetAbsPath, _ := filepath.Abs(dir)
            if absPath == targetAbsPath {
                collections = append(collections, collectionName)
                break
            }
        }
    }

    return collections, nil
}

// loadCollectionMeta loads the metadata from a specified file path
func loadCollectionMeta(metaFilePath string) (CollectionMeta, error) {
    meta := CollectionMeta{}
    data, err := os.ReadFile(metaFilePath)
    if err != nil {
        return meta, fmt.Errorf("failed to read metadata file: %w", err)
    }

    if err := yaml.Unmarshal(data, &meta); err != nil {
        return meta, fmt.Errorf("failed to unmarshal metadata: %w", err)
    }
    return meta, nil
}

