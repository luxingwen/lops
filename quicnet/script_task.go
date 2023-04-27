package quicnet

import (
	"context"
	"time"
)

type ScriptTask struct {
	TaskID          string
	Type            string
	Content         string
	InterpreterArgs []string
	Params          []string
	Timeout         time.Duration
	Interpreter     string
	Stdin           string
	Created         time.Time
	Updated         time.Time
	Status          TaskStatus
	Suffix          string
	Cancel          context.CancelFunc
	ScriptResult    *ScriptResult
}

type ScriptErrorCode string

const (
	CodeCreateTempFileFailed ScriptErrorCode = "CREATE_TEMP_FILE_FAILED"
	CodeWriteTempFileFailed  ScriptErrorCode = "WRITE_TEMP_FILE_FAILED"
	CodeCloseTempFileFailed  ScriptErrorCode = "CLOSE_TEMP_FILE_FAILED"
	CodeChmodTempFileFailed  ScriptErrorCode = "CHMOD_TEMP_FILE_FAILED"
	CodeTimeout              ScriptErrorCode = "TIMEOUT"
	CodeStopped              ScriptErrorCode = "STOPPED"
	CodeSuccess              ScriptErrorCode = "SUCCESS"
)

type ScriptResult struct {
	Code      ScriptErrorCode
	Stdout    chan string
	Stderr    chan string
	Error     string
	ExitCode  int
	StartTime time.Time
	EndTime   time.Time
}

func NewScriptTask(request *ScriptTaskRequest) *ScriptTask {
	return &ScriptTask{
		TaskID:      request.TaskID,
		Type:        request.Type,
		Content:     request.Content,
		Interpreter: request.Interpreter,
		Stdin:       request.Stdin,
		Status:      TaskStatusCreated,
		Created:     time.Now(),
		Updated:     time.Now(),
	}
}

func (st *ScriptTask) GetType() string {
	return st.Type
}

func (st *ScriptTask) GetStatus() TaskStatus {
	return st.Status
}

func (st *ScriptTask) GetContent() []byte {
	return []byte(st.Content)
}

func (st *ScriptTask) Run() error {
	// 实现运行脚本的逻辑
	// ...
	return nil
}

func (st *ScriptTask) Stop() error {
	// 实现停止脚本的逻辑
	st.Cancel()
	return nil
}

func (st *ScriptTask) SetStatus(status TaskStatus) {
	st.Status = status
	st.Updated = time.Now()
}
