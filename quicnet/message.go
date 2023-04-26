package quicnet

import (
	"encoding/json"
)

type Message struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data []byte `json:"data"`
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func UnmarshalMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

type ScriptTaskRequest struct {
	TaskID       string            `json:"task_id"`
	Type         string            `json:"type"`
	Content      string            `json:"content"`
	Params       map[string]string `json:"params"`
	Timeout      int               `json:"timeout"`
	Interpreter  string            `json:"interpreter"`
	Stdin        string            `json:"stdin"`
}

