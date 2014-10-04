package logger

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	_VER string = "1.0.0"
)

const DATEFORMAT = "2006-01-02"

type LEVEL int32
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
	WARNING
	ERROR
	FATAL
	OFF
)

type logFile struct {
	dir             string
	filename        string
	_suffix         int
	isCover         bool
	_date           *time.Time
	mu              *sync.RWMutex
	logfile         *os.File
	lg              *log.Logger
	maxFileSize     int64
	maxFileCount    int32
	dailyRolling    bool
	consoleAppender bool
	RollingFile     bool
	stdlog          *log.Logger
	logLevel        LEVEL
	quitChan        chan int
}

//指定是否控制台打印，默认为true
func (l *logFile) SetConsole(isConsole bool) {
	l.consoleAppender = isConsole
}

//指定日志级别  ALL，DEBUG，INFO，WARN，ERROR，FATAL，OFF 级别由低到高
//一般习惯是测试阶段为debug，生成环境为info以上
func (l *logFile) SetLevel(_level LEVEL) {
	l.logLevel = _level
}

//默认方式新建日志，日志不进行分割
//第一个参数为日志文件存放目录
//第二个参数为日志文件命名
func Open(fileDir, fileName string) (*logFile, error) {
	t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
	logObj := &logFile{
		dir:             fileDir,
		filename:        fileName,
		_date:           &t,
		isCover:         false,
		mu:              new(sync.RWMutex),
		consoleAppender: true,
		logLevel:        INFO,
		RollingFile:     false,
		dailyRolling:    false,
	}
	logObj.mu.Lock()
	defer logObj.mu.Unlock()

	os.MkdirAll(fileDir, 0777)
	logfile, err := os.OpenFile(fileDir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	} else {
		logObj.logfile = logfile
	}
	logObj.lg = log.New(logObj.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
	logObj.stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	return logObj, nil
}

//指定日志文件备份方式为日期的方式
//第一个参数为日志文件存放目录
//第二个参数为日志文件命名
func OpenRollDaily(fileDir, fileName string) (*logFile, error) {
	t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
	logObj := &logFile{
		dir:             fileDir,
		filename:        fileName,
		_date:           &t,
		isCover:         false,
		mu:              new(sync.RWMutex),
		consoleAppender: true,
		logLevel:        INFO,
		RollingFile:     false,
		dailyRolling:    true,
	}
	logObj.mu.Lock()
	defer logObj.mu.Unlock()

	if !logObj.isMustRename() {
		os.MkdirAll(fileDir, 0777)
		logfile, err := os.OpenFile(fileDir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		} else {
			logObj.logfile = logfile
		}
		logObj.lg = log.New(logObj.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
		logObj.stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logObj.rename()
	}
	return logObj, nil
}

//指定日志文件备份方式为文件大小的方式
//第一个参数为日志文件存放目录
//第二个参数为日志文件命名
//第三个参数为备份文件最大数量
//第四个参数为备份文件大小
//第五个参数为文件大小的单位
//logger.OpenRollSize("/var/log", "test.log", 10, 100, logger.MB)
func OpenRollSize(fileDir, fileName string, maxNumber int32, maxSize int64, _unit UNIT) (*logFile, error) {
	quitChan := make(chan int, 0)
	logObj := &logFile{
		dir:             fileDir,
		filename:        fileName,
		isCover:         false,
		mu:              new(sync.RWMutex),
		maxFileCount:    maxNumber,
		maxFileSize:     maxSize * int64(_unit),
		consoleAppender: true,
		logLevel:        INFO,
		RollingFile:     true,
		dailyRolling:    false,
		quitChan:        quitChan,
	}
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
		os.MkdirAll(fileDir, 0777)
		logfile, err := os.OpenFile(fileDir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		} else {
			logObj.logfile = logfile
		}
		logObj.lg = log.New(logObj.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
		logObj.stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logObj.rename()
	}
	go fileMonitor(logObj)
	return logObj, nil
}

func (l *logFile) Close() {
	l.mu.RLock()
	if l.quitChan != nil {
		l.quitChan <- 1
	}
	if l.logfile != nil {
		l.logfile.Close()
	}
}

func (l *logFile) isMustRename() bool {
	if l.dailyRolling {
		t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
		if t.After(*l._date) {
			return true
		}
	} else {
		if l.maxFileCount > 1 {
			if fileSize(l.dir+"/"+l.filename) >= l.maxFileSize {
				return true
			}
		}
	}
	return false
}

func (l *logFile) rename() {
	if l.dailyRolling {
		fn := l.dir + "/" + l.filename + "." + l._date.Format(DATEFORMAT)
		if !isExist(fn) && l.isMustRename() {
			if l.logfile != nil {
				l.logfile.Close()
			}
			err := os.Rename(l.dir+"/"+l.filename, fn)
			if err != nil {
				l.lg.Println("rename err", err.Error())
			}
			t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
			l._date = &t
			l.logfile, _ = os.Create(l.dir + "/" + l.filename)
			l.lg = log.New(l.logfile, "\n", log.Ldate|log.Ltime|log.Lshortfile)
		}
	} else {
		l.coverNextOne()
	}
}

func (l *logFile) nextSuffix() int {
	return int(l._suffix%int(l.maxFileCount) + 1)
}

func (l *logFile) coverNextOne() {
	l._suffix = l.nextSuffix()
	if l.logfile != nil {
		l.logfile.Close()
	}
	if isExist(l.dir + "/" + l.filename + "." + strconv.Itoa(int(l._suffix))) {
		os.Remove(l.dir + "/" + l.filename + "." + strconv.Itoa(int(l._suffix)))
	}
	os.Rename(l.dir+"/"+l.filename, l.dir+"/"+l.filename+"."+strconv.Itoa(int(l._suffix)))
	l.logfile, _ = os.Create(l.dir + "/" + l.filename)
	l.lg = log.New(l.logfile, "\n", log.Ldate|log.Ltime|log.Lshortfile)
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

func fileMonitor(l *logFile) {
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			l.fileCheck()
		case <-l.quitChan:
			break
		}
	}
}

func (l *logFile) fileCheck() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if l != nil && l.isMustRename() {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.rename()
	}
}

func catchError() {
	if err := recover(); err != nil {
		log.Println("err", err)
	}
}

func (l *logFile) Debugln(v ...interface{}) {
	if l.logLevel <= DEBUG {
		if l.dailyRolling {
			l.fileCheck()
		}
		defer catchError()
		l.mu.RLock()
		defer l.mu.RUnlock()

		l.lg.Output(2, fmt.Sprintf("[DEBUG] %s", fmt.Sprintln(v...)))
		if l.consoleAppender {
			l.stdlog.Output(2, fmt.Sprintf("[DEBUG] %s", fmt.Sprintln(v...)))
		}
	}
}

func (l *logFile) Infoln(v ...interface{}) {
	if l.logLevel <= INFO {
		if l.dailyRolling {
			l.fileCheck()
		}
		defer catchError()
		l.mu.RLock()
		defer l.mu.RUnlock()
		l.lg.Output(2, fmt.Sprintf("[INFO] %s", fmt.Sprintln(v...)))
		if l.consoleAppender {
			l.stdlog.Output(2, fmt.Sprintf("[INFO] %s", fmt.Sprintln(v...)))
		}
	}
}

func (l *logFile) Warnln(v ...interface{}) {
	if l.logLevel <= WARNING {
		if l.dailyRolling {
			l.fileCheck()
		}
		defer catchError()
		l.mu.RLock()
		defer l.mu.RUnlock()
		l.lg.Output(2, fmt.Sprintf("[WARNING] %s", fmt.Sprintln(v...)))
		if l.consoleAppender {
			l.stdlog.Output(2, fmt.Sprintf("[WARNING] %s", fmt.Sprintln(v...)))
		}
	}
}

func (l *logFile) Errorln(v ...interface{}) {
	if l.logLevel <= ERROR {
		if l.dailyRolling {
			l.fileCheck()
		}
		defer catchError()
		l.mu.RLock()
		defer l.mu.RUnlock()
		l.lg.Output(2, fmt.Sprintf("[ERROR] %s", fmt.Sprintln(v...)))
		if l.consoleAppender {
			l.stdlog.Output(2, fmt.Sprintf("[ERROR] %s", fmt.Sprintln(v...)))
		}
	}
}

func (l *logFile) Fatalln(v ...interface{}) {
	if l.logLevel <= FATAL {
		if l.dailyRolling {
			l.fileCheck()
		}
		defer catchError()
		l.mu.RLock()
		defer l.mu.RUnlock()
		l.lg.Output(2, fmt.Sprintf("[FATAL] %s", fmt.Sprintln(v...)))
		if l.consoleAppender {
			l.stdlog.Output(2, fmt.Sprintf("[FATAL] %s", fmt.Sprintln(v...)))
		}
	}
}

func (l *logFile) Debug(format string, v ...interface{}) {
	if l.logLevel <= DEBUG {
		if l.dailyRolling {
			l.fileCheck()
		}
		defer catchError()
		l.mu.RLock()
		defer l.mu.RUnlock()

		l.lg.Output(2, fmt.Sprintf("[DEBUG] "+format, v...))
		if l.consoleAppender {
			l.stdlog.Output(2, fmt.Sprintf("[DEBUG] "+format, v...))
		}
	}
}

func (l *logFile) Info(format string, v ...interface{}) {
	if l.logLevel <= INFO {
		if l.dailyRolling {
			l.fileCheck()
		}
		defer catchError()
		l.mu.RLock()
		defer l.mu.RUnlock()
		l.lg.Output(2, fmt.Sprintf("[INFO] "+format, v...))
		if l.consoleAppender {
			l.stdlog.Output(2, fmt.Sprintf("[INFO] "+format, v...))
		}
	}
}

func (l *logFile) Warn(format string, v ...interface{}) {
	if l.logLevel <= WARNING {
		if l.dailyRolling {
			l.fileCheck()
		}
		defer catchError()
		l.mu.RLock()
		defer l.mu.RUnlock()
		l.lg.Output(2, fmt.Sprintf("[WARNING] "+format, v...))
		if l.consoleAppender {
			l.stdlog.Output(2, fmt.Sprintf("[WARNING] "+format, v...))
		}
	}
}

func (l *logFile) Error(format string, v ...interface{}) {
	if l.logLevel <= ERROR {
		if l.dailyRolling {
			l.fileCheck()
		}
		defer catchError()
		l.mu.RLock()
		defer l.mu.RUnlock()
		l.lg.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
		if l.consoleAppender {
			l.stdlog.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
		}
	}
}

func (l *logFile) Fatal(format string, v ...interface{}) {
	if l.logLevel <= FATAL {
		if l.dailyRolling {
			l.fileCheck()
		}
		defer catchError()
		l.mu.RLock()
		defer l.mu.RUnlock()
		l.lg.Output(2, fmt.Sprintf("[FATAL] "+format, v...))
		if l.consoleAppender {
			l.stdlog.Output(2, fmt.Sprintf("[FATAL] "+format, v...))
		}
	}
}
