package logmod

import "time"

type LogConfig struct {
	SplitMaxTime time.Duration // 最大拆分时间
	SplitMaxSize int           // 最大拆分尺寸
	TempPath     string        // 临时路径
	DisablePrint bool          // 关闭终端输出
}
