// internal/collections/utils.go
package collections

import (
    "errors"
    "regexp"
)

// validNameRegex enforces alphanumeric-only characters for collection names
var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// IsValidCollectionName checks if a collection name contains only alphanumeric characters
func IsValidCollectionName(name string) error {
    if !validNameRegex.MatchString(name) {
        return errors.New("collection name must contain only alphanumeric characters")
    }
    return nil
}

