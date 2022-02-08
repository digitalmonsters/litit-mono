package boilerplate_testing

import "testing"

func TestLock(t *testing.T) {
	l := getLock()

	if err := l.Lock(); err != nil {
		t.Fatal(err)
	}

	if err := l.Unlock(); err != nil {
		t.Fatal(err)
	}
}
