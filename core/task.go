package core

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

var (
	CallBackTaskType   int32 = 10000
	FrameMergeTaskType int32 = 20000
)

type TaskInterface interface {
	SetMerchantId(string)                   //设置商户ID
	SetExecTime(time.Time)                  //设置运行时间
	GetExecTime() time.Time                 //获取运行时间
	SetParams(map[string]interface{}) error //设置任务参数
	Run() error                             //执行
	Status() interface{}                    //结果状态
	HookRecover(interface{}) error          //异常处理
}

type Task struct {
	TaskId        string
	Type          int32                  //类型
	MerchantId    string                 //商户号
	CreatedTime   time.Time              //创建时间
	UpdatedTime   time.Time              //更新时间
	ExecuteTime   time.Time              //执行时间
	Status        int32                  //状态: 0 未执行 1 执行成功 2 执行失败  4 等待重试
	RetryCount    int8                   //重试次数
	NextRetryTime time.Time              //下次重试时间
	Params        map[string]interface{} //任务参数
}

func Sign(merchant *Merchant, data string) string {
	return strings.ToLower(fmt.Sprintf("%x", md5.Sum([]byte(data+merchant.MerchantSecret))))
}

func CheckSign(merchant *Merchant, data, sign string) bool {
	if strings.Compare(strings.ToLower(sign), Sign(merchant, data)) == 0 {
		return true
	} else {
		return false
	}
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
