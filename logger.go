package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	_VER string = "1.0.0"
)

type LEVEL int32

var logLevel LEVEL = 1
var maxFileSize int64
var maxFileCount int32
var dailyRolling bool = true
var consoleAppender bool = true
var RollingFile bool = false
var logObj *_FILE

const DATEFORMAT = "2006-01-02"

type UNIT int64

const (
	_       = iota
	KB UNIT = 1 << (iota * 10)
	MB
	GB
	TB
)

const (
	ALL LEVEL = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	OFF
)

type _FILE struct {
	dir      string
	filename string
	_suffix  int
	isCover  bool
	_date    *time.Time
	mu       *sync.RWMutex
	logfile  *os.File
	lg       *log.Logger
}

//指定是否控制台打印，默认为true
func SetConsole(isConsole bool) {
	consoleAppender = isConsole
}

//指定日志级别  ALL，DEBUG，INFO，WARN，ERROR，FATAL，OFF 级别由低到高
//一般习惯是测试阶段为debug，生成环境为info以上
func SetLevel(_level LEVEL) {
	logLevel = _level
}

//默认方式新建日志，日志不进行分割
//第一个参数为日志文件存放目录
//第二个参数为日志文件命名
func Open(fileDir, fileName string) {
	RollingFile = false
	dailyRolling = false
	t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
	logObj = &_FILE{dir: fileDir, filename: fileName, _date: &t, isCover: false, mu: new(sync.RWMutex)}
	logObj.mu.Lock()
	defer logObj.mu.Unlock()

	logObj.logfile, _ = os.OpenFile(fileDir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	logObj.lg = log.New(logObj.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
}

//指定日志文件备份方式为日期的方式
//第一个参数为日志文件存放目录
//第二个参数为日志文件命名
func OpenRollDaily(fileDir, fileName string) {
	RollingFile = false
	dailyRolling = true
	t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
	logObj = &_FILE{dir: fileDir, filename: fileName, _date: &t, isCover: false, mu: new(sync.RWMutex)}
	logObj.mu.Lock()
	defer logObj.mu.Unlock()

	if !logObj.isMustRename() {
		logObj.logfile, _ = os.OpenFile(fileDir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		logObj.lg = log.New(logObj.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logObj.rename()
	}
}

//指定日志文件备份方式为文件大小的方式
//第一个参数为日志文件存放目录
//第二个参数为日志文件命名
//第三个参数为备份文件最大数量
//第四个参数为备份文件大小
//第五个参数为文件大小的单位
//logger.OpenRollSize("/var/log", "test.log", 10, 100, logger.MB)
func OpenRollSize(fileDir, fileName string, maxNumber int32, maxSize int64, _unit UNIT) {
	maxFileCount = maxNumber
	maxFileSize = maxSize * int64(_unit)
	RollingFile = true
	dailyRolling = false
	logObj = &_FILE{dir: fileDir, filename: fileName, isCover: false, mu: new(sync.RWMutex)}
	logObj.mu.Lock()
	defer logObj.mu.Unlock()
	for i := 1; i <= int(maxNumber); i++ {
		if isExist(fileDir + "/" + fileName + "." + strconv.Itoa(i)) {
			logObj._suffix = i
		} else {
			break
		}
	}
	if !logObj.isMustRename() {
		logObj.logfile, _ = os.OpenFile(fileDir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		logObj.lg = log.New(logObj.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logObj.rename()
	}
	go fileMonitor()
}

func Close() {
	logObj.mu.RLock()
	if logObj.logfile != nil {
		logObj.logfile.Close()
	}
}

func console(format string, s ...interface{}) {
	if consoleAppender {
		_, file, line, _ := runtime.Caller(2)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		//log.Println(file+":"+strconv.Itoa(line), s)
		log.Printf(file+":"+strconv.Itoa(line)+": "+format, s...)
	}
}
func consoleln(s ...interface{}) {
	if consoleAppender {
		_, file, line, _ := runtime.Caller(2)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		log.Println(file+":"+strconv.Itoa(line), s)
	}
}

func catchError() {
	if err := recover(); err != nil {
		log.Println("err", err)
	}
}

func Debugln(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()

	if logLevel <= DEBUG {
		logObj.lg.Output(2, fmt.Sprintln("[DEBUG]", v))
		consoleln("[DEBUG]", v)
	}
}
func Infoln(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= INFO {
		logObj.lg.Output(2, fmt.Sprintln("[INFO]", v))
		consoleln("[INFO]", v)
	}
}
func Warnln(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= WARN {
		logObj.lg.Output(2, fmt.Sprintln("[WARNING]", v))
		consoleln("[WARNING]", v)
	}
}
func Errorln(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= ERROR {
		logObj.lg.Output(2, fmt.Sprintln("[ERROR]", v))
		consoleln("[ERROR]", v)
	}
}
func Fatalln(v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= FATAL {
		logObj.lg.Output(2, fmt.Sprintln("[FATAL]", v))
		consoleln("[FATAL]", v)
	}
}

func Debug(format string, v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()

	if logLevel <= DEBUG {
		logObj.lg.Output(2, fmt.Sprintf("[DEBUG] "+format, v...))
		console("[DEBUG] "+format, v...)
	}
}
func Info(format string, v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= INFO {
		logObj.lg.Output(2, fmt.Sprintf("[INFO] "+format, v...))
		console("[INFO] "+format, v...)
	}
}
func Warn(format string, v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= WARN {
		logObj.lg.Output(2, fmt.Sprintf("[WARNING] "+format, v...))
		console("[WARNING] "+format, v...)
	}
}
func Error(format string, v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= ERROR {
		logObj.lg.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
		console("[ERROR] "+format, v...)
	}
}
func Fatal(format string, v ...interface{}) {
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()
	if logLevel <= FATAL {
		logObj.lg.Output(2, fmt.Sprintf("[FATAL] "+format, v...))
		console("[FATAL] "+format, v...)
	}
}

func (f *_FILE) isMustRename() bool {
	if dailyRolling {
		t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
		if t.After(*f._date) {
			return true
		}
	} else {
		if maxFileCount > 1 {
			if fileSize(f.dir+"/"+f.filename) >= maxFileSize {
				return true
			}
		}
	}
	return false
}

func (f *_FILE) rename() {
	if dailyRolling {
		fn := f.dir + "/" + f.filename + "." + f._date.Format(DATEFORMAT)
		if !isExist(fn) && f.isMustRename() {
			if f.logfile != nil {
				f.logfile.Close()
			}
			err := os.Rename(f.dir+"/"+f.filename, fn)
			if err != nil {
				f.lg.Println("rename err", err.Error())
			}
			t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
			f._date = &t
			f.logfile, _ = os.Create(f.dir + "/" + f.filename)
			f.lg = log.New(logObj.logfile, "\n", log.Ldate|log.Ltime|log.Lshortfile)
		}
	} else {
		f.coverNextOne()
	}
}

func (f *_FILE) nextSuffix() int {
	return int(f._suffix%int(maxFileCount) + 1)
}

func (f *_FILE) coverNextOne() {
	f._suffix = f.nextSuffix()
	if f.logfile != nil {
		f.logfile.Close()
	}
	if isExist(f.dir + "/" + f.filename + "." + strconv.Itoa(int(f._suffix))) {
		os.Remove(f.dir + "/" + f.filename + "." + strconv.Itoa(int(f._suffix)))
	}
	os.Rename(f.dir+"/"+f.filename, f.dir+"/"+f.filename+"."+strconv.Itoa(int(f._suffix)))
	f.logfile, _ = os.Create(f.dir + "/" + f.filename)
	f.lg = log.New(logObj.logfile, "\n", log.Ldate|log.Ltime|log.Lshortfile)
}

func fileSize(file string) int64 {
	fmt.Println("fileSize", file)
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func fileMonitor() {
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			fileCheck()
		}
	}
}

func fileCheck() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if logObj != nil && logObj.isMustRename() {
		logObj.mu.Lock()
		defer logObj.mu.Unlock()
		logObj.rename()
	}
}
