package backgroundtasks

import (
	"time"

	"github.com/toastate/toastainer/internal/db/redisdb"
)

// since both certificates and statistics are stored in the local Redis, we may use it for the lock mechanism
// If it is nit reachable, we are not doing anything anyway
// Eventually replace this mechanism with a Raft like one

type taskLock struct {
	TaskName string
	Deadline time.Time
}

func refreshTaskLock(taskUniqueName string, timeout time.Duration) (*taskLock, error) {
	ok, err := redisdb.LockTryOnce("backgroundtask_"+taskUniqueName, timeout)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &taskLock{
		TaskName: taskUniqueName,
		Deadline: time.Now().Add(timeout - timeout/10),
	}, nil
}

func (l *taskLock) active() bool {
	return time.Now().Before(l.Deadline)
}

func (l *taskLock) release() error {
	err := redisdb.Unlock("backgroundtask_" + l.TaskName)
	if err != nil && err != redisdb.ErrNil {
		return err
	}

	return nil
}
