package socatwrapper

import (
	"context"
	"testing"
	"time"
)

// go test -timeout 30s -run ^TestSocatServer github.com/i4de/rulex/test -v -count=1
func TestSocatServer(t *testing.T) {
	server := NewSocatServer(60000, 60000+100)
	port, err := server.StartTunnel(context.WithCancel(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(28 * time.Second)
	server.StopTunnel(port)
}
func TestSocatClient(t *testing.T) {
	client := NewSocatClient("127.0.0.1", 60000, 2580)
	err := client.StartTunnel(context.WithCancel(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(28 * time.Second)
	client.Stop()
}
