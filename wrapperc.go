package socatwrapper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type SocatClient struct {
	serverHost  string      // 远程服务器
	remotePort  uint        // 远程端口
	localPort   uint        // 本地端口
	socatTunnel socatTunnel // 正在打通的隧道进程
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewSocatClient(host string, rp, lp uint) *SocatClient {
	SocatClient := new(SocatClient)
	SocatClient.serverHost = host
	SocatClient.remotePort = rp
	SocatClient.localPort = lp
	return SocatClient
}

/*
*
* 启动一个通道
*
 */
func (client *SocatClient) StartTunnel(ctx context.Context, cancel context.CancelFunc) error {
	client.ctx = ctx
	client.cancel = cancel
	if runtime.GOOS != "linux" {
		return fmt.Errorf("not support current os:%s, only support linux at now", runtime.GOOS)
	}
	_, err := exec.LookPath("socat")
	if err != nil {
		return err
	}
	cmdStr := "socat tcp:%s:%d,forever,intervall=5,fork tcp:localhost:%d"
	shellCmd := exec.CommandContext(ctx, fmt.Sprintf(cmdStr, client.serverHost, client.remotePort, client.localPort))
	shellCmd.Stdout = os.Stdout
	shellCmd.Stderr = os.Stderr
	client.socatTunnel = socatTunnel{
		ctx:      ctx,
		cancel:   cancel,
		shellCmd: shellCmd,
	}
	if err := client.socatTunnel.shellCmd.Start(); err != nil {
		return err
	}
	go func(cmd *exec.Cmd) {
		cmd.Process.Wait() // blocked until exited
	}(client.socatTunnel.shellCmd)
	return nil
}

/*
*
* 停止隧道
*
 */
func (client *SocatClient) Stop() error {
	if client.socatTunnel.shellCmd != nil {
		client.socatTunnel.shellCmd.Process.Kill()
		client.cancel()
		return nil
	}
	return fmt.Errorf("tunnel not exists")

}

/*
*
* 获取当前所有的隧道
*
 */
func (client *SocatClient) Tunnel() socatTunnel {
	return client.socatTunnel
}
