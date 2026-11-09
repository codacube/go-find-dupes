package gdrive

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/codacube/go-find-dupes/internal/auth"
	"github.com/codacube/go-find-dupes/internal/model"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// "github.com/codacube/go-find-dupes/internal/auth"

// Implements the interfaces for GDrive
// Stores the service connection
type Client struct {
	srv *drive.Service
}

// TODO Currently uses a Service Account, needs to be oauth2

// NewClient initializes the Drive service
// Ideally, you pass an authorized http.Client or credentials path here.
func NewClient(credentialsPath string) (*Client, error) {
	// TODO What is Background? was cmd.Context() before
	ctx := context.Background()

	// TODO Check it exists?
	// 1. Read the credentials file (downloaded from Google Cloud Console)
	// b, err := os.ReadFile(credentialsPath)
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to read client secret file: %v", err)
	// }

	client, err := auth.GetGoogleDriveClient(ctx, credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Unable to get Drive service: %v", err)
	}

	return &Client{srv: srv}, nil
}

func (g *Client) GetAllFiles(rootFolderId string, onProgress func(count int)) ([]model.File, error) {
	var allFiles []model.File
	totalCount := 0
	rootName, _ := g.GetFullPath(rootFolderId, 0)

	// Recursive function to scan folders
	var scanFolder func(folderId string, currentPath string) error
	scanFolder = func(currentFolderId string, currentPath string) error {
		var pageToken string

		// Query for all items (files and folders) in this folder (and not in trash)
		query := fmt.Sprintf("'%s' in parents and trashed = false", currentFolderId)

		for {
			q := g.srv.Files.List().
				Q(query).
				PageSize(1000).
				Fields("nextPageToken, files(id, name, size, md5Checksum, sha256Checksum, parents, mimeType)").
				PageToken(pageToken)
			fileList, err := q.Do()

			if err != nil {
				return fmt.Errorf("Unable to retrieve files for folder %s: %v", currentFolderId, err)
			}

			for _, f := range fileList.Files {
				// Recursively scan folders
				if f.MimeType == "application/vnd.google-apps.folder" {
					newPath := filepath.Join(currentPath, f.Name)
					if err := scanFolder(f.Id, newPath); err != nil {
						return err
					}
				} else {
					// Nothing to see here, just some hangover
					// hardcoded laziness for targetting cr2 files
					// ext := strings.ToLower(filepath.Ext(f.Name))
					// if len(ext) > 1 && ext[1:] == "cr2" {
					// allFiles = append(allFiles, f)
					// }

					fullPath := filepath.Join(currentPath, f.Name)

					allFiles = append(allFiles, model.File{
						Name:           f.Name,
						Size:           f.Size,
						ID:             f.Id,
						Sha256Checksum: f.Sha256Checksum,
						Md5Checksum:    f.Md5Checksum,
						DisplayPath:    fullPath,
						Metadata:       f,
					})

					totalCount++

					if onProgress != nil {
						onProgress(totalCount)
					}
				}
			}

			pageToken = fileList.NextPageToken
			if pageToken == "" {
				break
			}
		}

		return nil
	}

	err := scanFolder(rootFolderId, rootName)
	if onProgress != nil {
		onProgress(totalCount)
	}

	return allFiles, err
}

func (g *Client) GetHash(f model.File, onProgress func(int64, int64)) (string, error) {
	// We already have the md5 hash in the metadata
	if val, ok := f.Metadata.(*drive.File); ok {
		return val.Sha256Checksum, nil
	}
	return "", fmt.Errorf("metadata is missing or invalid for file ID %s", f.ID)
}

func (g *Client) DeleteFile(f model.File) error {
	// We know that for GDrive, ID == FileID
	// log.Printf("Deleted file: %s, %s\n", f.DisplayPath, f.ID)
	return g.srv.Files.Delete(f.ID).Do()

	// return nil
}

// TODO This isn't part of the Scanner interface, does that matter?
// GetFullPath takes a folderID and walks up the tree to build the path (until
// it reaches the root or depth)
// Note: It's slow and counts towards API usage - use sparingly!
func (g *Client) GetFullPath(folderID string, depth int) (string, error) {
	var pathParts []string
	currentID := folderID
	currentDepth := 0
	if depth < 0 {
		depth = 0
	}

	// Loop until we hit the root or run out of parents
	for {
		// We only need the Name and Parents fields to reconstruct the path
		// Using 'fields' greatly reduces the API response size
		file, err := g.srv.Files.Get(currentID).
			Fields("id, name, parents").
			Do()

		if err != nil {
			return "", fmt.Errorf("unable to retrieve file %s: %v", currentID, err)
		}

		// Prepend the current folder name to our list
		// (We are walking backwards, add to the front)
		pathParts = append([]string{file.Name}, pathParts...)

		// Check if we have reached the root (no parents)- or depth
		if len(file.Parents) == 0 || currentDepth == depth {
			break
		}

		// Move up to the next parent
		// Note: Drive allows multiple parents, but we just follow the first one
		currentID = file.Parents[0]
	}

	// Join the parts with a slash
	//return "/" + strings.Join(pathParts, "/"), nil
	return filepath.Join(pathParts...), nil
}
