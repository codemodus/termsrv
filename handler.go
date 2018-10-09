package main

import (
	"net/http"

	"github.com/codemodus/termsrv/internal/msgq"
	"github.com/gorilla/websocket"
)

func wsHandler(ug *websocket.Upgrader, mq *msgq.Msgq) http.HandlerFunc {
	scp := "websocket endpoint malfunction"

	return func(w http.ResponseWriter, r *http.Request) {
		cx, err := ug.Upgrade(w, r, nil)
		if err != nil {
			logError(scp, err)
			return
		}
		defer func() {
			if err = cx.Close(); err != nil {
				logError(scp, err)
			}
		}()

		cx.SetReadLimit(1)

		done := make(chan struct{})
		defer close(done)

		c, err := mq.Attach(done)
		if err != nil {
			logError(scp, err)
			return
		}

		go func() {
			for v := range c {
				if werr := cx.WriteMessage(websocket.TextMessage, v); werr != nil {
					logError(scp, err)
				}
			}
		}()

		if _, _, rerr := cx.ReadMessage(); rerr != nil {
			if !websocket.IsCloseError(rerr, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				logError(scp, err)
			}
		}
	}
}
