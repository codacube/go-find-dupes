package tui

import (
	"github.com/codacube/go-find-dupes/internal/model"
	"github.com/codacube/go-find-dupes/internal/processor"
	"github.com/codacube/go-find-dupes/internal/storage"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Contains the TUI state, setup and init, etc

// Model can contain helper methods, but only pure data mutations
// eg. func (m *Model) ToggleSelection(index int) - changes the
//   internal state of the model

// Tracks the current state of the TUI
type State int

const (
	StateScanning State = iota
	StatePass1
	StatePass2
	StatePass3
	StateReview
	StateDone
	StateError
)

type Model struct {
	// Job config
	target   string        // The folder path or ID
	drive    storage.Drive // The interface (Local or GDrive)
	pipeline processor.Pipeline

	// TUI state
	state        State
	err          error // Store errors if they happen
	activeGroup  int
	reviewCursor int

	// Data
	files  []model.File
	groups []model.FileGroup

	// UI elements
	spinner     spinner.Model
	progressBar progress.Model

	// Channel for streaming progress updates - use the same channel for all
	// states. It gets re-made after each state change to clear old
	// msgs (garbage collected)
	progressChan chan tea.Msg

	// Progress info
	// TODO Is this needed? or correct? just use a struct, like below?
	markedForDeletion          int   // A simple count for now - used for display only
	totalBytesMarkedForDeleted int64 // TODO unit64?

	// Stats per pass, reset after each phase
	progressStats ProgressData

	// Final stats per pass/phase
	results struct {
		scanCount    int // Num files found
		pass1Count   int // Num grouped by file size (possible dupes)
		pass2Count   int // Num grouped by hash (duplicates)
		deletedCount int // numDeleted
	}

	statusMessage string // For flash messages
}

// TODO These could be separated into deleted, hashed?
type ProgressData struct {
	// Overall progress
	NumProcessedFiles int
	TotalFiles        int
	OverallPercent    float64 // TODO Could just be calculated in the TUI

	// File specific
	CurrentFilename string
	FilePercent     float64 // 0.0 to 1.0

	DeletedCount int
}

func NewModel(drive storage.Drive, pipeline processor.Pipeline, target string) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	pg := progress.New(progress.WithDefaultGradient())
	pg.Width = 20

	return Model{
		drive:       drive,
		pipeline:    pipeline,
		target:      target,
		state:       StateScanning,
		spinner:     sp,
		progressBar: pg,
	}
}

func (m Model) Init() tea.Cmd {
	// Start the spinner and scanning files
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			return StartScanMsg{}
		},
	)
}

// TODO is target used?
func Start(drive storage.Drive, pipeline processor.Pipeline, target string) error {
	model := NewModel(drive, pipeline, target)

	// Log file - usage: log.Println("I'm being logged!")
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		return err
	}
	defer f.Close()

	p := tea.NewProgram(model)

	_, err = p.Run()
	return err
}
