package main

import (
	"sync"
)

type bond struct {
	c    chan []byte
	done chan struct{}
}

type msgq struct {
	c    chan []byte
	mu   *sync.Mutex
	q    []bond
	done chan struct{}
}

func newMsgq() (*msgq, error) {
	mq := msgq{
		c:    make(chan []byte),
		mu:   &sync.Mutex{},
		q:    make([]bond, 0),
		done: make(chan struct{}),
	}

	go func() {
		defer close(mq.c)

		for {
			select {
			case v := <-mq.c:
				mq.distribute(v)
			case <-mq.done:
				return
			}
		}
	}()

	return &mq, nil
}

func (mq *msgq) close() {
	close(mq.done)
}

func (mq *msgq) send(bs []byte) bool {
	select {
	case <-mq.done:
		return false
	default:
		mq.c <- bs
		return true
	}
}

func (mq *msgq) distribute(v []byte) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	for _, b := range mq.q {
		b.c <- v
	}
}

func (mq *msgq) attach(done chan struct{}) (chan []byte, error) {
	b := bond{
		c:    make(chan []byte),
		done: done,
	}

	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.q = append(mq.q, b)
	i := len(mq.q) - 1

	go func() {
		defer close(b.c)

		<-done

		func() {
			mq.mu.Lock()
			defer mq.mu.Unlock()

			mq.q = append(mq.q[:i], mq.q[i+1:]...)
		}()
	}()

	return b.c, nil
}
