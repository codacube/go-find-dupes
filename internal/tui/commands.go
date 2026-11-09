package tui

import (
	"fmt"
	"time"

	"github.com/codacube/go-find-dupes/internal/exporter"
	"github.com/codacube/go-find-dupes/internal/model"
	"github.com/codacube/go-find-dupes/internal/processor"
	"github.com/codacube/go-find-dupes/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

// Commands are functions that return a message to the Update loop

func clearStatusAfter(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return ClearStatusMsg{}
	})
}

func displayErr(err string) tea.Cmd {
	return func() tea.Msg {
		errMsg := fmt.Errorf("%s", err)
		return ErrMsg{err: errMsg}
	}
}

// Waits for progress updates and sends them to the channel
func waitForProgressUpdateCmd(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		progressMsg := <-ch
		return progressMsg
	}
}

func exportCmd(groups []model.FileGroup, exporterType string) tea.Cmd {
	return func() tea.Msg {
		var err error
		var filename string

		timestamp := time.Now().Format("20060102-150405")

		switch exporterType {
		case "json":
			filename = fmt.Sprintf("duplicates_%s.json", timestamp)
			err = exporter.ToJSON(groups, filename)
		case "csv":
			filename = fmt.Sprintf("duplicates_%s.csv", timestamp)
			err = exporter.ToCSV(groups, filename)
		default:
			err = fmt.Errorf("Unknown export type: %s", exporterType)
		}

		if err != nil {
			return ErrMsg{err}
		}

		return UpdateStatusMsg{status: fmt.Sprintf("Saved to %s", filename), duration: 3}
	}
}

// -- Async Commands --
// Non blocking functions so the UI doesn't freeze running our tasks

func scanFilesCmd(s storage.Scanner, target string, ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {

		lastUpdate := time.Now()
		updateInterval := 100 * time.Millisecond

		onProgress := func(count int) {
			// TODO Update comment
			// Send the raw number to the UI loop
			// Note: We use a separate channel or program.Send() if this is
			// running in a goroutine, but for simple commands, passing a
			// 'program' reference is tricky.

			// BETTER WAY for Bubble Tea Commands:
			// You actually cannot emit multiple messages from one Cmd EASILY.
			// (A Cmd returns exactly ONE Msg).

			// To do "Streaming" updates in Bubble Tea, you usually send
			// to a channel that the UI is listening to (via a 'WaitForActivity' Cmd).
			if time.Since(lastUpdate) > updateInterval {
				ch <- ScanProgressMsg(count)
				lastUpdate = time.Now()
			}
		}

		files, err := s.GetAllFiles(target, onProgress)

		// Force the final update
		if len(files) > 0 {
			ch <- ScanProgressMsg(len(files))
		}

		if err != nil {
			return ErrMsg{err}
		}
		return ScanFinishedMsg{files: files}
	}
}

func runPass1Cmd(files []model.File, p processor.Pipeline, ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		lastUpdate := time.Now()

		// TODO This is in all three functions, put into own helper?
		// Wrap the onProgress function to throttle updates
		onProgress := func(processor.ProgressUpdate) {
			now := time.Now()
			if now.Sub(lastUpdate) < 100*time.Millisecond {
				return // Throttle to every 100ms
			}
			lastUpdate = now
			// TODO Here we could send progress updates if needed
		}
		// Replace the call to Pass1 with one that includes progress
		// Call your pure processor function
		res := p.Pass1(files, onProgress)
		return Pass1FinishedMsg{groups: res}
	}
}

func runPass2Cmd(groups []model.FileGroup, p processor.Pipeline, ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		lastUpdate := time.Now()

		// Wrap the onProgress function to send updates to the channel
		onProgress := func(update processor.ProgressUpdate) {
			if update.CurrentItem == update.TotalItems || time.Since(lastUpdate) >= 100*time.Millisecond {
				pct := 0.0
				if update.FileBytesProcessed > 0 {
					pct = float64(update.FileBytesProcessed) / float64(update.FileTotalBytes)
				}
				msg := ProgressMsg{
					NumProcessedFiles: update.CurrentItem,
					TotalFiles:        update.TotalItems,
					OverallPercent:    float64(update.CurrentItem) / float64(update.TotalItems),
					CurrentFilename:   update.CurrentFile,
					FilePercent:       pct,
				}

				ch <- msg
				lastUpdate = time.Now()
			}
		}
		res := p.Pass2(groups, onProgress)
		return Pass2FinishedMsg{groups: res}
	}
}

func runPass3Cmd(files []model.File, p processor.Pipeline, ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		lastUpdate := time.Now()

		// Wrap the onProgress function to send updates to the channel
		onProgress := func(update processor.ProgressUpdate) {
			if update.CurrentItem == update.TotalItems || time.Since(lastUpdate) >= 100*time.Millisecond {
				ch <- ProgressMsg{
					NumProcessedFiles: update.CurrentItem,
					TotalFiles:        update.TotalItems,
					OverallPercent:    float64(update.CurrentItem) / float64(update.TotalItems),
					CurrentFilename:   "",
					FilePercent:       0.0,
				}
				lastUpdate = time.Now()
			}
		}
		numFilesToDelete, numFilesDeleted := p.Pass3(files, onProgress)
		return Pass3FinishedMsg{numFilesToDelete: numFilesToDelete, numFilesDeleted: numFilesDeleted}
	}
}
