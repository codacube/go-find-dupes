package processor

import (
	"log"
	"sort"

	"github.com/codacube/go-find-dupes/internal/model"
	"github.com/codacube/go-find-dupes/internal/storage"
)

// TODO Callback for progress updates
// type ProgressFunc func(current, total int)
// type ByteProgressFunc func(processedBytes, totalBytes int64)

type ProgressUpdate struct {
	CurrentItem int
	TotalItems  int

	// File stats
	CurrentFile        string
	FileBytesProcessed int64
	FileTotalBytes     int64
}

type ProgressFunc func(update ProgressUpdate)

// type Hasher interface {
// 	GetHash(f model.File, onProgress func(int64, int64)) (string, error)
// }

// type Deleter interface {
// 	DeleteFile(f model.File) error
// }

type Pass1Func func([]model.File, ProgressFunc) []model.FileGroup
type Pass2Func func([]model.FileGroup, ProgressFunc) []model.FileGroup
type Pass3Func func([]model.File, ProgressFunc) (int, int)

// Pipeline holds the three stages
type Pipeline struct {
	Pass1 Pass1Func
	Pass2 Pass2Func
	Pass3 Pass3Func
}

func NewDefaultPipeline(fileStore storage.Drive) Pipeline {
	return Pipeline{
		Pass1: GroupByFileSize,
		Pass2: func(groups []model.FileGroup, onProgress ProgressFunc) []model.FileGroup {
			return GroupByHashedDuplicates(groups, fileStore, onProgress)
		},
		Pass3: func(files []model.File, onProgress ProgressFunc) (int, int) {
			return DeleteDuplicates(files, fileStore, onProgress)
		},
	}
}

// --- Pass 1: Group files by size (using 'allFiles' list) ---
/*
	- Fail fast and filter out the non dupes.
	- Checking file size is much faster than hashing.
	- Files with unique sizes won't be a dupe, and can be skipped in the next pass.
	- Returns groups of files that have the same file size
*/
func GroupByFileSize(files []model.File, onProgress ProgressFunc) []model.FileGroup {
	grouped := make(map[int64][]model.File)
	var result []model.FileGroup
	total := len(files)

	for i, f := range files {

		if onProgress != nil {
			onProgress(ProgressUpdate{
				CurrentItem: i + 1,
				TotalItems:  total,
			})
		}

		// Skip files with no content
		if f.Size == 0 {
			continue
		}
		grouped[f.Size] = append(grouped[f.Size], f)
	}

	// Keep only groups with duplicate file sizes
	for size, group := range grouped {
		if len(group) < 2 {
			continue
		}

		result = append(result, model.FileGroup{
			FileSize: size,
			Files:    group,
		})
	}

	return result
}

// Pass 2: Returns groups of files that have the same hash
func GroupByHashedDuplicates(groups []model.FileGroup, fileStore storage.Hasher, onProgress ProgressFunc) []model.FileGroup {
	var duplicates []model.FileGroup

	// Get total num files to hash
	totalFilesToHash := 0
	for _, g := range groups {
		if len(g.Files) > 1 {
			totalFilesToHash += len(g.Files)
		}
	}
	processedCount := 0

	for _, group := range groups {

		// Must be more than one file in a group (otherwise no duplicates)
		if (len(group.Files)) < 2 {
			continue
		}

		// For each group (of file size), group files with same hash
		hashGroup := make(map[string][]model.File)
		for _, f := range group.Files {
			// Define the callback for THIS specific file
			fileProgress := func(processedBytes, totalBytes int64) {
				if onProgress != nil {
					// Send a combined update
					onProgress(ProgressUpdate{
						CurrentItem:        processedCount + 1,
						TotalItems:         totalFilesToHash,
						CurrentFile:        f.Name,
						FileBytesProcessed: processedBytes,
						FileTotalBytes:     totalBytes,
					})
				}
			}

			hash, err := fileStore.GetHash(f, fileProgress)
			if err != nil {
				log.Printf("Error: Failed to hash file '%s': %v\n", f.DisplayPath, err)
				continue
			}
			hashGroup[hash] = append(hashGroup[hash], f)

			processedCount++

			if onProgress != nil {
				// Send a combined update
				onProgress(ProgressUpdate{
					CurrentItem:        processedCount,
					TotalItems:         totalFilesToHash,
					CurrentFile:        f.Name,
					FileBytesProcessed: f.Size,
					FileTotalBytes:     f.Size,
				})
			}
		}

		// TODO Check this is correct - what if it's the first file found with that hash?
		// Convert back to FileGroups
		for hash, dupes := range hashGroup {
			// Must be more than one file in a group (otherwise no duplicates)
			if len(dupes) > 1 {
				duplicates = append(duplicates, model.FileGroup{
					Files:      dupes,
					FileSize:   group.FileSize,
					Identifier: hash,
				})
			}
		}
	}

	// Sort largest files first (remove biggest filesize duplicates first)
	sort.Slice(duplicates, func(i, j int) bool {
		return duplicates[i].FileSize > duplicates[j].FileSize
	})

	return duplicates
}

// Pass 3: Deletes duplicates
// TODO Could return a list of files not deleted? or an error?
func DeleteDuplicates(files []model.File, fileStore storage.Deleter, onProgress ProgressFunc) (int, int) {
	// TODO It would be better if we received a file list to delete, rather than the full groups
	// TODO That way I can have one counter, rather than seperate one for DeleteCount (ie. we'd use processed and total only)

	numFilesDeleted := 0
	totalFilesToDelete := len(files)

	// For each group (of file size), group files with same hash
	for _, file := range files {

		if file.MarkForDeletion {
			err := fileStore.DeleteFile(file)
			if err != nil {
				log.Printf("Error: Failed to delete file '%s': %v\n", file.DisplayPath, err)
				continue
			}

			numFilesDeleted++

			if onProgress != nil {
				// Send a combined update
				onProgress(ProgressUpdate{
					CurrentItem:        numFilesDeleted,
					TotalItems:         totalFilesToDelete,
					CurrentFile:        file.Name,
					FileBytesProcessed: file.Size,
					FileTotalBytes:     file.Size,
				})
			}
		}
	}

	return totalFilesToDelete, numFilesDeleted
}
