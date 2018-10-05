package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/codemodus/termsrv/statik"
	"github.com/gorilla/websocket"
	"github.com/hpcloud/tail"
	"github.com/rakyll/statik/fs"
)

func main() {
	ug := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	mq, err := newMsgq()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer mq.close()

	http.HandleFunc("/ws/term", wsHandler(ug, mq))

	sfs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	hfs := http.FileServer(sfs)

	http.Handle("/", hfs)

	go func() {
		t, err := tail.TailFile("/tmp/scriptit", tail.Config{Follow: true})
		if err != nil {
			fmt.Println(err)
			return
		}
		for l := range t.Lines {
			if !mq.send([]byte(l.Text)) {
				fmt.Println("gone!")
			}
		}
	}()

	if err := http.ListenAndServe(":4286", nil); err != nil {
		fmt.Println(err)
	}
}

func logError(msg string, err error) {
	fmt.Fprintf(os.Stderr, msg+": %s\n", err) //nolint
}
