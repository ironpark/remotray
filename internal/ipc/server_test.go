package ipc

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestNewServer(t *testing.T) {
	s, _ := NewServer("servertest")
	s.SetMessageProcessor(1, func(data []byte) (interface{}, error) {
		return map[string]string{
			"a": "b",
		}, nil
	})
	go s.Run()
	client, _ := NewClient("servertest")
	id, _ := client.WriteMessage(1, map[string]string{
		"a": "b",
	})
	aa := map[string]string{}
	client.ReadReplyMessage(id, &aa)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
