package socatwrapper

import (
	"context"
	"log"
	"os/exec"
)

type logOutFilter struct {
}

func (*logOutFilter) Write(p []byte) (n int, err error) {
	log.Println("logOutFilter ========>", string(p))
	return 0, nil

}

type logErrFilter struct {
}

func (*logErrFilter) Write(p []byte) (n int, err error) {
	log.Println("logErrFilter ========>", string(p))
	return 0, nil

}

type socatTunnel struct {
	shellCmd *exec.Cmd
	ctx      context.Context
	cancel   context.CancelFunc
}
