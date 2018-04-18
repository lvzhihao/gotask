package core

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/lvzhihao/goutils"
)

var (
	FrameMergeFlagMap *sync.Map
)

func init() {
	FrameMergeFlagMap = new(sync.Map)
}

type FrameMergeTask struct {
	Task
	rsp      *http.Response
	lk       sync.Mutex
	isMerge  bool
	values   []interface{}
	merchant *Merchant
}

func NewFrameMergeTask() *FrameMergeTask {
	return &FrameMergeTask{
		Task: Task{
			Type:       FrameMergeTaskType,
			CreateTime: time.Now(),
			Params:     make(map[string]interface{}, 0),
		},
		isMerge: false,
	}
}

func (c *FrameMergeTask) SetExecTime(t time.Time) {
	c.ExecuteTime = t
}

func (c *FrameMergeTask) GetExecTime() time.Time {
	return c.ExecuteTime
}

func (c *FrameMergeTask) UpdateParams(input []interface{}) {
	c.lk.Lock()
	defer c.lk.Unlock()
	c.values = append(c.values, input...)
}

func (c *FrameMergeTask) CheckParams() error {
	_, ok := c.Params["merchant"]
	if !ok {
		return errors.New("no merchant no")
	}
	var err error
	c.merchant, err = LoadMerchant(goutils.ToString(c.Params["merchant"]))
	if err != nil {
		return errors.New("merchant error")
	}
	_, ok = c.Params["frame"]
	if !ok {
		return errors.New("no frame id")
	}
	timestamp, ok := c.Params["timestamp"]
	if !ok {
		return errors.New("no timestamp")
	}
	if math.Abs(float64(time.Now().Unix()-goutils.ToInt64(timestamp))) > 600 {
		return errors.New("timestamp error")
	}
	sign, ok := c.Params["sign"]
	if !ok {
		return errors.New("no sign")
	}
	data, ok := c.Params["values"]
	if !ok {
		return errors.New("values must slice type")
	}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &c.values)
	if err != nil {
		return err
	} // values 必须是个maps
	if strings.Compare(strings.ToLower(goutils.ToString(sign)), c.Sign(goutils.ToInt64(timestamp))) != 0 {
		return errors.New("check sign error")
	}
	return nil
}

func (c *FrameMergeTask) Sign(timestamp int64) string {
	return strings.ToLower(fmt.Sprintf("%x", md5.Sum([]byte(goutils.ToString(timestamp)+c.merchant.MerchantSecret))))
}

func (c *FrameMergeTask) SetParams(input map[string]interface{}) error {
	c.Params = input
	// 判断数据完整性
	// load maps
	err := c.CheckParams()
	if err != nil {
		return err
	}
	obj, ok := FrameMergeFlagMap.Load(c.GetFlagId())
	if ok {
		obj.(*FrameMergeTask).UpdateParams(c.values)
		// 本次调用设置成合并成功，实际上不做任务操作
		c.SetExecTime(time.Now())
		c.isMerge = true
	} else {
		c.isMerge = false
		FrameMergeFlagMap.Store(c.GetFlagId(), c)
	}
	return nil
}

func (c *FrameMergeTask) GetFlagId() string {
	return fmt.Sprintf("%s-%s", c.Params["merchant"], c.Params["frame"])
}

func (c *FrameMergeTask) Run() error {
	c.lk.Lock()
	defer c.lk.Unlock()
	if c.isMerge {
		// merge success, do nothing
		return nil
	} else {
		// delete flagMap
		FrameMergeFlagMap.Delete(c.GetFlagId())
	}
	backurl, ok := c.Params["url"]
	if !ok {
		return errors.New("no callback url")
	}
	p := url.Values{}
	timestamp := time.Now().Unix()
	p.Set("timestamp", goutils.ToString(timestamp))
	p.Set("values", goutils.ToString(c.values))
	p.Set("merchant", c.merchant.MerchantNo)
	p.Set("sign", c.Sign(timestamp))
	req, err := http.NewRequest("POST", goutils.ToString(backurl), bytes.NewBufferString(p.Encode()))
	if err != nil {
		return err
	}
	if header, ok := c.Params["header"]; ok {
		for k, v := range header.(map[string]string) {
			req.Header.Set(k, v)
		}
	}
	//req.Header.Add("Content-Type", "application/json;charset=utf-8")
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	c.rsp, err = client.Do(req)
	return err
}

func (c *FrameMergeTask) Status() interface{} {
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
		Values []interface{}
		Status string
		Body   string
		err    error
	}{
		c.Params,
		c.values,
		status,
		goutils.ToString(b),
		err,
	}
}

func (c *FrameMergeTask) HookRecover(recover interface{}) error {
	//todo
	return nil
}
