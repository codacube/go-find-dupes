package model

type File struct {
	Name string
	Size int64

	// TODO ID -> should be Path?
	// ID is the "Primary Key" for the file in its source system
	// - Local: "/home/user/docs/file.txt" (Absolute Path)
	// - GDrive: "1A2B3C4D5E..." (The File ID)
	ID string

	Sha256Checksum string
	Md5Checksum    string

	// DisplayPath is human-readable for the UI
	// - Local: "/home/user/docs/file.txt"
	// - GDrive: "My Drive > Projects > file.txt"
	DisplayPath string

	// Stores the raw original object or extra data needed by the source.
	// The UI/Common logic ignores this, but the specific Scanner can read it back.
	// The common pipeline functions shouldn't need to look at this
	Metadata any

	// Used in the review phase
	// Note: In a purist world, this is part of the UI and would not be in the
	//       model, but for simplicity I've put it here.
	// eg. in TUI: type FileItem struct { File model.File; MarkedForDeletion bool }
	// and we'd have to convert from/to when passing data between UI and processing
	MarkForDeletion bool
}

type FileGroup struct {
	Files      []File
	FileSize   int64  // Size of the files in this group (all be the same)
	Identifier string // e.g. "1024_bytes" or a Hash
}

// Use count for calculating multiples of filesize (eg. many files of same size)
func (fg *FileGroup) TotalSize(count int) int64 {
	if count > 0 {
		return int64(count) * fg.FileSize
	}

	// Default to all files in the group
	return int64(len(fg.Files)) * fg.FileSize
}

// TODO Should this go here or TUI?
func GetFilesToDelete(groups []FileGroup) []File {
	var deleteList []File

	for _, group := range groups {
		for _, f := range group.Files {
			if f.MarkForDeletion {
				deleteList = append(deleteList, f)
			}
		}

	}
	return deleteList
}
