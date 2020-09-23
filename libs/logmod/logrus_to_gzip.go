package logmod

import (
	"compress/gzip"
	"errors"
	"log"
	"os"
	"time"
)

func NewGZip(fileName string) (gf *GzipFile, err error) {
	gf = new(GzipFile)
	err = gf.Open(fileName)
	gf.timer = time.Now()
	return gf, err
}

type FileEngine interface {
	IsNil() bool                // 确认文件是否打开状态
	Open(filePath string) error //进入打开状态
	Write([]byte) (int, error)
	Size() (int64, error)         //获取尺寸
	Flush() error                 //落磁盘
	Close() error                 //关闭文件
	Time(time.Time) time.Duration // 时间对比
	GetFileName() string          // 输出当前文件路径
}

// gzip 压缩， 非线程安全
type GzipFile struct {
	fw       *os.File     // 磁盘文件连接句柄
	gw       *gzip.Writer // 数据压缩句柄
	writeNum int          //写入次数自动落盘
	timer    time.Time
	fileName string
}

func (gf *GzipFile) Time(t time.Time) time.Duration {
	return t.Sub(gf.timer)
}

func (gf *GzipFile) IsNil() bool {
	return gf.fw == nil && gf.gw == nil
}
func (gf *GzipFile) Open(filePath string) (err error) {
	gf.fileName = filePath
	gf.fw, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0775)
	if err != nil {
		return
	}
	gf.gw = gzip.NewWriter(gf.fw)
	if gf.gw == nil {
		err = errors.New("gzip create is error ")
	}
	return
}
func (gf *GzipFile) GetFileName() string {
	return gf.fileName
}
func (gf *GzipFile) Write(line []byte) (int, error) {
	gf.writeNum++ //每500次释放一次内存
	if gf.writeNum >= 500 {
		defer func() {
			gf.Flush()
			gf.writeNum = 0
		}()
	}
	return gf.gw.Write(line)
}

//计算尺寸
func (gf *GzipFile) Size() (int64, error) {
	gf.Flush()
	info, err := gf.fw.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

//将gzip 压缩流向磁盘写入一次。
// gzip自带一个flush 函数， 该函数会造成想磁盘中声明一个占位符的问题。 最终大小和实际大小可能会不符.
// 且gzip的flush 提示主要用于网络传输
func (gf *GzipFile) Flush() error {
	err := gf.gw.Close()
	if err != nil {
		log.Println(err)
	}
	gf.gw = nil
	gf.gw = gzip.NewWriter(gf.fw)
	return err
}

// 文件close 操作会去掉 磁盘文件的.tmp扩展名
func (gf *GzipFile) Close() (err error) {

	err = gf.gw.Close()
	if err != nil {
		return err
	}
	err = gf.fw.Close()
	if err != nil {
		return err
	}
	gf.fw = nil
	gf.gw = nil

	return err
}
