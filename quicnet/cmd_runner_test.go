package quicnet

import (
	"fmt"
	"testing"
	"time"
)

func TestRunScript(t *testing.T) {
	cmdRunner := NewCmdRunner()

	script := `echo "Hello, World!"`
	reqTask := &ScriptTask{
		TaskID:       "testTask",
		Interpreter:  "", // 使用默认解释器
		Content:      script,
		Timeout:      time.Second * 5,
		ScriptResult: &ScriptResult{},
	}

	// 运行脚本
	cmdRunner.RunScript(reqTask)

	// 检查输出
	if reqTask.ScriptResult.Code != CodeSuccess {
		t.Errorf("Expected script to succeed, but got error code %s with error message: %s", reqTask.ScriptResult.Code, reqTask.ScriptResult.Error)
	}

	expectedStdout := "Hello, World!\n"
	if reqTask.ScriptResult.Stdout != expectedStdout {
		t.Errorf("Expected stdout to be %q, but got %q", expectedStdout, reqTask.ScriptResult.Stdout)
	}

	if reqTask.ScriptResult.Stderr != "" {
		t.Errorf("Expected stderr to be empty, but got %q", reqTask.ScriptResult.Stderr)
	}

	fmt.Printf("stdout: %s\n", reqTask.ScriptResult.Stdout)
	fmt.Printf("stderr: %s\n", reqTask.ScriptResult.Stderr)
}
