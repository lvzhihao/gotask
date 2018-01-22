package core

import (
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	DefaultTimeZone string         = "Asia/Shanghai" //default timezone
	Loc             *time.Location                   //时区
)

func init() {
	timeZone := os.Getenv("TASK_OS_TIMEZONE")
	if timeZone == "" {
		timeZone = DefaultTimeZone
	}
	var err error
	Loc, err = time.LoadLocation(timeZone)
	if err != nil {
		// 时区设置错误
		log.Fatal(err)
	}
}

//todo task manager

type Server struct {
	cmd chan string
	lk  sync.Mutex
	log *zap.Logger
}

func NewServer(log *zap.Logger) *Server {
	return &Server{
		cmd: make(chan string, 0),
		log: log,
	}
}

func (c *Server) Add(taskType, taskTime string, params map[string]interface{}) error {
	c.lk.Lock()
	defer c.lk.Unlock()
	var task interface{}
	switch taskType {
	case "callback":
		task = NewCallBackTask()
	default:
		return errors.New("TaskType Don't Found")
	}
	executeTime, err := time.ParseInLocation("2006-01-02 15:04:05", taskTime, Loc)
	if err != nil {
		return errors.New("TaskTime Error")
	}
	task.(TaskInterface).SetExecTime(executeTime)
	task.(TaskInterface).SetParams(params)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				//todo call task recover hook
				c.log.Error("task recover", zap.Any("panic", r))
			}
		}()
		t := time.After(executeTime.Sub(time.Now()))
		<-t
		//run
		err := task.(TaskInterface).Run()
		if err != nil {
			//todo call task error hook
			c.log.Error("task error", zap.Error(err))
		} else {
			//log success
			//todo call task succes hook
			c.log.Info("task status", zap.Any("status", task.(TaskInterface).Status()))
		}
	}()
	return nil
}

func (c *Server) Start() {
	select {
	case cmd := <-c.cmd:
		switch cmd {
		case "stop":
			os.Exit(0)
		}
	}
	os.Exit(1)
}

func (c *Server) Stop() {
	c.lk.Lock()
	defer c.lk.Unlock()
	c.cmd <- "stop"
}
