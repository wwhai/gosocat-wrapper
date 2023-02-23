package socatwrapper

import (
	"context"
	"log"
	"testing"
	"time"
)

// go test -timeout 30s -run ^TestSocatServer gosocat-wrapper -v -count=1
func _SocatServer() {
	server := NewSocatServer(60000, 60000+100)
	port, err := server.StartTunnel(context.WithCancel(context.Background()))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("StartTunnel on port:", port)
	time.Sleep(20 * time.Second)
	server.StopTunnel(port)
}
func _SocatClient() {
	client := NewSocatClient("127.0.0.1", 60001, 7681)
	err := client.StartTunnel(context.WithCancel(context.Background()))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connect to port:", 60001)
	time.Sleep(20 * time.Second)
	client.Stop()
}
func TestProxy(t *testing.T) {
	go _SocatServer()
	time.Sleep(1 * time.Second)
	go _SocatClient()
	time.Sleep(29 * time.Second)
}
