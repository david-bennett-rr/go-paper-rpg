package util

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	debugFile *os.File
	debugOnce sync.Once
)

// DebugLog writes a formatted message to debug.log in the working directory.
func DebugLog(format string, args ...any) {
	debugOnce.Do(func() {
		var err error
		debugFile, err = os.Create("debug.log")
		if err != nil {
			return
		}
		fmt.Fprintf(debugFile, "=== Paper RPG Debug Log - %s ===\n", time.Now().Format(time.RFC3339))
	})
	if debugFile == nil {
		return
	}
	fmt.Fprintf(debugFile, format+"\n", args...)
	debugFile.Sync()
}

// CloseDebugLog flushes and closes the debug log file.
func CloseDebugLog() {
	if debugFile != nil {
		debugFile.Close()
	}
}
