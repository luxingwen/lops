package quicnet

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRunScript(t *testing.T) {
	cmdRunner := NewCmdRunner()

	script := `echo "Hello, World!"`
	reqTask := &ScriptTask{
		TaskID:      "testTask",
		Interpreter: "", // 使用默认解释器
		Content:     script,
		Timeout:     time.Second * 5,
		ScriptResult: &ScriptResult{
			Stdout: make(chan string, 10),
			Stderr: make(chan string, 10),
		},
	}

	// 运行脚本
	go cmdRunner.RunScript(reqTask)

	var stdoutBuilder strings.Builder
	var stderrBuilder strings.Builder

	// 从输出通道中读取输出
	for s := range reqTask.ScriptResult.Stdout {
		stdoutBuilder.WriteString(s)
	}
	for s := range reqTask.ScriptResult.Stderr {
		stderrBuilder.WriteString(s)
	}

	stdout := stdoutBuilder.String()
	stderr := stderrBuilder.String()

	// 检查输出
	if reqTask.ScriptResult.Code != CodeSuccess {
		t.Errorf("Expected script to succeed, but got error code %s with error message: %s", reqTask.ScriptResult.Code, reqTask.ScriptResult.Error)
	}

	expectedStdout := "Hello, World!\n"
	if stdout != expectedStdout {
		t.Errorf("Expected stdout to be %q, but got %q", expectedStdout, stdout)
	}

	if stderr != "" {
		t.Errorf("Expected stderr to be empty, but got %q", stderr)
	}

	fmt.Printf("stdout: %s\n", stdout)
	fmt.Printf("stderr: %s\n", stderr)
}
