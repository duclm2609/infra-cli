package util

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Spinner provides a simple terminal spinner for long-running operations
type Spinner struct {
	message  string
	frames   []string
	interval time.Duration
	writer   io.Writer
	stop     chan struct{}
	done     chan struct{}
	mu       sync.Mutex
	running  bool
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message:  message,
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		interval: 100 * time.Millisecond,
		writer:   os.Stdout,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// SetWriter sets the output writer
func (s *Spinner) SetWriter(w io.Writer) {
	s.writer = w
}

// SetMessage updates the spinner message
func (s *Spinner) SetMessage(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stop = make(chan struct{})
	s.done = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.done)
		frameIdx := 0
		for {
			select {
			case <-s.stop:
				// Clear the line
				fmt.Fprintf(s.writer, "\r\033[K")
				return
			default:
				s.mu.Lock()
				msg := s.message
				s.mu.Unlock()

				fmt.Fprintf(s.writer, "\r%s %s", s.frames[frameIdx], msg)
				frameIdx = (frameIdx + 1) % len(s.frames)
				time.Sleep(s.interval)
			}
		}
	}()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stop)
	<-s.done
}

// Success stops the spinner and shows a success message
func (s *Spinner) Success(message string) {
	s.Stop()
	fmt.Fprintf(s.writer, "\r✓ %s\n", message)
}

// Fail stops the spinner and shows a failure message
func (s *Spinner) Fail(message string) {
	s.Stop()
	fmt.Fprintf(s.writer, "\r✗ %s\n", message)
}

// Info stops the spinner and shows an info message
func (s *Spinner) Info(message string) {
	s.Stop()
	fmt.Fprintf(s.writer, "\rℹ %s\n", message)
}

// Progress represents a progress indicator
type Progress struct {
	total   int
	current int
	message string
	writer  io.Writer
	mu      sync.Mutex
}

// NewProgress creates a new progress indicator
func NewProgress(total int, message string) *Progress {
	return &Progress{
		total:   total,
		current: 0,
		message: message,
		writer:  os.Stdout,
	}
}

// SetWriter sets the output writer
func (p *Progress) SetWriter(w io.Writer) {
	p.writer = w
}

// Increment increments the progress by 1
func (p *Progress) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current++
	p.render()
}

// Set sets the current progress value
func (p *Progress) Set(value int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = value
	p.render()
}

// SetMessage updates the progress message
func (p *Progress) SetMessage(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.message = message
	p.render()
}

// render displays the current progress
func (p *Progress) render() {
	percent := 0
	if p.total > 0 {
		percent = (p.current * 100) / p.total
	}
	fmt.Fprintf(p.writer, "\r[%3d%%] %s (%d/%d)", percent, p.message, p.current, p.total)
}

// Done completes the progress
func (p *Progress) Done() {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(p.writer, "\r\033[K✓ %s (completed)\n", p.message)
}

// StatusPrinter provides simple status messages
type StatusPrinter struct {
	writer io.Writer
	quiet  bool
}

// NewStatusPrinter creates a new status printer
func NewStatusPrinter(quiet bool) *StatusPrinter {
	return &StatusPrinter{
		writer: os.Stdout,
		quiet:  quiet,
	}
}

// SetWriter sets the output writer
func (sp *StatusPrinter) SetWriter(w io.Writer) {
	sp.writer = w
}

// Print prints a status message (respects quiet mode)
func (sp *StatusPrinter) Print(message string) {
	if !sp.quiet {
		fmt.Fprintln(sp.writer, message)
	}
}

// Printf prints a formatted status message (respects quiet mode)
func (sp *StatusPrinter) Printf(format string, args ...interface{}) {
	if !sp.quiet {
		fmt.Fprintf(sp.writer, format, args...)
	}
}

// Success prints a success message (always shown)
func (sp *StatusPrinter) Success(message string) {
	fmt.Fprintf(sp.writer, "✓ %s\n", message)
}

// Error prints an error message (always shown)
func (sp *StatusPrinter) Error(message string) {
	fmt.Fprintf(os.Stderr, "✗ %s\n", message)
}

// Warning prints a warning message (respects quiet mode)
func (sp *StatusPrinter) Warning(message string) {
	if !sp.quiet {
		fmt.Fprintf(sp.writer, "⚠ %s\n", message)
	}
}

// Info prints an info message (respects quiet mode)
func (sp *StatusPrinter) Info(message string) {
	if !sp.quiet {
		fmt.Fprintf(sp.writer, "ℹ %s\n", message)
	}
}
