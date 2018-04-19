package core

import (
	"log"
	"os"
	"time"

	"go.uber.org/zap"
)

var (
	DefaultTimeZone          string         = "Asia/Shanghai"  //default timezone
	DefaultTaskBlockInterval time.Duration  = 60 * time.Second //default Task Block Interval
	Loc                      *time.Location                    //时区
	Logger                   *zap.Logger
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
	mgr *TaskManager
}

func NewServer() *Server {
	return &Server{
		cmd: make(chan string, 0),
		mgr: NewTaskManager(DefaultTaskBlockInterval),
	}
}

func (c *Server) Add(taskType int32, merchantId, taskTime string, params map[string]interface{}) error {
	return c.mgr.CreateTask(taskType, merchantId, taskTime, params)
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
	c.cmd <- "stop"
}
