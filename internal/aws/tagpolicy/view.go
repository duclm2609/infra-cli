// Package tagpolicy provides functionality for retrieving and displaying
// AWS Organizations tag policies through an interactive TUI.
package tagpolicy

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ViewModel represents the view state for the TUI display.
// It tracks which tag keys are available, which one is selected,
// and which ones are expanded to show their allowed values.
type ViewModel struct {
	// TagKeys contains all tag keys from the effective tag policy.
	TagKeys []TagKey

	// SelectedIndex tracks which tag key is currently selected.
	// Valid range is 0 to len(TagKeys)-1.
	SelectedIndex int

	// ExpandedKeys tracks which tag keys are expanded to show their values.
	// The key is the index of the tag key in TagKeys slice.
	// If an index is present and true, that tag key is expanded.
	ExpandedKeys map[int]bool
}

// TagPolicyView handles the interactive TUI display for tag policies.
// It manages the view model and provides methods for rendering and
// handling user input.
type TagPolicyView struct {
	model *ViewModel
}

// NewViewModel creates a new ViewModel initialized from a TagPolicy.
// The SelectedIndex is set to 0 (first tag key selected) and all
// tag keys start in collapsed state.
func NewViewModel(policy *TagPolicy) *ViewModel {
	if policy == nil {
		return &ViewModel{
			TagKeys:       []TagKey{},
			SelectedIndex: 0,
			ExpandedKeys:  make(map[int]bool),
		}
	}

	return &ViewModel{
		TagKeys:       policy.Tags,
		SelectedIndex: 0,
		ExpandedKeys:  make(map[int]bool),
	}
}

// NewTagPolicyView creates a new TUI view from a TagPolicy.
// It initializes the view model with the policy's tag keys,
// sets the first tag key as selected, and all keys as collapsed.
func NewTagPolicyView(policy *TagPolicy) *TagPolicyView {
	return &TagPolicyView{
		model: NewViewModel(policy),
	}
}

// Model returns the underlying ViewModel for the view.
// This is useful for testing and inspecting the view state.
func (v *TagPolicyView) Model() *ViewModel {
	return v.model
}

// IsExpanded returns whether the tag key at the given index is expanded.
func (vm *ViewModel) IsExpanded(index int) bool {
	return vm.ExpandedKeys[index]
}

// SetExpanded sets the expanded state for the tag key at the given index.
func (vm *ViewModel) SetExpanded(index int, expanded bool) {
	if expanded {
		vm.ExpandedKeys[index] = true
	} else {
		delete(vm.ExpandedKeys, index)
	}
}

// ToggleExpanded toggles the expanded state for the tag key at the given index.
func (vm *ViewModel) ToggleExpanded(index int) {
	if vm.ExpandedKeys[index] {
		delete(vm.ExpandedKeys, index)
	} else {
		vm.ExpandedKeys[index] = true
	}
}

// TagKeyCount returns the number of tag keys in the view model.
func (vm *ViewModel) TagKeyCount() int {
	return len(vm.TagKeys)
}

// Key constants for keyboard navigation.
// Arrow keys are escape sequences in terminal raw mode (ESC [ A/B/C/D).
// We also support vim-style navigation keys as alternatives.
const (
	// KeyUp represents vim-style up ('k')
	KeyUp = 'k'
	// KeyDown represents vim-style down ('j')
	KeyDown = 'j'
	// KeyLeft represents vim-style left ('h')
	KeyLeft = 'h'
	// KeyRight represents vim-style right ('l')
	KeyRight = 'l'
	// KeyEnter represents the Enter key
	KeyEnter = '\r'
	// KeyNewline represents the newline character (also Enter on some systems)
	KeyNewline = '\n'
	// KeySpace represents the Space key
	KeySpace = ' '
	// KeyQuit represents the quit key
	KeyQuit = 'q'
	// KeyEscape represents the escape key (start of arrow key sequences)
	KeyEscape = 27
)

// Arrow key escape sequence codes (the third byte after ESC [)
const (
	ArrowUp    = 'A'
	ArrowDown  = 'B'
	ArrowRight = 'C'
	ArrowLeft  = 'D'
)

// HandleKeyPress processes keyboard input and updates the view state accordingly.
// It returns true if the user wants to quit (pressed 'q'), false otherwise.
//
// Supported keys:
//   - Up arrow / 'k': Move selection up (bounded at 0)
//   - Down arrow / 'j': Move selection down (bounded at len-1)
//   - Right arrow / 'l' / Enter / Space: Expand selected tag key
//   - Left arrow / 'h': Collapse selected tag key
//   - 'q': Quit the TUI
//
// Requirements: 3.2, 3.3, 3.5
func (v *TagPolicyView) HandleKeyPress(key rune) (quit bool) {
	switch key {
	case KeyUp:
		// Move selection up, bounded at 0
		v.moveUp()
	case KeyDown:
		// Move selection down, bounded at len-1
		v.moveDown()
	case KeyRight, KeyEnter, KeyNewline, KeySpace:
		// Expand the currently selected tag key
		v.expand()
	case KeyLeft:
		// Collapse the currently selected tag key
		v.collapse()
	case KeyQuit:
		// User wants to quit
		return true
	}
	return false
}

// HandleArrowKey processes arrow key input (the direction byte after ESC [).
func (v *TagPolicyView) HandleArrowKey(direction byte) (quit bool) {
	switch direction {
	case ArrowUp:
		v.moveUp()
	case ArrowDown:
		v.moveDown()
	case ArrowRight:
		v.expand()
	case ArrowLeft:
		v.collapse()
	}
	return false
}

// moveUp moves the selection up, bounded at 0.
func (v *TagPolicyView) moveUp() {
	if v.model.SelectedIndex > 0 {
		v.model.SelectedIndex--
	}
}

// moveDown moves the selection down, bounded at len-1.
func (v *TagPolicyView) moveDown() {
	if v.model.SelectedIndex < v.model.TagKeyCount()-1 {
		v.model.SelectedIndex++
	}
}

// expand expands the currently selected tag key.
func (v *TagPolicyView) expand() {
	if v.model.TagKeyCount() > 0 {
		v.model.SetExpanded(v.model.SelectedIndex, true)
	}
}

// collapse collapses the currently selected tag key.
func (v *TagPolicyView) collapse() {
	if v.model.TagKeyCount() > 0 {
		v.model.SetExpanded(v.model.SelectedIndex, false)
	}
}

// Visual indicators for the TUI display
const (
	// SelectionIndicator shows which tag key is currently selected
	SelectionIndicator = ">"
	// NoSelectionIndicator is used for non-selected tag keys (same width as SelectionIndicator)
	NoSelectionIndicator = " "
	// ExpandedIndicator shows that a tag key is expanded (values visible)
	ExpandedIndicator = "▼"
	// CollapsedIndicator shows that a tag key is collapsed (values hidden)
	CollapsedIndicator = "▶"
	// ValueIndent is the indentation for values under an expanded tag key
	ValueIndent = "    "
)

// Render draws the current view state to the terminal.
// It displays all tag keys with selection and expansion indicators,
// and shows values indented under expanded tag keys.
//
// Output format:
//
//	AWS Tag Policy
//
//	> ▼ Environment
//	      Production
//	      Development
//	      Staging
//	  ▶ CostCenter
//	  ▶ Project
//
//	↑/↓: navigate, →: expand, ←: collapse, q: quit
//
// Requirements: 3.1, 3.4, 3.6, 3.7
func (v *TagPolicyView) Render() string {
	var sb strings.Builder

	// In raw terminal mode, \n only moves down without returning to column 0.
	// We need to use \r\n to properly start each line at the left edge.
	newline := "\r\n"

	// Header
	sb.WriteString("AWS Tag Policy")
	sb.WriteString(newline)
	sb.WriteString(newline)

	// Handle empty policy
	if v.model.TagKeyCount() == 0 {
		sb.WriteString("No tag keys found.")
		sb.WriteString(newline)
		sb.WriteString(newline)
		sb.WriteString("q: quit")
		sb.WriteString(newline)
		return sb.String()
	}

	// Render each tag key
	for i, tagKey := range v.model.TagKeys {
		isSelected := i == v.model.SelectedIndex

		// Selection indicator
		if isSelected {
			sb.WriteString(SelectionIndicator)
		} else {
			sb.WriteString(NoSelectionIndicator)
		}
		sb.WriteString(" ")

		// Expansion indicator
		if v.model.IsExpanded(i) {
			sb.WriteString(ExpandedIndicator)
		} else {
			sb.WriteString(CollapsedIndicator)
		}
		sb.WriteString(" ")

		// Tag key name (yellow if selected)
		if isSelected {
			sb.WriteString(ColorYellow)
		}
		sb.WriteString(tagKey.Name)
		if isSelected {
			sb.WriteString(ColorReset)
		}
		sb.WriteString(newline)

		// If expanded, show values indented
		if v.model.IsExpanded(i) {
			for _, value := range tagKey.Values {
				sb.WriteString(ValueIndent)
				sb.WriteString(value)
				sb.WriteString(newline)
			}
		}
	}

	// Help text
	sb.WriteString(newline)
	sb.WriteString("↑/↓: navigate, →: expand, ←: collapse, q: quit")
	sb.WriteString(newline)

	return sb.String()
}

// ANSI escape codes for terminal control
const (
	// ClearScreen clears the entire screen and moves cursor to home position
	ClearScreen = "\033[H\033[2J"
	// ColorYellow sets text color to yellow
	ColorYellow = "\033[33m"
	// ColorReset resets text color to default
	ColorReset = "\033[0m"
)

// Run starts the interactive TUI loop.
// It sets up terminal raw mode for keyboard input, renders the view,
// reads key presses, and handles them until the user quits.
// The terminal state is always restored on exit, even if an error occurs.
//
// Requirements: 3.1, 3.5
func (v *TagPolicyView) Run() error {
	// Get the file descriptor for stdin
	fd := int(os.Stdin.Fd())

	// Save the current terminal state and set raw mode
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to set terminal raw mode: %w", err)
	}
	// Always restore terminal state on exit
	defer term.Restore(fd, oldState)

	// Main loop
	for {
		// Clear screen and render
		fmt.Print(ClearScreen)
		fmt.Print(v.Render())

		// Read input - may be single byte or escape sequence
		buf := make([]byte, 3)
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		// Check for escape sequence (arrow keys)
		if n >= 3 && buf[0] == KeyEscape && buf[1] == '[' {
			// Arrow key: ESC [ A/B/C/D
			if v.HandleArrowKey(buf[2]) {
				break // quit
			}
		} else if n >= 1 {
			// Single key press
			if v.HandleKeyPress(rune(buf[0])) {
				break // quit
			}
		}
	}

	// Clear screen before exiting
	fmt.Print(ClearScreen)
	return nil
}
