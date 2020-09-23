package dbcache

import (
	"context"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	autoUpdateList = []UpdateMode{}
)

type UpdateMode interface {
	Update() error
}

func Add(obj ...UpdateMode) {
	autoUpdateList = append(autoUpdateList, obj...)
}

func AutoUpdate(ctx context.Context, timeout time.Duration) {
	tick := time.NewTicker(timeout)
	defer tick.Stop()
	update := func() {
		for _, r := range autoUpdateList {
			go r.Update()
		}
	}
	update()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			go update()
		}
	}
}

func ConnectDB(addr string) (*sqlx.DB, error) {
	// 带链接池
	db, err := sqlx.Connect("mysql", addr)
	if err != nil {
		return nil, err
	}
	//db.DB.SetMaxIdleConns(32)
	db.DB.SetMaxOpenConns(20)
	db.DB.SetConnMaxLifetime(0)
	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return db, err

}
