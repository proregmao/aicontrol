package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic错误
				logrus.WithFields(logrus.Fields{
					"error": err,
					"stack": string(debug.Stack()),
					"path":  c.Request.URL.Path,
					"method": c.Request.Method,
				}).Error("服务器内部错误")

				// 返回统一的错误响应
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    50000,
					"message": "服务器内部错误",
					"data":    nil,
				})
				c.Abort()
			}
		}()

		c.Next()

		// 处理其他错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"path":  c.Request.URL.Path,
				"method": c.Request.Method,
			}).Error("请求处理错误")

			// 如果还没有设置响应状态码，设置为500
			if c.Writer.Status() == http.StatusOK {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    50001,
					"message": "请求处理失败",
					"data":    nil,
				})
			}
		}
	}
}

// NotFoundHandler 404处理器
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "接口不存在",
			"data":    nil,
		})
	}
}

// MethodNotAllowedHandler 405处理器
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    40500,
			"message": "请求方法不允许",
			"data":    nil,
		})
	}
}
