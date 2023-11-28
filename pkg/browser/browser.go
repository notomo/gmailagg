package browser

import (
	"context"
	"io"
	"os"
	"os/exec"
)

type Command struct {
	Executable   string
	stdoutWriter io.Writer
	stderrWriter io.Writer
}

func New(
	stdoutWriter io.Writer,
	stderrWriter io.Writer,
) *Command {
	return &Command{
		Executable:   os.Getenv("BROWSER"),
		stdoutWriter: stdoutWriter,
		stderrWriter: stderrWriter,
	}
}

func (c *Command) Open(ctx context.Context, url string) error {
	cmd := exec.CommandContext(ctx, c.Executable, url)
	cmd.Stdout = c.stdoutWriter
	cmd.Stderr = c.stderrWriter
	if err := cmd.Run(); err != nil {
		return err
	}

	if _, err := c.stdoutWriter.Write([]byte("Authorize in your browser.\n")); err != nil {
		return err
	}

	return nil
}
