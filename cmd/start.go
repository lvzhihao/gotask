// Copyright © 2017 edwin <edwin.lzh@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"
	"github.com/lvzhihao/gotask/core"
	"github.com/lvzhihao/goutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	apiTaskTypeMap map[string]int32 = map[string]int32{
		"callback":   core.CallBackTaskType,
		"framemerge": core.FrameMergeTaskType,
	}
)

type ApiResult struct {
	Code string      `json:"code"` //code: 000000
	Data interface{} `json:"data"` //result data
}

type NewTaskInput struct {
	TaskType string                 `json:"task_type"` //task_type
	TaskTime string                 `json:"task_time"` //task_exec_time
	Params   map[string]interface{} `json:"params"`    //task_parasm
	Durable  bool                   `json:"durable"`   //has_durable
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "gotask start",
	Long:  `go task console`,
	Run: func(cmd *cobra.Command, args []string) {
		var logger *zap.Logger
		if os.Getenv("DEBUG") == "true" {
			logger, _ = zap.NewDevelopment()
		} else {
			logger, _ = zap.NewProduction()
		}
		defer logger.Sync()
		//app.Logger.SetLevel(log.INFO)
		app := goutils.NewEcho()
		core.Logger = logger
		server := core.NewServer()
		// load merchant
		var merchants []map[string]interface{}
		err := viper.UnmarshalKey("merchants", &merchants)
		if err != nil {
			logger.Fatal("load merchants error", zap.Error(err))
		}
		for _, v := range merchants {
			var merchant core.Merchant
			err := json.Unmarshal([]byte(goutils.ToString(v)), &merchant)
			if err != nil {
				logger.Fatal("load merchants error", zap.Error(err))
			} else {
				core.EnableMerchant(&merchant)
				logger.Info("load merchant success", zap.Any("merchant", merchant))
			}
		}
		// action
		app.POST("/api/task", func(ctx echo.Context) error {
			//new task
			var params []NewTaskInput
			err := json.Unmarshal([]byte(ctx.FormValue("data")), &params)
			if err != nil {
				logger.Error("api error", zap.String("code", "000002"), zap.Error(err))
				return ctx.JSON(http.StatusOK, ApiResult{Code: "000002", Data: "error input"})
			}
			for _, p := range params {
				if t, ok := apiTaskTypeMap[p.TaskType]; ok {
					err := server.Add(t, p.TaskTime, p.Params)
					if err != nil {
						logger.Error("add task error", zap.Error(err))
						return ctx.JSON(http.StatusOK, ApiResult{Code: "000004", Data: err.Error()})
					}
				} else {
					logger.Error("task type error", zap.Any("type", p.TaskType))
					return ctx.JSON(http.StatusOK, ApiResult{Code: "000003", Data: "error task type"})
				}
			}
			return ctx.JSON(http.StatusOK, ApiResult{Code: "000000", Data: "success"})
		})
		app.POST("/sys/stop", func(ctx echo.Context) error {
			//todo
			logger.Info("Server stop...", zap.String("source", "api"))
			go func() {
				time.AfterFunc(1*time.Second, server.Stop)
			}()
			return ctx.NoContent(http.StatusOK)
		})
		// server
		go server.Start()
		// graceful shutdown
		goutils.EchoStartWithGracefulShutdown(app, ":8179")
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
