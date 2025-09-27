package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ServiceLogger struct {
	logger *zap.SugaredLogger
}

func NewServiceLogger() *ServiceLogger {
	newLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer newLogger.Sync()
	return &ServiceLogger{logger: newLogger.Sugar()}
}

type (
	ResponseData struct {
		Status int
		Size   int
	}
	LoggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode
}

func (log *ServiceLogger) RequestLogger(h http.Handler) http.Handler {
	logerFunc := func(w http.ResponseWriter, req *http.Request) {

		start := time.Now()
		responseData := &ResponseData{Status: 200, Size: 0}
		lw := LoggingResponseWriter{ResponseWriter: w, ResponseData: responseData}

		h.ServeHTTP(&lw, req)

		duration := time.Since(start)
		log.logger.Infoln(
			"uri", req.RequestURI,
			"method", req.Method,
			"status", responseData.Status,
			"duration", duration,
			"size", responseData.Size,
		)
	}
	return http.HandlerFunc(logerFunc)
}
