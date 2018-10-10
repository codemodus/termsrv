package main

import (
	"context"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/codemodus/sigmon/v2"
	_ "github.com/codemodus/termsrv/statik"
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
		scp := "system signal handling abnormal"

		sc, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if err := es.srv.Shutdown(sc); err != nil {
			logError(scp, err)
		}

		sm.Stop()
	})

	go func() {
		if err := es.mq.Feed(es.t); err != nil {
			scp := "feed goroutine abnormal"
			logError(scp, err)

			sm.Stop()

			sc, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			if err := es.srv.Shutdown(sc); err != nil {
				logError(scp, err)
			}
		}
	}()

	logInfof("serving on %s\n", es.srv.Addr)
	defer logInfof("server on %s is closed\n", es.srv.Addr)

	if err := es.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
