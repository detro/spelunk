package echo

import (
	"testing"
)


func TestPing(t *testing.T) {
	got := Ping()
	want := "Pong"
	
	if got != want {
		t.Fatalf("Ping() = %s; want %s", got, want)
	}
}