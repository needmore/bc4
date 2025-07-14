package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// PagerOptions contains options for the pager
type PagerOptions struct {
	// Pager command to use (e.g., "less", "more")
	Pager string
	// Whether to force pager even if not a TTY
	Force bool
	// Whether to disable pager
	NoPager bool
}

// ShowInPager displays content using the configured pager
func ShowInPager(content string, opts *PagerOptions) error {
	if opts == nil {
		opts = &PagerOptions{}
	}

	// Check if pager should be disabled
	if opts.NoPager {
		fmt.Print(content)
		return nil
	}

	// Check if output is to a TTY (unless forced)
	if !opts.Force && !isTerminal(os.Stdout) {
		fmt.Print(content)
		return nil
	}

	// Determine pager command
	pager := opts.Pager
	if pager == "" {
		// Check PAGER environment variable
		pager = os.Getenv("PAGER")
		if pager == "" {
			pager = "less"
		}
	}

	// Try to run the pager
	cmd := exec.Command("sh", "-c", pager)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set LESS options if using less and not already set
	if strings.Contains(pager, "less") && os.Getenv("LESS") == "" {
		cmd.Env = append(os.Environ(), "LESS=FRX")
	}

	err := cmd.Run()
	if err != nil {
		// Fallback to direct output if pager fails
		fmt.Print(content)
		// Don't return the error since we successfully displayed the content
		return nil
	}

	return nil
}

// WriteToPager creates a writer that will display content through a pager
func WriteToPager(opts *PagerOptions) (io.WriteCloser, error) {
	if opts == nil {
		opts = &PagerOptions{}
	}

	// Check if pager should be disabled
	if opts.NoPager {
		return &passthroughWriter{w: os.Stdout}, nil
	}

	// Check if output is to a TTY (unless forced)
	if !opts.Force && !isTerminal(os.Stdout) {
		return &passthroughWriter{w: os.Stdout}, nil
	}

	// Determine pager command
	pager := opts.Pager
	if pager == "" {
		// Check PAGER environment variable
		pager = os.Getenv("PAGER")
		if pager == "" {
			pager = "less"
		}
	}

	// Create a buffer to collect output
	buf := &bytes.Buffer{}

	return &pagerWriter{
		buf:   buf,
		pager: pager,
	}, nil
}

// pagerWriter collects output and displays it through a pager when closed
type pagerWriter struct {
	buf   *bytes.Buffer
	pager string
}

func (p *pagerWriter) Write(data []byte) (int, error) {
	return p.buf.Write(data)
}

func (p *pagerWriter) Close() error {
	content := p.buf.String()

	// Try to run the pager
	cmd := exec.Command("sh", "-c", p.pager)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set LESS options if using less and not already set
	if strings.Contains(p.pager, "less") && os.Getenv("LESS") == "" {
		cmd.Env = append(os.Environ(), "LESS=FRX")
	}

	err := cmd.Run()
	if err != nil {
		// Fallback to direct output if pager fails
		fmt.Print(content)
		// Don't return the error since we successfully displayed the content
		return nil
	}

	return nil
}

// passthroughWriter is a simple writer that passes through to another writer
type passthroughWriter struct {
	w io.Writer
}

func (p *passthroughWriter) Write(data []byte) (int, error) {
	return p.w.Write(data)
}

func (p *passthroughWriter) Close() error {
	return nil
}

// isTerminal checks if the given file is a terminal
func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}

	// Get file info
	fi, err := f.Stat()
	if err != nil {
		return false
	}

	// Check if it's a character device (terminal)
	return (fi.Mode() & os.ModeCharDevice) != 0
}
