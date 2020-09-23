package logmod

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/eatools/gservice/application/onstop"

	"github.com/sirupsen/logrus"
)

type msg struct {
	logtype string
	level   string
	timer   time.Time
	body    []byte
}

func NewGzipHook(temp string, size int) *GzipHook {
	hook := new(GzipHook)
	hook.typeFile = map[string]FileEngine{}
	hook.tempFilePath = temp // configure.Config.LogToFile.Temp
	hook.maxSize = size      //configure.Config.LogToFile.MaxSize
	hook.writeChan = make(chan msg, 20000)
	go hook.AutoSplit()

	// 系统退出时运行。
	onstop.Append(func() {
		hook.StopAll()
	})
	return hook
}

type GzipHook struct {
	// levelFile map[logrus.Level]FileEngine
	tempFilePath string
	lock         sync.Mutex
	/**/ maxSize int
	typeFile     map[string]FileEngine
	writeChan    chan msg
}

func (h *GzipHook) AutoSplit() {
	var err error
	var i = 0
	af := time.NewTicker(time.Duration(logConf.SplitMaxTime))
	for {
		select {
		case m := <-h.writeChan:
			i = i + 1
			key := m.logtype + m.level
			if ff, ok := h.typeFile[key]; !ok || ff == nil {
				buf := bytes.NewBuffer(nil)
				buf.WriteString("/tmp/")
				buf.WriteString(m.logtype)
				buf.WriteString("/")
				buf.WriteString(m.level)
				buf.WriteString("/")
				buf.WriteString(m.timer.Format("2006/01/02/15/"))
				// 创建文件夹
				p := buf.String()
				os.MkdirAll(p, 0776)
				os.MkdirAll(h.tempFilePath+p[5:len(p)], 0776)
				buf.WriteString(m.logtype)
				buf.WriteString("-")
				buf.WriteString(m.level)
				buf.WriteString("-")
				if name, err := os.Hostname(); err == nil {
					buf.WriteString(name)
					buf.WriteString("-")
				}
				buf.WriteString(m.timer.Format("2006-01-02T15.04.05.999999"))
				buf.WriteString(".gz.tmp")
				h.typeFile[key], err = NewGZip(buf.String())
				if err != nil {
					logrus.Error(err)
				}
			}
			h.typeFile[key].Write(m.body)
			h.typeFile[key].Write([]byte("\n"))
			if i >= 3 {
				h.typeFile[key].Flush()
				i = 0
			}

			s, _ := h.typeFile[key].Size()
			if int(s) >= h.maxSize {
				f := h.typeFile[key]
				h.typeFile[key].Close()
				delete(h.typeFile, key)
				n := f.GetFileName()
				log.Println("Rename:", n, h.tempFilePath+n[5:len(n)-4])
				err = MoveFile(n, h.tempFilePath+n[5:len(n)-4])
				if err != nil {
					log.Println("rename error:", err)
				}
			}
		case <-af.C:
			h.StopAll()
		}

	}
}

func (h *GzipHook) StopAll() {
	// // 关闭回写硬盘
	// h.lock.Lock()
	for k, f := range h.typeFile {
		f.Flush()
		f.Close()
		delete(h.typeFile, k)
		n := f.GetFileName()
		log.Println("Rename:", n, h.tempFilePath+n[5:len(n)-4])
		err := MoveFile(n, h.tempFilePath+n[5:len(n)-4])
		if err != nil {
			log.Println("rename error:", err)
		}
	}
	// h.lock.Unlock()
}

// Levels 只定义 error 和 panic 等级的日志,其他日志等级不会触发 hook
func (h *GzipHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
	}
}

// Fire 将异常日志写入到指定日志文件中
func (h *GzipHook) Fire(entry *logrus.Entry) error {
	fields := make(logrus.Fields)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			fields[k] = v.Error()
		default:
			fields[k] = v
		}
	}
	fields["time"] = entry.Time.Format(time.RFC3339)

	v, ok := entry.Data["message"]
	if ok {
		fields["fields.message"] = v
	}
	fields["message"] = entry.Message

	v, ok = entry.Data["level"]
	if ok {
		entry.Data["fields.level"] = v
	}
	fields["level"] = entry.Level.String()
	if body, err := json.Marshal(fields); err == nil {
		select {
		case h.writeChan <- msg{entry.Data["logtype"].(string), entry.Level.String(), entry.Time, body}:
		default:
			log.Println("logmod.writeChan , error")
			//logrus.Debug("xxxxx")
		}

	} else {
		logrus.Error("xxxxxxx", err)
	}

	return nil
}

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}
