package apis

import (
	"net/http"

	"github.com/labstack/echo"
)

func NewTask(ctx echo.Context) error {
	return ctx.HTML(http.StatusOK, "xx")
}
