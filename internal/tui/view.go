package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Presentation layer functions, functions that generate the string that gets printed to the terminal
// Only reads the model and formats it for display, doesn't change any state

var deleteFileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)           // Red
var deleteFileStyleSecondary = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(false) // Red
var keepFileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)            // Green

func (m Model) View() string {
	if m.state == StateError {
		return fmt.Sprintf("\n❌ Error: %v\n\nPress q to quit.\n", m.err)
	}

	s := "\n 🔍 Duplicate File Remover\n\n"

	// TODO If no dupes, it must crash here somewhere? Msg2Complete is received OK, but crashes on review stage
	// Render each stage
	s += m.renderStage("Finding files", StateScanning,
		fmt.Sprintf("%d files found", m.results.scanCount))

	s += m.renderStage("Pass 1: Group by file size", StatePass1,
		fmt.Sprintf("%d groups", m.results.pass1Count))

	s += m.renderStage("Pass 2: Checking hashes", StatePass2,
		fmt.Sprintf("%d duplicate sets confirmed", m.results.pass2Count))

	if m.state == StatePass3 {
		s += m.renderStage("Pass 3: Deletion", StatePass3,
			fmt.Sprintf("%d files removed (%s recovered)", m.results.deletedCount, FormatSize(m.totalBytesMarkedForDeleted)))
	} else {
		s += m.renderStage("Pass 3: Deletion Review", StateReview,
			fmt.Sprintf("%d files marked (%s recovered)", m.markedForDeletion, FormatSize(m.totalBytesMarkedForDeleted)))
	}

	// TODO Make this a function
	if m.state == StateReview {
		if len(m.groups) == 0 {
			s += "\nNo duplicate files found!\n"
			s += "Press q to exit.\n"
			return s
		}

		keepText := keepFileStyle.Render("✔ KEEP")
		deleteText := deleteFileStyle.Render("✖ DELETE")

		s += "\n 📝 Review Duplicates:\n\n"
		s += "    [↑/↓] Navigate Files\n"
		s += "    [Space] Mark file " + keepText + " or " + deleteText + "\n"
		s += "    [←/→/Tab] Switch Groups\n"
		s += "    [Enter] Delete All marked\n"
		s += "    [s] Save as json/csv\n"
		s += "    [q] To Quit\n\n"

		s += fmt.Sprintf("Group %d of %d (%s) - %s\n", m.activeGroup+1, len(m.groups), FormatSize(m.groups[m.activeGroup].FileSize), m.groups[m.activeGroup].Identifier)

		// Get the maxLen of the filenames for formatting, and aligning the DisplayPath column
		var maxLen int
		for _, file := range m.groups[m.activeGroup].Files {
			if len(file.Name) > maxLen {
				maxLen = len(file.Name)
			}
		}

		// Render the visible window of groups (simple scrolling)
		for i, file := range m.groups[m.activeGroup].Files {
			cursor := " "
			if m.reviewCursor == i {
				cursor = ">"
			}

			actionText := ""
			fileText := ""
			if file.MarkForDeletion {
				actionText = deleteFileStyle.Render(fmt.Sprintf("✖ DELETE %-*s", maxLen, file.Name))
				fileText = deleteFileStyleSecondary.Render(file.DisplayPath)
			} else {
				actionText = keepFileStyle.Render(fmt.Sprintf("✔ KEEP   %-*s", maxLen, file.Name))
				fileText = file.DisplayPath
			}
			s += fmt.Sprintf("%s  %s (%s)\n", cursor, actionText, fileText)
		}
		s += "\n"
	}

	if m.statusMessage != "" {
		s += fmt.Sprintf("\n\n\033[32m[INFO] %s\033[0m", m.statusMessage)
	}

	if m.state == StateDone {
		s += fmt.Sprintf("\n✨ Operation Complete! %d files deleted\n\n Press q to exit.\n", m.results.deletedCount)
	}

	return s
}

// TODO Review needs it's own line
// TODO Seperate into their own functions
// renderStage handles the logic: Waiting -> Running -> Done
func (m Model) renderStage(label string, stageState State, resultText string) string {
	// 1. Future Stage (Pending)
	if m.state < stageState {
		return fmt.Sprintf("  ○  %s\n", label)
	}

	// 2. Current Stage (Running) - Show Spinner & Progress
	if m.state == stageState {
		if m.state == StateScanning {
			return fmt.Sprintf("  %s %s: %d files found\n", m.spinner.View(), label, m.progressStats.NumProcessedFiles)
		}
		if m.state == StateReview {
			return fmt.Sprintf("  %s %s: %d files marked (%s recovered)\n", m.spinner.View(), label, m.markedForDeletion, FormatSize(m.totalBytesMarkedForDeleted))
		}

		// TODO Pretty sure bubbletea has it's own progress bar
		// Calculate percentage for the bar
		bar := m.progressBar.View()
		// percent := m.progress.OverallPercent
		// barWidth := 20
		// filled := int(percent * float64(barWidth))
		// bar := "[" + repeat("=", filled) + repeat("-", barWidth-filled) + "]"

		// If current state is Pass2, show current file being hashed
		fileProgress := fmt.Sprintf(" %d of %d", m.progressStats.NumProcessedFiles, m.progressStats.TotalFiles)
		if stageState == StatePass2 && m.progressStats.CurrentFilename != "" {
			fileProgress += fmt.Sprintf(" Hashing: %s (%.1f%%)", m.progressStats.CurrentFilename, m.progressStats.FilePercent*100)
		}

		// if stageState == StateReview {
		// 	fileProgress = fmt.Sprintf(" (%d files marked)", m.markedForDeletion)
		// }

		return fmt.Sprintf("  %s %s: %s %s\n", m.spinner.View(), label, bar, fileProgress)
	}

	// TODO Use Lipgloss instead of the ANSI chars
	return fmt.Sprintf("  ✅ %s: \033[1m%s\033[0m\n", label, resultText)
}

// --- Helper functions ---

// TODO Lipgloss handles this type of thing, use that instead
func repeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	res := ""
	for i := 0; i < count; i++ {
		res += s
	}
	return res
}

// General formatting function, belongs in the TUI
func FormatSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	unitIndex := 0

	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}
	return fmt.Sprintf("%d %s", size, units[unitIndex])
}
