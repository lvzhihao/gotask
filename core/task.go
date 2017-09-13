package core

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/lvzhihao/goutils"
)

var (
	CallBackTaskType int32 = 10000
)

type TaskInterface interface {
	SetExecTime(time.Time)
	SetParams(map[string]interface{})
	Run() error
	Status() interface{}
}

type Task struct {
	Type        int32
	CreateTime  time.Time
	UpdateTime  time.Time
	ExecuteTime time.Time
	Params      map[string]interface{}
}

type CallBackTask struct {
	Task
	rsp *http.Response
}

func NewCallBackTask() *CallBackTask {
	return &CallBackTask{
		Task: Task{
			Type:       CallBackTaskType,
			CreateTime: time.Now(),
			Params:     make(map[string]interface{}, 0),
		},
	}
}

func (c *CallBackTask) SetExecTime(t time.Time) {
	c.ExecuteTime = t
}

func (c *CallBackTask) SetParams(input map[string]interface{}) {
	c.Params = input
}

func (c *CallBackTask) Run() error {
	url, ok := c.Params["url"]
	if !ok {
		return errors.New("no callback url")
	}
	req, err := http.NewRequest("GET", goutils.ToString(url), nil)
	if err != nil {
		return err
	}
	//req.Header.Add("Content-Type", "application/json;charset=utf-8")
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	c.rsp, err = client.Do(req)
	return err
}

func (c *CallBackTask) Status() interface{} {
	//todo 处理
	var b []byte
	var err error
	var status string
	if c.rsp != nil {
		b, err = ioutil.ReadAll(c.rsp.Body)
		status = c.rsp.Status
	} else {
		b = []byte("")
		err = errors.New("empty response")
		status = ""
	}
	return struct {
		Params map[string]interface{}
		Status string
		Body   string
		err    error
	}{
		c.Params,
		status,
		goutils.ToString(b),
		err,
	}
}
