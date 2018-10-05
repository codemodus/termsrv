package main

import (
	"context"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/codemodus/sigmon"
	_ "github.com/codemodus/termsrv/statik"
	"github.com/hpcloud/tail"
)

func main() {
	if err := run(); err != nil {
		cmd := path.Base(os.Args[0])
		logError(cmd, err)
		os.Exit(1)
	}
}

func run() error {
	sm := sigmon.New(nil)
	sm.Start()
	defer sm.Stop()

	es, err := newElements()
	if err != nil {
		return err
	}
	defer es.close()

	sm.Set(func(s *sigmon.State) {
		scp := "while handling a system signal"

		if err := es.t.Stop(); err != nil {
			logError(scp, err)
		}

		sc, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if err := es.srv.Shutdown(sc); err != nil {
			logError(scp, err)
		}
	})

	go feedQ(es.mq, es.t)

	if err := es.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func feedQ(mq *msgq, t *tail.Tail) {
	for l := range t.Lines {
		if !mq.send([]byte(l.Text)) {
			logError("mq is gone!", nil)
		}
	}
}
