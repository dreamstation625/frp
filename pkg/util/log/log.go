// Copyright 2016 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/fatedier/golib/log"
)

var (
	TraceLevel = log.TraceLevel
	DebugLevel = log.DebugLevel
	InfoLevel  = log.InfoLevel
	WarnLevel  = log.WarnLevel
	ErrorLevel = log.ErrorLevel
)

var Logger *log.Logger

func init() {
	Logger = log.New(
		log.WithLevel(log.InfoLevel),
	)
}

func InitLogger(logPath string, levelStr string, maxDays int, disableLogColor bool, writeAndConsole bool) {
	options := []log.Option{}
	if logPath == "console" {
		options = append(options, log.WithOutput(newConsoleOutput(disableLogColor)))
	} else {
		writer := log.NewRotateFileWriter(log.RotateFileConfig{
			FileName: logPath,
			Mode:     log.RotateFileModeDaily,
			MaxDays:  maxDays,
		})
		writer.Init()
		if writeAndConsole {
			options = append(options, log.WithOutput(&teeOutput{
				file:    writer,
				console: newConsoleOutput(disableLogColor),
			}))
		} else {
			options = append(options, log.WithOutput(writer))
		}
	}

	level, err := log.ParseLevel(levelStr)
	if err != nil {
		level = log.InfoLevel
	}
	options = append(options, log.WithLevel(level))
	Logger = Logger.WithOptions(options...)
}

func newConsoleOutput(disableLogColor bool) io.Writer {
	if disableLogColor {
		return os.Stdout
	}
	return log.NewConsoleWriter(log.ConsoleConfig{
		Colorful: true,
	}, os.Stdout)
}

type teeOutput struct {
	file    io.Writer
	console io.Writer
}

func (w *teeOutput) Write(p []byte) (n int, err error) {
	if w.console != nil {
		if _, err = w.console.Write(p); err != nil {
			return 0, err
		}
	}
	if w.file != nil {
		return w.file.Write(p)
	}
	return len(p), nil
}

func (w *teeOutput) WriteLog(p []byte, level log.Level, when time.Time) (n int, err error) {
	if w.console != nil {
		if lw, ok := w.console.(log.Writer); ok {
			if _, err = lw.WriteLog(p, level, when); err != nil {
				return 0, err
			}
		} else {
			if _, err = w.console.Write(p); err != nil {
				return 0, err
			}
		}
	}
	if w.file != nil {
		return w.file.Write(p)
	}
	return len(p), nil
}

func Errorf(format string, v ...any) {
	Logger.Errorf(format, v...)
}

func Warnf(format string, v ...any) {
	Logger.Warnf(format, v...)
}

func Infof(format string, v ...any) {
	Logger.Infof(format, v...)
}

func Debugf(format string, v ...any) {
	Logger.Debugf(format, v...)
}

func Tracef(format string, v ...any) {
	Logger.Tracef(format, v...)
}

func Logf(level log.Level, offset int, format string, v ...any) {
	Logger.Logf(level, offset, format, v...)
}

type WriteLogger struct {
	level  log.Level
	offset int
}

func NewWriteLogger(level log.Level, offset int) *WriteLogger {
	return &WriteLogger{
		level:  level,
		offset: offset,
	}
}

func (w *WriteLogger) Write(p []byte) (n int, err error) {
	Logger.Log(w.level, w.offset, string(bytes.TrimRight(p, "\n")))
	return len(p), nil
}
