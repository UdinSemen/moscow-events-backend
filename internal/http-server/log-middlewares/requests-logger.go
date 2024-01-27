package log_middlewares

import (
	"fmt"
	"time"

	httprnd "github.com/UdinSemen/moscow-events-backend/pkg/http-rnd"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	RequestIDCtx         = "requestIDCtx"
	mes                  = "request_logger"
	nameFieldReqID       = "request_id"
	nameFieldPath        = "path"
	nameFieldIp          = "client_ip"
	nameFieldTimeProcess = "work_time"
)

func RequestLogger(c *gin.Context) {
	const op = "log_middleware.RequestLogger"

	timeNow := time.Now()
	reqID, err := httprnd.MakeReqId()
	if err != nil {
		zap.S().Errorf("%s:%v", op, err)
	}
	c.Set(RequestIDCtx, reqID)

	// call next middleware in stack
	c.Next()

	logger := zap.L().WithOptions(zap.WithCaller(false))

	logger.Info(mes,
		zap.String(nameFieldPath, c.FullPath()),
		zap.String(nameFieldIp, c.ClientIP()),
		zap.String(nameFieldReqID, reqID),
		zap.String(nameFieldTimeProcess, fmt.Sprintf("%v", time.Since(timeNow.Round(time.Microsecond)))),
	)

}
