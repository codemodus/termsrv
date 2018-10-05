package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

func wsHandler(ug *websocket.Upgrader, mq *msgq) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cx, err := ug.Upgrade(w, r, nil)
		if err != nil {
			logError("cannot upgrade connection", err)
			return
		}
		defer func() {
			if err = cx.Close(); err != nil {
				logError("cannot close connection", err)
			}
		}()

		cx.SetReadLimit(1)

		done := make(chan struct{})
		defer close(done)

		c, err := mq.attach(done)
		if err != nil {
			logError("cannot attach to message queue", err)
			return
		}

		go func() {
			for v := range c {
				if werr := cx.WriteMessage(websocket.TextMessage, v); werr != nil {
					logError("cannot write to connection", werr)
				}
			}
		}()

		if _, _, rerr := cx.ReadMessage(); rerr != nil {
			if !websocket.IsCloseError(rerr, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				logError("connection close error", rerr)
			}
		}
	}
}
