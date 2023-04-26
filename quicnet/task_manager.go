package quicnet

import (
	"time"

	"github.com/patrickmn/go-cache"
)



type TaskManager struct {
	tasks *cache.Cache
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: cache.New(cache.NoExpiration, 10*time.Minute),
	}
}

func (tm *TaskManager) AddTask(task Task) {
	tm.tasks.Set(task.GetTaskID(), task, cache.DefaultExpiration)
}

func (tm *TaskManager) UpdateTaskStatus(taskID string, status TaskStatus) {
	if task, found := tm.tasks.Get(taskID); found {
		task.(Task).SetStatus(status)
		tm.tasks.Set(taskID, task, cache.DefaultExpiration)
	}
}

func (tm *TaskManager) GetTask(taskID string) Task {
	if task, found := tm.tasks.Get(taskID); found {
		return task.(Task)
	}
	return nil
}

func (tm *TaskManager) RemoveTask(taskID string) {
	tm.tasks.Delete(taskID)
}