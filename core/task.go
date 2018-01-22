package core

import (
	"time"
)

var (
	CallBackTaskType int32 = 10000
)

type TaskInterface interface {
	SetExecTime(time.Time)            //设置运行时间
	SetParams(map[string]interface{}) //设置任务参数
	Run() error                       //执行
	Status() interface{}              //结果状态
	HookRecover() error               //异常处理
}

type Task struct {
	Type        int32
	CreateTime  time.Time
	UpdateTime  time.Time
	ExecuteTime time.Time
	Params      map[string]interface{}
}
