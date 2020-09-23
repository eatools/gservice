package logmod

import (
	"fmt"
	"testing"

	"github.com/eatools/gservice/application/onstop"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {

	InitLog()
	m.Run()
	fmt.Println(".....")
	onstop.Exit()
}

func Test_log(t *testing.T) {
	Type(LogBidReq).WithFields(logrus.Fields{
		"aa":   1,
		"bb":   "2",
		"ccc":  map[string]string{"a": "1"},
		"list": []string{"bidder1", "bidder2"},
	}).Info("xxxx")
}
