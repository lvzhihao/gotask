package core

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/lvzhihao/goutils"
	"go.uber.org/zap"
)

const (
	TASK_TIME_PARSE_FORMAT string = "2006-01-02 15:04:05" //time format
	TASK_CMD_DELETE_BLOCK  string = "dtb"                 //"delete_task_block"
)

type TaskManager struct {
	blocks        *sync.Map     //task blocks
	blockInterval time.Duration //task block interval
	cmdChan       chan *TaskCmd //task cmd chan
}

type TaskCmd struct {
	cmd    string                 //命令
	params map[string]interface{} //参数
}

func (c *TaskCmd) Cmd() string {
	return c.cmd
}

func (c *TaskCmd) GetParams() map[string]interface{} {
	return c.params
}

func (c *TaskCmd) GetParam(key string) (v interface{}, ok bool) {
	v, ok = c.params[key]
	return
}

func (c *TaskCmd) GetString(key string) string {
	if v, ok := c.GetParam(key); ok {
		return goutils.ToString(v)
	} else {
		return ""
	}
}

func (c *TaskCmd) GetInt64(key string) int64 {
	if v, ok := c.GetParam(key); ok {
		return goutils.ToInt64(v)
	} else {
		return 0
	}
}

func NewDeleteBlockCmd(label int64) *TaskCmd {
	return &TaskCmd{
		cmd: TASK_CMD_DELETE_BLOCK,
		params: map[string]interface{}{
			"label": label,
		},
	}
}

func NewTaskManager(blockInterval time.Duration) *TaskManager {
	c := &TaskManager{
		blocks:        new(sync.Map),
		blockInterval: blockInterval,
		cmdChan:       make(chan *TaskCmd, 100),
	}
	go c.run()
	return c
}

func (c *TaskManager) run() {
	for {
		select {
		case v := <-c.cmdChan:
			switch v.Cmd() {
			case TASK_CMD_DELETE_BLOCK:
				c.DeleteBlock(v.GetInt64("label"))
				Logger.Debug("TaskBlock Delete Success", zap.Int64("label", v.GetInt64("label")))
			}
		}
	}
}

func (c *TaskManager) GetBlockInterval() time.Duration {
	return c.blockInterval
}

func (c *TaskManager) CreateTask(taskType int32, taskTime string, params map[string]interface{}) error {
	var task interface{}
	switch taskType {
	case CallBackTaskType: //http callback
		task = NewCallBackTask()
	case FrameMergeTaskType: // frame merge http callback
		task = NewFrameMergeTask()
	default:
		return errors.New("TaskType Don't Found")
	}
	executeTime, err := time.ParseInLocation(TASK_TIME_PARSE_FORMAT, taskTime, Loc)
	if err != nil {
		return errors.New("TaskTime Error")
	}
	task.(TaskInterface).SetExecTime(executeTime)
	err = task.(TaskInterface).SetParams(params)
	if err != nil {
		return err
	}
	c.PreperTask(task)
	return nil
}

func (c *TaskManager) PreperTask(task interface{}) {
	if task.(TaskInterface).GetExecTime().Sub(time.Now()) < c.GetBlockInterval()+10*time.Second {
		//直接运行任务
		c.ExecuteTask(task)
	} else {
		//进入block准备执行
		c.ApplyBlock(task)
	}
}

func (c *TaskManager) ExecuteTask(task interface{}) {
	ExecuteTask(task)
}

func (c *TaskManager) ApplyBlock(task interface{}) {
	label := int64(math.Floor(float64(task.(TaskInterface).GetExecTime().Unix()) / float64(c.GetBlockInterval().Seconds()))) //int64
	block, ok := c.LoadBlock(label)
	if !ok {
		origin := time.Unix(label*int64(c.GetBlockInterval().Seconds()), 0)
		block = &TaskBlock{
			label:    label,
			origin:   origin,
			interval: c.GetBlockInterval(),
			tasks:    make([]interface{}, 0),
			cmdChan:  c.cmdChan,
		}
		c.StoreBlock(label, block)
		// execute block
		block.Execute()
		Logger.Debug("TaskBlock Create Success", zap.Int64("label", label), zap.Time("origin", origin), zap.Duration("interval", c.GetBlockInterval()))
	}
	block.AddTask(task)
}

func (c *TaskManager) StoreBlock(label int64, block *TaskBlock) {
	c.blocks.Store(goutils.ToString(label), block)
}

func (c *TaskManager) LoadBlock(label int64) (*TaskBlock, bool) {
	if obj, ok := c.blocks.Load(goutils.ToString(label)); ok {
		return obj.(*TaskBlock), true
	} else {
		return nil, false
	}
}

func (c *TaskManager) DeleteBlock(label int64) {
	c.blocks.Delete(goutils.ToString(label))
}

type TaskBlock struct {
	label    int64         //label
	origin   time.Time     //起点
	interval time.Duration //步长
	tasks    []interface{} //任务集
	cmdChan  chan *TaskCmd
}

func (c *TaskBlock) Execute() {
	go func() {
		defer func() {
			c.cmdChan <- NewDeleteBlockCmd(c.label)
			if r := recover(); r != nil {
				//todo call task recover hook
				Logger.Error("block recover", zap.Any("panic", r))
			}
		}()
		t := time.After(c.origin.Sub(time.Now()))
		<-t
		//run
		for _, task := range c.tasks {
			ExecuteTask(task)
		}
	}()
}

func (c *TaskBlock) AddTask(task interface{}) {
	c.tasks = append(c.tasks, task)
}
