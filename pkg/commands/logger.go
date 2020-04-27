package commands

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func NewLogger(cmd *cobra.Command) *Logger {
	return &Logger{
		Out: cmd.OutOrStdout(),
		Err: cmd.OutOrStderr(),
	}
}

type Logger struct {
	Out io.Writer
	Err io.Writer
}

func (l *Logger) Infof(format string, a ...interface{}) {
	l.Out.Write([]byte(fmt.Sprintf(format+"\n", a...)))
}
