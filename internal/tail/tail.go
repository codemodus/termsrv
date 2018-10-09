package tail

import (
	"fmt"

	"github.com/hpcloud/tail"
)

type Tail struct {
	*tail.Tail
	s   string
	err error
}

func New(file string) (*Tail, error) {
	t, err := tail.TailFile(file, tail.Config{Follow: true})
	if err != nil {
		return nil, err
	}

	return &Tail{Tail: t}, nil
}

func (t *Tail) Scan() bool {
	if t.err != nil {
		return false
	}

	l, ok := <-t.Tail.Lines
	if !ok {
		return false
	}

	if l.Err != nil {
		t.err = fmt.Errorf("cannot get next line: %s", l.Err)
		return false
	}

	t.s = l.Text
	return true
}

func (t *Tail) Bytes() []byte {
	return []byte(t.s)
}

func (t *Tail) Err() error {
	return t.err
}
