package storage

import "github.com/codacube/go-find-dupes/internal/model"

// Interface describes methods that can be called on an object after
// it has been created

type Drive interface {
	Scanner // GetAllFiles
	Hasher  // GetHash
	Deleter // DeleteFiles
}

type Scanner interface {
	GetAllFiles(root string, onProgress func(count int)) ([]model.File, error)
}

type Hasher interface {
	GetHash(f model.File, onProgress func(int64, int64)) (string, error)
}

type Deleter interface {
	DeleteFile(f model.File) error
}
