logger
======

A simple log library for go programs.

### Features

Logger can output formated log to log file.  
Cut log file by daily or file size.  
Easy to use for any golang program.

### Usage

//指定是否控制台打印，默认为true  
func SetConsole(isConsole bool)

//指定日志级别  ALL，DEBUG，INFO，WARN，ERROR，FATAL，OFF 级别由低到高  
//一般习惯是测试阶段为debug，生成环境为info以上  
func SetLevel(_level LEVEL)

//默认方式新建日志，日志不进行分割  
//第一个参数为日志文件存放目录  
//第二个参数为日志文件命名  
func Open(fileDir, fileName string)

//指定日志文件备份方式为日期的方式  
//第一个参数为日志文件存放目录  
//第二个参数为日志文件命名  
func OpenRollDaily(fileDir, fileName string)

//指定日志文件备份方式为文件大小的方式  
//第一个参数为日志文件存放目录  
//第二个参数为日志文件命名  
//第三个参数为备份文件最大数量  
//第四个参数为备份文件大小  
//第五个参数为文件大小的单位  
//logger.OpenRollSize("/var/log", "test.log", 10, 100, logger.MB)  
func OpenRollSize(fileDir, fileName string, maxNumber int32, maxSize int64, _unit UNIT)

### Example

```
import (
	"github.com/buaazp/logger"
)

func main() {
	logger.OpenRollDaily("/var/log", "test.logger")
	logger.SetConsole(true)
	logger.SetLevel(logger.DEBUG)
	defer logger.Close()
	
	logger.Info("This is info test.\n")
	logger.Debugln("This is debugln test.")
	
	str := "This is a string."
	num := 4869
	
	logger.Debug("string: %s\nnumber: %d\n", str, num)	
}

```


### Install

```
go get github.com/buaazp/logger
```

### Lisence


Copyright (c) 2014, 招牌疯子  
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
