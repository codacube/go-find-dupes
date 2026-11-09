package local

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/codacube/go-find-dupes/internal/model"

	"github.com/hymkor/trash-go"
)

// Implements the interfaces for the local filesystem
// Stores a valid path
type Client struct {
	root string
}

// NewClient is the Factory - returns concrete struct
func NewClient(rootPath string) (*Client, error) {
	// Check if path exists
	info, err := os.Stat(rootPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", rootPath)
	}
	if err != nil {
		return nil, fmt.Errorf("error accessing path: %v", err)
	}

	// Check if it is actually a directory
	if !info.IsDir() {
		return nil, fmt.Errorf("path is a file, not a directory: %s", rootPath)
	}

	// Check read permissions by trying to open it
	// This catches "Permission Denied" errors early.
	f, err := os.Open(rootPath)
	if err != nil {
		return nil, fmt.Errorf("permission denied reading directory: %v", err)
	}
	f.Close() // Close immediately, we just wanted to check access

	// 4. Return the valid client
	return &Client{root: rootPath}, nil
}

// GetAllFiles recursively walks the folder at 'root'
func (c *Client) GetAllFiles(rootDir string, onProgress func(count int)) ([]model.File, error) {
	var allFiles []model.File
	totalCount := 0

	// 1. Walk the directory tree
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		// handle permission errors or non-existent paths gracefully
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		allFiles = append(allFiles, model.File{
			Name:        d.Name(),
			Size:        info.Size(),
			ID:          path,
			DisplayPath: path,
			Metadata:    nil,
		})

		totalCount++

		if onProgress != nil {
			onProgress(totalCount)
		}

		return nil
	})

	if onProgress != nil {
		onProgress(totalCount)
	}

	if err != nil {
		return nil, err
	}

	return allFiles, nil
}

// onProgress is an optional callback to report progress of hashing (bytes processed, total bytes)
func (c *Client) GetHash(f model.File, onProgress func(int64, int64)) (string, error) {
	file, err := os.Open(f.ID)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	const chunkSize = 32 * 1024 // 32KB
	buffer := make([]byte, chunkSize)
	var bytesRead int64 = 0

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			hasher.Write(buffer[:n])
			bytesRead += int64(n)

			if onProgress != nil {
				onProgress(bytesRead, f.Size)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (c *Client) DeleteFile(f model.File) error {

	var logFileDetails string
	if f.ID == f.DisplayPath {
		logFileDetails = f.ID
	} else {
		logFileDetails = fmt.Sprintf("ID=%s, Path=%s", f.ID, f.DisplayPath)
	}
	log.Printf("Deleted file: %s\n", logFileDetails)

	// TODO Check flag here for dry run
	// TODO Skip recycle bin
	return trash.Throw(f.ID)
	// return os.Remove(f.ID)
}
