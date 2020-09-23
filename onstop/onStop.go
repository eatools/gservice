package onstop

import "context"

var all = []func(){}

var Context, CAN = context.WithCancel(context.Background())

func Append(f func()) {
	all = append(all, f)
}

func Exit() {
	CAN()
	for _, fun := range all {
		fun()
	}
}
