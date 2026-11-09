package tui

import (
	"fmt"
	"time"

	"github.com/codacube/go-find-dupes/internal/exporter"
	"github.com/codacube/go-find-dupes/internal/model"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Update is the main update loop for the TUI application (keep heavy logic out of here, it will affect UI)
// Update has 2 Jobs:
//   1. Process and update the model
//   2. Return commands to run next

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	// 1. Key Presses (Always allow user to quit)
	case tea.KeyMsg:
		if m.state == StateReview {
			// TODO Need a "s" to save as file
			switch msg.String() {
			case "up", "k":
				if m.reviewCursor > 0 {
					m.reviewCursor--
				}
			case "down", "j":
				if m.reviewCursor < len(m.groups[m.activeGroup].Files)-1 {
					m.reviewCursor++
				}
			case "left", "h":
				// TODO Move previous group
				if m.activeGroup > 0 {
					m.activeGroup--
				}
				m.reviewCursor = 0

			case "right", "l", "tab":
				// TODO Move next group
				if m.activeGroup < len(m.groups)-1 {
					m.activeGroup++
				}
				m.reviewCursor = 0
			case " ":
				// TODO Which file to keep
				toggle := &m.groups[m.activeGroup].Files[m.reviewCursor]
				if toggle.MarkForDeletion {
					toggle.MarkForDeletion = false
					m.markedForDeletion--
					m.totalBytesMarkedForDeleted -= m.groups[m.activeGroup].FileSize
				} else {
					toggle.MarkForDeletion = true
					m.markedForDeletion++
					m.totalBytesMarkedForDeleted += m.groups[m.activeGroup].FileSize
				}
				// TODO SHould have a function to move cursor after toggling and use above
				if m.reviewCursor < len(m.groups[m.activeGroup].Files)-1 {
					m.reviewCursor++
				}
			case "enter":
				// TODO Confirm deletions and start Pass 3
				m.state = StatePass3
				m.files = model.GetFilesToDelete(m.groups)
				m.progressStats = ProgressData{} // Reset
				// Clear out and previous msgs in the channel and reuse
				m.progressChan = make(chan tea.Msg, 10)
				return m, tea.Batch(
					waitForProgressUpdateCmd(m.progressChan),
					runPass3Cmd(m.files, m.pipeline, m.progressChan))
			case "s":
				// TODO Finish this, had made a command for it
				timestamp := time.Now().Format("20060102-150405")
				jsonName := fmt.Sprintf("duplicates_%s.json", timestamp)
				csvName := fmt.Sprintf("duplicates_%s.csv", timestamp)

				errJSON := exporter.ToJSON(m.groups, jsonName)
				errCSV := exporter.ToCSV(m.groups, csvName)

				if errJSON != nil || errCSV != nil {
					return m, displayErr("export failed")
				}

				// Set a "flash message" to show success
				m.statusMessage = fmt.Sprintf("Saved to %s and %s", jsonName, csvName)
				return m, clearStatusAfter(3 * time.Second)
			case "q", "ctrl+c", "esc":
				// TODO As a tea.batch?
				m.statusMessage = "Exiting"
				return m, tea.Quit
			}
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.statusMessage = "Exiting"
			return m, tea.Quit
		}

	// --- Pipeline messages ---

	case StartScanMsg:
		m.progressChan = make(chan tea.Msg, 10)
		return m, tea.Batch(
			waitForProgressUpdateCmd(m.progressChan),
			scanFilesCmd(m.drive, m.target, m.progressChan))

	case ScanFinishedMsg:
		m.files = msg.files
		m.state = StatePass1
		m.results.scanCount = len(m.files)
		// Clear out and previous msgs in the channel and reuse
		m.progressChan = make(chan tea.Msg, 10)
		return m, tea.Batch(
			waitForProgressUpdateCmd(m.progressChan),
			runPass1Cmd(m.files, m.pipeline, m.progressChan))

	case Pass1FinishedMsg:
		m.groups = msg.groups
		m.state = StatePass2
		m.results.pass1Count = len(m.groups)
		m.progressStats = ProgressData{} // Reset
		// Clear out and previous msgs in the channel and reuse
		m.progressChan = make(chan tea.Msg, 10)
		return m, tea.Batch(
			waitForProgressUpdateCmd(m.progressChan),
			runPass2Cmd(m.groups, m.pipeline, m.progressChan))

	case Pass2FinishedMsg:
		m.groups = msg.groups
		m.state = StateReview
		m.results.pass2Count = len(m.groups)
		m.progressStats = ProgressData{} // Reset
		// Clear out and previous msgs in the channel and reuse
		m.progressChan = make(chan tea.Msg, 10)
		return m, waitForProgressUpdateCmd(m.progressChan)

	case Pass3FinishedMsg:
		// TODO Update stats - current gettings ints, but should return a list of non deleted files
		m.results.deletedCount = msg.numFilesDeleted
		// m.groups = msg.groups
		m.state = StateDone
		return m, nil // No more commands, just render final view

	// --- UI Messages ---

	case spinner.TickMsg:
		if m.state != StateDone && m.state != StateError {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progressBar.Update(msg)
		m.progressBar = progressModel.(progress.Model)
		return m, cmd

	case ErrMsg:
		m.err = msg.err
		m.state = StateError
		return m, tea.Quit

	case ScanProgressMsg:
		m.progressStats.NumProcessedFiles = int(msg)
		return m, waitForProgressUpdateCmd(m.progressChan)

	case ProgressMsg:
		m.progressStats = ProgressData(msg)
		return m, tea.Batch(waitForProgressUpdateCmd(m.progressChan), m.progressBar.SetPercent(m.progressStats.OverallPercent))

	case UpdateStatusMsg:
		// Set a "flash message"
		m.statusMessage = msg.status

		// If no duration, keep until new update
		if msg.duration == 0 {
			return m, nil
		}
		return m, clearStatusAfter(time.Duration(msg.duration) * time.Second)

	case ClearStatusMsg:
		// Clear the text
		m.statusMessage = ""
		return m, nil
	}

	return m, nil
}
