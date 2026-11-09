package processor

import (
	"path/filepath"

	"github.com/codacube/go-find-dupes/internal/model"
)

func ApplySafePathPolicy(groups []model.FileGroup, safePath string) {
	if safePath == "" {
		return
	}

	for i := range groups {
		// We need to ensure we don't delete ALL copies of a file.
		// So, first, count how many copies exist *outside* the safe path.
		nonSafeCount := 0
		for _, file := range groups[i].Files {
			// filepath.Dir() gets the file's immediate parent directory.
			if filepath.Dir(file.ID) != safePath {
				nonSafeCount++
			}
		}

		// If there is at least one keeper outside the safe path, then
		// it's safe to mark the ones inside it for deletion.
		if nonSafeCount > 0 {
			for j, file := range groups[i].Files {
				// Perform the exact directory match. This is the crucial logic that
				// prevents matching files in subdirectories.
				if filepath.Dir(file.ID) == safePath {
					// Use a pointer to modify the actual slice element
					groups[i].Files[j].MarkForDeletion = true
				}
			}
		}
	}
}
