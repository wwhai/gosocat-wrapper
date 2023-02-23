package socatwrapper

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
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
	connected   bool
	out         io.Writer
	err         io.Writer
}

func NewSocatClient(host string, rp, lp uint) *SocatClient {
	SocatClient := new(SocatClient)
	SocatClient.serverHost = host
	SocatClient.remotePort = rp
	SocatClient.localPort = lp
	SocatClient.out = &logOutFilter{}
	SocatClient.err = &logErrFilter{}
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
	if !client.checkNetworkAccess() {
		return fmt.Errorf("server unavailable, %s:%d", client.serverHost, client.remotePort)
	}
	// "tcp:%s:%d,forever,intervall=5,fork tcp:localhost:%d"
	c1 := "tcp:%s:%d,forever,intervall=5,fork"
	c2 := "tcp:localhost:%d"
	shellCmd := exec.CommandContext(ctx, "socat", "-d", "-d", "-d",
		fmt.Sprintf(c1, client.serverHost, client.remotePort),
		fmt.Sprintf(c2, client.localPort))
	shellCmd.Stdout = client.out
	shellCmd.Stderr = client.err
	client.socatTunnel = socatTunnel{
		ctx:      ctx,
		cancel:   cancel,
		shellCmd: shellCmd,
	}
	if err := client.socatTunnel.shellCmd.Start(); err != nil {
		return err
	}
	log.Println("Start:", client.socatTunnel.shellCmd.String())
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
		client.socatTunnel.shellCmd.Process.Signal(os.Kill)
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

/*
*
* 获取当前状态
*
 */
func (client *SocatClient) Connected() bool {
	return client.checkNetworkAccess()
}

//--------------------------------------------------------------------------------------------------
// 内部函数
//--------------------------------------------------------------------------------------------------
/*
*
* 网络是否可达
*
 */
func (client *SocatClient) checkNetworkAccess() bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.serverHost, client.remotePort-1))
	if err != nil {
		client.connected = false
		return false
	}
	conn.Close()
	client.connected = true
	return true
}
