package quicnet

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type CmdRunner struct{}

func NewCmdRunner() *CmdRunner {
	return &CmdRunner{}
}

func (cr *CmdRunner) RunScript(reqtask *ScriptTask) {

	r := reqtask.ScriptResult

	// 关闭输出通道

	if len(reqtask.Interpreter) == 0 {
		reqtask.Interpreter = defaultInterpreter
	}
	if len(reqtask.InterpreterArgs) == 0 {
		reqtask.InterpreterArgs = []string{defaultInterpreterArg}
	}
	if reqtask.Suffix == "" {
		reqtask.Suffix = defaultScriptSuffix
	}

	tmpfile, err := ioutil.TempFile("", reqtask.TaskID+reqtask.Suffix)
	if err != nil {
		r.Error = err.Error()
		r.Code = CodeCreateTempFileFailed
		reqtask.ScriptResult = r
		return
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(reqtask.Content)); err != nil {
		r.Error = err.Error()
		r.Code = CodeWriteTempFileFailed
		reqtask.ScriptResult = r
		return
	}
	if err := tmpfile.Close(); err != nil {
		r.Error = err.Error()
		r.Code = CodeCloseTempFileFailed
		reqtask.ScriptResult = r
		return
	}

	err = os.Chmod(tmpfile.Name(), 0755)
	if err != nil {
		r.Error = err.Error()
		r.Code = CodeChmodTempFileFailed
		reqtask.ScriptResult = r
		return
	}

	args := append(reqtask.InterpreterArgs, tmpfile.Name())
	args = append(args, reqtask.Params...)

	ctx, cancel := context.WithTimeout(context.Background(), reqtask.Timeout)
	defer cancel()
	reqtask.Cancel = cancel

	cmd := exec.CommandContext(ctx, reqtask.Interpreter, args...)

	if len(reqtask.Stdin) > 0 {
		cmd.Stdin = bytes.NewBufferString(reqtask.Stdin)
	}

	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()

	cmd.Stdout = stdoutPipeWriter
	cmd.Stderr = stderrPipeWriter

	var stdout, stderr strings.Builder
	stdoutDone := make(chan struct{})
	stderrDone := make(chan struct{})

	go func() {
		defer close(stdoutDone)
		defer stdoutPipeReader.Close()

		reader := bufio.NewReader(stdoutPipeReader)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			reqtask.ScriptResult.Stdout += line
		}
	}()

	go func() {
		defer close(stderrDone)
		defer stderrPipeReader.Close()

		reader := bufio.NewReader(stderrPipeReader)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			reqtask.ScriptResult.Stderr += line
		}
	}()
	startTime := time.Now()
	err = cmd.Start()
	if err != nil {
		return
	}

	err = cmd.Wait()
	endTime := time.Now()

	stdoutPipeWriter.Close()
	stderrPipeWriter.Close()

	<-stdoutDone
	<-stderrDone

	var exitCode int
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	if errors.Is(err, context.DeadlineExceeded) {
		r.Error = "script execution timeout"
		r.Code = CodeTimeout

		reqtask.ScriptResult = r
		return
	}

	// 添加错误信息到 ScriptResult
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
		fmt.Println("err:", errorMsg)
	}

	fmt.Println(stdout.String())
	fmt.Println(stderr.String())

	r.Code = CodeSuccess
	r.EndTime = endTime
	r.StartTime = startTime
	r.ExitCode = exitCode
	r.Error = errorMsg

}
