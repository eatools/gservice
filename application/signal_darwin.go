package application

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/eatools/gservice/onstop"
)

func ListenSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1,
		syscall.SIGUSR2, syscall.SIGTSTP)
	select {
	case <-sigs:
		fmt.Println("exitapp,sigs:", sigs)
		onstop.Exit()
		fmt.Println("exitapp,success!!!")
		os.Exit(0)
	}
}
