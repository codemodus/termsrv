package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/codemodus/sigmon"
	"github.com/codemodus/termsrv/internal/msgq"
	"github.com/codemodus/termsrv/internal/tail"
	"github.com/codemodus/veva"
	"github.com/gorilla/websocket"
	"github.com/rakyll/statik/fs"
)

type elements struct {
	mq   *msgq.Msgq
	t    *tail.Tail
	srv  *http.Server
	done chan struct{}
}

func newElements() (*elements, error) {
	done := make(chan struct{})

	fin := func(es *elements, err error) (*elements, error) {
		if err != nil {
			safeClose(done)
		}

		return es, err
	}

	sfs, err := fs.New()
	if err != nil {
		return fin(nil, err)
	}
	fs := http.FileServer(sfs)

	ug := newWebsocketUpgrader()

	mq, err := msgq.New()
	if err != nil {
		return fin(nil, err)
	}
	go func() {
		<-done
		mq.Close()
	}()

	t, err := tail.New("/tmp/scriptit")
	if err != nil {
		return fin(nil, err)
	}
	go func() {
		<-done
		t.Cleanup()
	}()

	mux := http.NewServeMux()
	mux.Handle("/", fs)
	mux.HandleFunc("/ws/term", wsHandler(ug, mq))

	srv, err := newHTTPServer("0.0.0.0", ":4286", mux)
	if err != nil {
		return fin(nil, err)
	}

	es := elements{
		mq:   mq,
		t:    t,
		srv:  srv,
		done: done,
	}

	return &es, nil
}

func (es *elements) close() {
	safeClose(es.done)
}

func (es *elements) term(s *sigmon.State) {
	scp := "while handling a system signal"

	if err := es.t.Stop(); err != nil {
		logError(scp, err)
	}

	sc, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := es.srv.Shutdown(sc); err != nil {
		logError(scp, err)
	}
}

func newWebsocketUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

func newHTTPServer(host, port string, h http.Handler) (*http.Server, error) {
	we := func(err error) error {
		return fmt.Errorf("cannot create new http server: %s", err)
	}

	p, err := veva.Port(port)
	if err != nil {
		return nil, we(err)
	}

	s := http.Server{
		Addr:    host + p,
		Handler: h,
	}

	return &s, nil
}
