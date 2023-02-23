package socatwrapper

import (
	"context"
	"testing"
	"time"
)

// go test -timeout 30s -run ^TestSocatServer gosocat-wrapper -v -count=1
func TestSocatServer(t *testing.T) {
	server := NewSocatServer(60000, 60000+100)
	port, err := server.StartTunnel(context.WithCancel(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("StartTunnel on port:", port)
	time.Sleep(20 * time.Second)
	server.StopTunnel(port)
}
func TestSocatClient(t *testing.T) {
	client := NewSocatClient("127.0.0.1", 60001, 7681)
	err := client.StartTunnel(context.WithCancel(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Connect to port:", 60001)
	time.Sleep(20 * time.Second)
	client.Stop()
}
func TestProxy(t *testing.T) {
	go TestSocatServer(t)
	time.Sleep(1 * time.Second)
	go TestSocatClient(t)
	time.Sleep(29 * time.Second)
}
