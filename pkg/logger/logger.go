package logger

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var (
	logFilePath = "./"
	logFileName = "akita.log"
)

func LogerMiddleware() gin.HandlerFunc {
	// 日志文件
	logFile := path.Join(logFilePath, logFileName)
	// 写入文件
	src, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Open log file err:", err)
	}
	// 实例化
	logger := logrus.New()
	//设置日志级别
	logger.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	//设置输出
	logger.Out = src

	// 设置 rotatelogs
	logWriter, err := rotatelogs.New(
		logFile+".%Y%m%d",                         // 分割后的文件名称
		rotatelogs.WithLinkName(logFile),          // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 设置最大保存时间(7天)
		rotatelogs.WithRotationTime(24*time.Hour), // 设置日志切割时间间隔(1天)
	)
	if err != nil {
		fmt.Printf("Failed to create rotatelogs: %s", err)
	}

	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  logWriter,
		logrus.FatalLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}

	logger.AddHook(lfshook.NewHook(writeMap, &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	}))

	return func(c *gin.Context) {
		startTime := time.Now()               //开始时间
		c.Next()                              //处理请求
		endTime := time.Now()                 //结束时间
		latencyTime := endTime.Sub(startTime) // 执行时间
		reqMethod := c.Request.Method         //请求方式
		reqUrl := c.Request.RequestURI        //请求路由
		statusCode := c.Writer.Status()       //状态码
		clientIP := c.ClientIP()              //请求IP
		// 日志格式
		logger.WithFields(logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqUrl,
		}).Info()
	}
}
