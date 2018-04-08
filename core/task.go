package core

import (
	"time"

	"go.uber.org/zap"
)

var (
	CallBackTaskType int32 = 10000
)

type TaskInterface interface {
	SetExecTime(time.Time)                  //设置运行时间
	GetExecTime() time.Time                 //获取运行时间
	SetParams(map[string]interface{}) error //设置任务参数
	Run() error                             //执行
	Status() interface{}                    //结果状态
	HookRecover(interface{}) error          //异常处理
}

type Task struct {
	Type        int32
	CreateTime  time.Time
	UpdateTime  time.Time
	ExecuteTime time.Time
	Params      map[string]interface{}
}

func ExecuteTask(task interface{}) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				//todo call task recover hook
				Logger.Error("task recover", zap.Any("panic", r))
				err := task.(TaskInterface).HookRecover(r)
				if err != nil {
					Logger.Error("task recover hook error", zap.Error(err))
				}
			}
		}()
		t := time.After(task.(TaskInterface).GetExecTime().Sub(time.Now()))
		<-t
		//run
		err := task.(TaskInterface).Run()
		if err != nil {
			//log error
			//todo
			Logger.Error("task error", zap.Error(err))
		} else {
			//log success
			Logger.Info("task status", zap.Any("status", task.(TaskInterface).Status()))
		}
	}()
}
