package utils

import (
	"log"
	"os"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	InfoLogger = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	ErrorLogger = log.New(os.Stderr, "[ERROR] ", log.LstdFlags)
}

func Info(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

func Error(format string, v ...interface{}) {
	ErrorLogger.Printf(format, v...)
}

func FormatTimestamp(ts int64) string {
	return time.Unix(ts, 0).Format(time.RFC3339)
}

func LogBlock(blockNum uint64, timestamp int64) {
	Info("Block #%d created at %s", blockNum, FormatTimestamp(timestamp))
}

