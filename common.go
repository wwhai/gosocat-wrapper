package socatwrapper

import (
	"context"
	"os/exec"
)

type socatTunnel struct {
	shellCmd *exec.Cmd
	ctx      context.Context
	cancel   context.CancelFunc
}
