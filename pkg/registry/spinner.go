// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0	// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"io"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

const framerate = time.Millisecond * 150

var spinners = []string{"|", "/", "-", "\\"}

type uploadSpinner struct {
	size     string
	stopChan chan struct{}
	doneChan chan struct{}
	Output   io.Writer
	NotTty   bool
}

func newUploadSpinner(writer io.Writer, size int64) *uploadSpinner {
	sp := &uploadSpinner{
		size:     readableSize(size),
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		Output:   writer,
		NotTty:   !terminal.IsTerminal(syscall.Stdout),
	}
	return sp
}

func (s *uploadSpinner) Stop() {
	close(s.stopChan)
	<-s.doneChan
}

func (s *uploadSpinner) Write() {
	defer close(s.doneChan)

	if s.NotTty {
		return
	}

	index := 0
	fmt.Fprint(s.Output, "\n")
	for {
		select {
		case <-s.stopChan:
			s.clear()
			return
		case <-time.After(framerate):
			s.clear()

			fmt.Fprintf(s.Output, "\t %s %s", spinners[index], s.size)

			index++
			if index == len(spinners) {
				index = 0
			}
		}
	}
}

func (s *uploadSpinner) clear() {
	fmt.Fprint(s.Output, "\033[2K")
	fmt.Fprint(s.Output, "\n")
	fmt.Fprint(s.Output, "\033[1A")
}

func readableSize(length int64) string {
	const (
		gb = 1000000000
		mb = 1000000
		kb = 1000
	)

	switch {
	case length > gb:
		return fmt.Sprintf("%0.2f GB", float64(length)/gb)
	case length > mb:
		return fmt.Sprintf("%0.2f MB", float64(length)/mb)
	case length > kb:
		return fmt.Sprintf("%0.2f KB", float64(length)/kb)
	default:
		return strconv.FormatInt(length, 10) + " B"
	}
}
