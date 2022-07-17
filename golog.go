//Package logs implements a simple library for log.
package golog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Logger log struct
type Logger struct {
	level    Level      //log level, 0:debug, 1:info, 2:warn, 3:error
	output   io.Writer  //stdout or file to write log
	buf      []byte     //concat byte
	mux      sync.Mutex //sync lock
	dir      string     //log file dir
	path     string     //log file name
	ext      string     //log file ext
	rolling  Rolling    //rolling mode, 0:no, 1:time, 2:size
	interval int64      //rolling interval
	record   int64      //record of time or size
	init     bool       // init flag
}

// Level log level type
type Level int

// Rolling log rolling type
type Rolling int

const (
	// DebugLevel log level 0
	DebugLevel Level = iota
	// InfoLevel log level 1
	InfoLevel
	// WarnLevel log level 2
	WarnLevel
	// ErrorLevel log level 3
	ErrorLevel
	// LastLevel level boundary
	LastLevel
)
const (
	noRolling Rolling = iota
	// TimeRolling rolling mode
	TimeRolling
	// SizeRolling rolling mode
	SizeRolling
)
const (
	// KB size 1024 * byte
	KB = 1024
	// MB size 1024 * KB
	MB = KB * 1024
	// MinSizeInterval min size of size rolling
	MinSizeInterval = 10 * KB //10KB
	// Second time 1
	Second = 1
	// Minute time 60 * Second
	Minute = 60 * Second
	// MinTimeInterval min time of time rolling
	MinTimeInterval = 10 * Second //10Second
)
const (
	// DebugFlag debug level flag
	DebugFlag = "debug"
	// InfoFlag info level flag
	InfoFlag = "info"
	// WarnFlag warn level flag
	WarnFlag = "warn"
	// ErrorFlag error level flag
	ErrorFlag = "error"
	// TimeFlag time rolling mode flag
	TimeFlag = "time"
	// SizeFlag size rolling mode flag
	SizeFlag = "size"
)

var (
	logger = &Logger{level: InfoLevel, init: false, rolling: noRolling}
)

func init() {
	log.SetPrefix("[logger] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// Log return Logger pointer
func Log() *Logger {
	return logger
}

// Level set log level
func (l *Logger) Level(flag string) *Logger {
	level := InfoLevel
	switch flag {
	case DebugFlag:
		level = DebugLevel
	case InfoFlag:
		level = InfoLevel
	case WarnFlag:
		level = WarnLevel
	case ErrorFlag:
		level = ErrorLevel
	default:
		panic(fmt.Sprintf("log flag %s not support(debug,info,warn,error)", flag))
	}
	l.level = level
	return l
}

// LogFile set log file
func (l *Logger) LogFile(logf string) *Logger {
	dir := filepath.Dir(logf)
	file := filepath.Base(logf)
	ext := path.Ext(file)
	name := strings.TrimSuffix(file, ext)
	if dir == "" || file == "" {
		panic("invalid log file")
	}
	if strings.Contains(file, "/") {
		panic("invalid path")
	}
	l.dir = dir
	l.path = name
	l.ext = ext
	return l
}

// Rolling set log rolling
// flag: time, size, empty(not rolling)
func (l *Logger) Rolling(flag string, interval int64) *Logger {
	if l.dir == "" || l.path == "" {
		panic("please set log file first")
	}
	switch flag {
	case TimeFlag:
		l.rolling = TimeRolling
	case SizeFlag:
		l.rolling = SizeRolling
	default:
		l.rolling = noRolling
	}
	if l.rolling != noRolling && interval <= 0 {
		panic("invalid interval value")
	}
	l.interval = interval
	switch l.rolling {
	case TimeRolling:
		if l.interval < MinTimeInterval {
			localPrint("Warn", "time rolling interval too small", yellow)
		}
	case SizeRolling:
		if l.interval < MinSizeInterval {
			localPrint("Warn", "size rolling interval too small", yellow)
		}
	default:
		panic("invalid rolling type")
	}
	return l
}

// initialize init logger
func (l *Logger) initialize() {
	if l.dir != "" && l.path != "" {
		fpath := fmt.Sprintf("%s%s", filepath.Join(l.dir, l.path), l.ext) //log file
		if l.rolling != noRolling {
			count := 1
			for {
				if !isExist(fpath) {
					break
				}
				fpath = fmt.Sprintf("%s_%d%s", filepath.Join(l.dir, l.path), count, l.ext)
				count++
			}
		}
		output, err := logfile(fpath)
		if err != nil {
			panic("initialize log file error: " + err.Error())
		}
		l.output = output
		if l.rolling == TimeRolling {
			l.record = time.Now().Unix()
		}
		if l.rolling == SizeRolling {
			l.record = 0
		}
	} else {
		l.output = os.Stdout

	}
	l.init = true
}

// logPrint print log with level
func (l *Logger) logPrint(level Level, s string, lineFeed bool) {
	var file string
	var line int
	var ok bool
	l.mux.Lock()
	_, file, line, ok = runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	l.mux.Unlock()
	l.buf = l.buf[:0]
	prefix(&l.buf, file, line)
	l.buf = append(l.buf, s...)
	var prefix, logstr string
	var color func(string) string
	switch level {
	case ErrorLevel:
		prefix = "E"
		color = red
	case WarnLevel:
		prefix = "W"
		color = yellow
	case InfoLevel:
		prefix = "I"
		color = green
	case DebugLevel:
		prefix = "D"
		color = blue
	default:
		return
	}
	logstr = printStruct(prefix, string(l.buf), color)
	if lineFeed {
		logstr = logstr + "\n"
	}
	l.foutput(logstr)
}

// foutput print msg to file
func (l *Logger) foutput(msg string) {
	l.mux.Lock()
	if !l.init {
		logger.initialize()
	}
	l.mux.Unlock()
	if l.rolling == TimeRolling {
		if l.record+l.interval < time.Now().Unix() {
			l.initialize()
		}
	}
	if l.rolling == SizeRolling {
		if l.record > l.interval {
			l.initialize()
		}
		l.record += int64(len(msg))
	}
	_, err := l.output.Write([]byte(msg))
	if err != nil {
		localPrint("Error", fmt.Sprintf("log output error %v\n", err), red)
		localPrint("OutPut", msg, blue)
	}
}

// Debug print log at debug level
func Debug(args ...interface{}) {
	if DebugLevel < logger.level {
		return
	}
	logger.logPrint(DebugLevel, fmt.Sprint(args...), true)
}

// Debugf print log at debug level with format
func Debugf(format string, args ...interface{}) {
	if DebugLevel < logger.level {
		return
	}
	logger.logPrint(DebugLevel, fmt.Sprintf(format, args...), false)
}

// Info print log at info level
func Info(args ...interface{}) {
	if InfoLevel < logger.level {
		return
	}
	logger.logPrint(InfoLevel, fmt.Sprint(args...), true)
}

// Infof print log at info level with format
func Infof(format string, args ...interface{}) {
	if InfoLevel < logger.level {
		return
	}
	logger.logPrint(InfoLevel, fmt.Sprintf(format, args...), false)
}

// Warn print log at warn level
func Warn(args ...interface{}) {
	if WarnLevel < logger.level {
		return
	}
	logger.logPrint(WarnLevel, fmt.Sprint(args...), true)
}

// Warnf print log at warn level with format
func Warnf(format string, args ...interface{}) {
	if WarnLevel < logger.level {
		return
	}
	logger.logPrint(WarnLevel, fmt.Sprintf(format, args...), false)
}

// Error print log at error level
func Error(args ...interface{}) {
	if ErrorLevel < logger.level {
		return
	}
	logger.logPrint(ErrorLevel, fmt.Sprint(args...), true)
}

// Errorf print log at error level with format
func Errorf(format string, args ...interface{}) {
	if ErrorLevel < logger.level {
		return
	}
	logger.logPrint(ErrorLevel, fmt.Sprintf(format, args...), false)
}

// Fatal print log and exit
func Fatal(args ...interface{}) {
	logger.logPrint(ErrorLevel, fmt.Sprint(args...), true)
	os.Exit(1)
}

// Fatalf print log with format and exit
func Fatalf(format string, args ...interface{}) {
	logger.logPrint(ErrorLevel, fmt.Sprintf(format, args...), false)
	os.Exit(1)
}

// logfile create log writer
func logfile(fpath string) (io.Writer, error) {
	if isExist(fpath) {
		localPrint("Warn", "log file exist, append it", yellow)
		return os.OpenFile(fpath, os.O_WRONLY|os.O_APPEND, 0644)
	}
	err := os.MkdirAll(path.Dir(fpath), os.ModePerm)
	if err == nil {
		return os.Create(fpath)
	}
	return nil, err
}

// isExist whether file exist
func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		log.Println(err)
		return false
	}
	return true
}

// prefix add prefix to buf
func prefix(buf *[]byte, file string, line int) {
	*buf = append(*buf, ' ')
	*buf = append(*buf, filepath.Base(file)...)
	*buf = append(*buf, ':')
	itoa(buf, line, -1)
	*buf = append(*buf, " "...)
}

// itoa parse int to bytes
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// printStruct assemble msg with time, color
func printStruct(prefix, msg string, color func(string) string) string {
	now := time.Now().Unix()
	time1 := time.Unix(now, 0).Format("2006-01-02")
	time2 := time.Unix(now, 0).Format("15:04:05")
	return color(fmt.Sprintf("[%s|%s|%s]%s", prefix, time1, time2, msg))
}

// localPrint logger inner print
func localPrint(prefix, msg string, color func(string) string) {
	fmt.Println(printStruct(prefix, msg, color))
}

// red print red log
func red(s string) string {
	return fmt.Sprintf("\x1b[%d;%dm%s\x1b[0m", 31, 3, s)
}

// green print green log
func green(s string) string {
	return fmt.Sprintf("\x1b[%d;%dm%s\x1b[0m", 32, 3, s)
}

// yellow print yellow log
func yellow(s string) string {
	return fmt.Sprintf("\x1b[%d;%dm%s\x1b[0m", 33, 3, s)
}

// blue print blue log
func blue(s string) string {
	return fmt.Sprintf("\x1b[%d;%dm%s\x1b[0m", 36, 3, s)
}
