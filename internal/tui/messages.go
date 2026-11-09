package tui

import "github.com/codacube/go-find-dupes/internal/model"

// These carry data back from the async commands to the Update loop

type ErrMsg struct{ err error }

type UpdateStatusMsg struct {
	status   string
	duration int // in seconds
}
type ClearStatusMsg struct{}

type StartScanMsg struct{} // StartScanMsg is a command to kick off the scanning process
type ScanFinishedMsg struct{ files []model.File }
type Pass1FinishedMsg struct{ groups []model.FileGroup }
type Pass2FinishedMsg struct{ groups []model.FileGroup }
type Pass3FinishedMsg struct {
	numFilesToDelete int
	numFilesDeleted  int
}

type ScanProgressMsg int // Used for file scanning progress

type ProgressMsg ProgressData
