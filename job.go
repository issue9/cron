// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cron

import (
	"fmt"
	"log"
	"time"

	"github.com/issue9/cron/schedule"
)

// 表示任务状态
const (
	Stoped State = iota
	Running
	Failed
)

// State 状态值类型
type State int8

// JobFunc 每一个定时任务实际上执行的函数签名
type JobFunc func() error

// Job 一个定时任务的基本接口
type Job struct {
	name      string
	f         JobFunc
	scheduler schedule.Scheduler
	state     State
	err       error // 出错时的错误内容

	prev, next time.Time
}

// New 添加一个新的定时任务
func (c *Cron) New(name string, f JobFunc, s schedule.Scheduler) error {
	if c.running {
		return ErrRunning
	}

	c.jobs = append(c.jobs, &Job{
		name:      name,
		f:         f,
		scheduler: s,
		state:     Stoped,
	})
	return nil
}

// Name 任务的名称
func (j *Job) Name() string { return j.name }

// Next 该任务关联的 Nexter 接口
func (j *Job) Next() schedule.Scheduler { return j.scheduler }

// State 获取当前的状态
func (j *Job) State() State { return j.state }

// Err 返回当前的错误信息
func (j *Job) Err() error { return j.err }

// 运行当前的任务
//
// errlog 在出错时，日志的输出通道，可以为空，表示不输出。
func (j *Job) run(now time.Time, errlog *log.Logger) {
	defer func() {
		if msg := recover(); msg != nil {
			if err, ok := msg.(error); ok {
				j.err = err
			} else {
				j.err = fmt.Errorf("job error: %v", msg)
			}

			j.state = Failed
		}

		if errlog != nil && j.err != nil {
			errlog.Println(j.err)
		}
	}()

	j.state = Running
	j.err = j.f()

	if j.err != nil {
		j.state = Failed
	} else {
		j.state = Stoped
		j.err = nil
	}

	j.prev = j.next
	j.next = j.scheduler.Next(j.next)
}

// 初始化当前任务，获取其下次执行时间。
func (j *Job) init(now time.Time) {
	j.next = j.scheduler.Next(now)
}
