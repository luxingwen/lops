package quicnet

import (
	"time"
)

type ScriptTask struct {
	TaskID       string
	Type         string
	Content      string
	Params       []string
	Timeout      time.Duration
	Interpreter  string
	Stdin        string
	Created      time.Time
	Updated      time.Time
	Status       TaskStatus
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
	// ...
	return nil
}


func (st *ScriptTask) SetStatus(status TaskStatus) {
	st.Status = status
	st.Updated = time.Now()
}