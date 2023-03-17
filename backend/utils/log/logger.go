package log

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Logger interface {
	LogInfo(ctx context.Context, s string, p ...interface{})
	LogTrace(ctx context.Context, s string, p ...interface{})
	LogError(ctx context.Context, s string, p ...interface{})
	IsTraceEnabled(ctx context.Context) bool
}

type SimpleLogger struct {
	Module string
	Trace  bool
}

func ToStrPtr(s string) *string {
	return &s
}

func ToJSONString(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "    ")
	return string(b)
}

func NewSimpleLogger(trace bool, module string) Logger {
	return &SimpleLogger{
		Module: module,
		Trace:  trace,
	}
}

func (l *SimpleLogger) IsTraceEnabled(ctx context.Context) bool {
	return l.Trace
}

func (l *SimpleLogger) LogInfo(ctx context.Context, s string, p ...interface{}) {
	fmt.Printf("%s [%s] %s %s\n", time.Now().Format(time.RFC3339Nano), l.Module, "INFO", fmt.Sprintf(s, p...))
}
func (l *SimpleLogger) LogTrace(ctx context.Context, s string, p ...interface{}) {
	if l.Trace {
		fmt.Printf("%s [%s] %s %s\n", time.Now().Format(time.RFC3339Nano), l.Module, "TRACE", fmt.Sprintf(s, p...))
	}
}
func (l *SimpleLogger) LogError(ctx context.Context, s string, p ...interface{}) {
	fmt.Printf("%s [%s] %s %s\n", time.Now().Format(time.RFC3339Nano), l.Module, "ERROR", fmt.Sprintf(s, p...))
}
