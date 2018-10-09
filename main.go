package main

import (
	"net/http"
	"os"
	"path"

	"github.com/codemodus/sigmon"
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

	sm.Set(es.term)

	go func() {
		if err := es.mq.Feed(es.t); err != nil {
			logError("feed failure", err)
		}
	}()

	logInfof("serving on %s\n", es.srv.Addr)
	if err := es.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	logInfof("goodbye\n")
	return nil
}
