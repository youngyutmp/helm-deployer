package logging

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type loggerKeyType int

const loggerKey loggerKeyType = iota

var logEntry *log.Entry

func init() {
	logEntry = log.WithField("logger", "default")
}

//NewContextWithLogger returns new Context with attached logger
func NewContextWithLogger(ctx context.Context, logger *log.Entry) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

//FromContext returns logger from Context
func FromContext(ctx context.Context) *log.Entry {
	if ctx == nil {
		return logEntry
	}
	if ctxLogEntry, ok := ctx.Value(loggerKey).(*log.Entry); ok {
		return ctxLogEntry
	}
	return logEntry
}
