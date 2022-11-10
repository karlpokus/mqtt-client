package stream

import (
	"errors"
	"testing"
)

// This test asserts that by subscribing, res will first
// produce a subscription ack, then an error
func TestClient(t *testing.T) {
	req, res := Client(&fakeStream{})
	go func() {
		req <- Subscribe("test")
	}()
	r := <-res
	got := r.Message
	want := "subscription acked" // fragile
	if got != want {
		t.Fatalf("%s does not match %s", got, want)
	}
	r = <-res
	if !errors.Is(r.Err, ErrConnClosed) {
		t.Fatalf("%s is not %s", r.Err, ErrConnClosed)
	}
}
