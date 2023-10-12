/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-01-05 17:49:03
 * @LastEditTime: 2023-10-12 17:55:03
 */
package goutils

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// MyLog 自定义 log
type MyLog struct {
	zap.Logger
	InfoFile   string // info 日志路径
	ErrorFile  string // error 日志路径
	ToConsole  bool   // 日志是否显示在 console
	ToFile     bool   // 日志是否写入文件
	ShowCaller bool   // 日志是否显示 caller

	LevelKey   string // 日志中 level 的 key
	TimeKey    string // 日志中 time 的 key
	CallerKey  string // 日志中 caller 的 key
	MessageKey string // 日志中 msg 的 key

	MaxSize    int  // 文件大小限制，单位 MB
	MaxBackups int  // 最大保留日志文件数量
	MaxAge     int  // 日志文件保留天数
	Compress   bool // 是否压缩处理
}

// NewLog 初始化 MyLog
//
//	log := pkg.NewLog().L()
//	log.Info("info")
//	log.Error("error")
func NewLog() *MyLog {
	return &MyLog{
		InfoFile:   "./logs/info.log",
		ErrorFile:  "./logs/error.log",
		ToConsole:  true,
		ToFile:     false,
		ShowCaller: false,

		LevelKey:   "level",
		TimeKey:    "time",
		CallerKey:  "path",
		MessageKey: "msg",

		MaxSize:    10,
		MaxBackups: 50,
		MaxAge:     30,
		Compress:   false,
	}
}

// L 返回 *zap.Logger
func (mylog *MyLog) L() *zap.Logger {

	var coreArr []zapcore.Core
	if mylog.ToConsole && mylog.ToFile {
		consoleCore := setConsole(mylog)
		infoFileCore, errorFileCore := setFile(mylog)
		coreArr = []zapcore.Core{consoleCore, infoFileCore, errorFileCore}
	} else {
		if mylog.ToConsole {
			consoleCore := setConsole(mylog)
			coreArr = []zapcore.Core{consoleCore}
		}
		if mylog.ToFile {
			infoFileCore, errorFileCore := setFile(mylog)
			coreArr = []zapcore.Core{infoFileCore, errorFileCore}
		}
	}

	var log *zap.Logger
	if mylog.ShowCaller {
		log = zap.New(
			zapcore.NewTee(coreArr...),
			zap.AddCaller(), // zap.AddCaller() 设为显示文件名和行号
		)
	} else {
		log = zap.New(
			zapcore.NewTee(coreArr...),
		)
	}

	return log
}

// SetInfoFile 设置 Info 日志路径
func (mylog *MyLog) SetInfoFile(logfile string) *MyLog {
	mylog.InfoFile = logfile
	return mylog
}

// SetErrorFile 设置 Error 日志路径
func (mylog *MyLog) SetErrorFile(logfile string) *MyLog {
	mylog.ErrorFile = logfile
	return mylog
}

// IsToConsole 日志输出到终端
func (mylog *MyLog) IsToConsole(value bool) *MyLog {
	mylog.ToConsole = value
	return mylog
}

// IsToFile 日志输出到文件
func (mylog *MyLog) IsToFile(value bool) *MyLog {
	mylog.ToFile = value
	return mylog
}

// IsShowCaller 是否显示 Caller
func (mylog *MyLog) IsShowCaller(value bool) *MyLog {
	mylog.ShowCaller = value
	return mylog
}

// SetLevelKey 设置日志中 level 的 key
func (mylog *MyLog) SetLevelKey(key string) *MyLog {
	mylog.LevelKey = key
	return mylog
}

// SetTimeKey 设置日志中 time 的 key
func (mylog *MyLog) SetTimeKey(key string) *MyLog {
	mylog.TimeKey = key
	return mylog
}

// SetCallerKey 设置日志中 caller 的 key
func (mylog *MyLog) SetCallerKey(key string) *MyLog {
	mylog.CallerKey = key
	return mylog
}

// SetMessageKey 设置日志中 message 的 key
func (mylog *MyLog) SetMessageKey(key string) *MyLog {
	mylog.MessageKey = key
	return mylog
}

// SetMaxSize 设置日志文件的大小
func (mylog *MyLog) SetMaxSize(size int) *MyLog {
	mylog.MaxSize = size
	return mylog
}

// SetMaxBackups 设置日志文件的留存个数
func (mylog *MyLog) SetMaxBackups(num int) *MyLog {
	mylog.MaxBackups = num
	return mylog
}

// SetMaxAge 设置日志文件的留存天数
func (mylog *MyLog) SetMaxAge(num int) *MyLog {
	mylog.MaxAge = num
	return mylog
}

// IsCompress 日志文件是否压缩存储
func (mylog *MyLog) IsCompress(value bool) *MyLog {
	mylog.Compress = value
	return mylog
}

func setConsole(mylog *MyLog) zapcore.Core {
	// 配置终端日志显示格式，为普通文本格式
	encoderConsole := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		LevelKey:   mylog.LevelKey,
		TimeKey:    mylog.TimeKey,
		CallerKey:  mylog.CallerKey,
		MessageKey: mylog.MessageKey,
		// EncodeTime:   zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeLevel:  zapcore.CapitalColorLevelEncoder, // 按级别显示不同颜色
		EncodeCaller: zapcore.ShortCallerEncoder,       // 显示短文件路径
	})

	// 配置 Console 中日志格式
	consoleWriteSyncer := zapcore.AddSync(
		os.Stdout,
	)
	// zapcore.NewCore 第一个参数为日志格式，第二个参数为输出到哪里，第三个参数为日志级别
	consoleCore := zapcore.NewCore(encoderConsole, zapcore.NewMultiWriteSyncer(consoleWriteSyncer), zap.DebugLevel)

	return consoleCore
}

func setFile(mylog *MyLog) (zapcore.Core, zapcore.Core) {
	// 配置日志文件中日志的格式，为 json 格式
	encoderFile := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		LevelKey:     mylog.LevelKey,
		TimeKey:      mylog.TimeKey,
		CallerKey:    mylog.CallerKey,
		MessageKey:   mylog.MessageKey,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder, // 显示短文件路径
		// EncodeCaller: zapcore.FullCallerEncoder, // 显示完整文件路径
	})

	// 设置日志级别，debug/info/warn/error/dpanic/panic/fatal 对应 -1/0/1/2/3/4/5
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { // 低于 error 级别的记录
		return lev < zap.ErrorLevel
	})
	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { // 大于等于 error 级别的记录
		return lev >= zap.ErrorLevel
	})

	// 配置 debug/info
	infoFileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   mylog.InfoFile,   // 日志文件存放目录，如果文件夹不存在会自动创建
		MaxSize:    mylog.MaxSize,    // 文件大小限制，单位 MB
		MaxBackups: mylog.MaxBackups, // 最大保留日志文件数量
		MaxAge:     mylog.MaxAge,     // 日志文件保留天数
		Compress:   mylog.Compress,   // 是否压缩处理
	})
	infoFileCore := zapcore.NewCore(encoderFile, zapcore.NewMultiWriteSyncer(infoFileWriteSyncer), lowPriority)

	// error 文件 writeSyncer
	errorFileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   mylog.ErrorFile,  // 日志文件存放目录
		MaxSize:    mylog.MaxSize,    // 文件大小限制，单位 MB
		MaxBackups: mylog.MaxBackups, // 最大保留日志文件数量
		MaxAge:     mylog.MaxAge,     // 日志文件保留天数
		Compress:   mylog.Compress,   // 是否压缩处理
	})
	errorFileCore := zapcore.NewCore(encoderFile, zapcore.NewMultiWriteSyncer(errorFileWriteSyncer), highPriority)

	return infoFileCore, errorFileCore
}
