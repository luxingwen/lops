package quicnet

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

type CmdRunner struct{}

func NewCmdRunner() *CmdRunner {
	return &CmdRunner{}
}

func (cr *CmdRunner) RunScript(scriptContent string, interpreter string, params []string, timeout time.Duration, stdin string) (string, error) {
	if len(interpreter) == 0 {
		interpreter = "sh"
	}

	script := strings.NewReader(scriptContent)
	cmd := exec.Command(interpreter, params...)
	cmd.Stdin = script

	if len(stdin) > 0 {
		cmd.Stdin = bytes.NewBufferString(stdin)
	}

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd.Start()
	err := cmd.Wait()
	if errors.Is(err, context.DeadlineExceeded) {
		return "", errors.New("script execution timeout")
	}

	return output.String(), err
}
