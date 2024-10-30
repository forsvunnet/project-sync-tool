// internal/collections/changes.go

package collections

import (
    "os"
    "path/filepath"
    "fmt"
)

// ChangeStatus represents the comparison result between local and central files.
type ChangeStatus struct {
    LocalNewer   []string // Files in the project that are newer than those in central
    CentralNewer []string // Files in central that are newer than those in the project
}

// CheckForChanges compares files in the local directory and central collection.
// First checks file checksums; if they match, no further checks are needed.
func CheckForChanges(collectionName, projectPath string) (ChangeStatus, error) {
    collectionFiles, err := GetCollectionFiles(collectionName)
    if err != nil {
        return ChangeStatus{}, err
    }

    status := ChangeStatus{}
    for _, centralFilePath := range collectionFiles {
        relPath, err := filepath.Rel(GetCollectionPath(collectionName), centralFilePath)
        if err != nil {
            return ChangeStatus{}, fmt.Errorf("failed to calculate relative path: %w", err)
        }

        projectFilePath := filepath.Join(projectPath, relPath)

        // If the project file doesn’t exist, mark it as a new local file
        projectInfo, err := os.Stat(projectFilePath)
        if os.IsNotExist(err) {
            continue
        } else if err != nil {
            return ChangeStatus{}, fmt.Errorf("failed to stat project file %s: %w", projectFilePath, err)
        }

        // Calculate checksums for both files
        centralChecksum, err := calculateChecksum(centralFilePath)
        if err != nil {
            return ChangeStatus{}, fmt.Errorf("failed to calculate checksum for central file %s: %w", centralFilePath, err)
        }

        projectChecksum, err := calculateChecksum(projectFilePath)
        if err != nil {
            return ChangeStatus{}, fmt.Errorf("failed to calculate checksum for project file %s: %w", projectFilePath, err)
        }

        // If checksums match, skip further checks for this file
        if centralChecksum == projectChecksum {
            continue
        }

        // If checksums differ, compare timestamps to determine which is newer
        centralInfo, err := os.Stat(centralFilePath)
        if err != nil {
            return ChangeStatus{}, fmt.Errorf("failed to stat central file %s: %w", centralFilePath, err)
        }

        if projectInfo.ModTime().After(centralInfo.ModTime()) {
            status.LocalNewer = append(status.LocalNewer, projectFilePath)
        } else if centralInfo.ModTime().After(projectInfo.ModTime()) {
            status.CentralNewer = append(status.CentralNewer, projectFilePath)
        }
    }

    return status, nil
}

