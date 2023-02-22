package socatwrapper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type SocatServer struct {
	// 端口池，默认最多连接1000个客户端, 建议使用大数空闲端口段
	PortPool     [1000]uint           // 当数组的某处为0的时候表示是空闲的
	portBegin    uint                 // 端口开始范围
	portEnd      uint                 // 端口结束范围
	socatTunnels map[uint]socatTunnel // 正在打通的隧道进程
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewSocatServer(b, e uint) *SocatServer {
	PortPool := [1000]uint{0}
	SocatServer := new(SocatServer)
	for i := b; i < e-b; i++ {
		PortPool[i] = uint(i)
	}
	SocatServer.PortPool = PortPool
	SocatServer.portBegin = b
	SocatServer.portEnd = e
	return SocatServer
}

/*
*
* 启动一个通道
*
 */
func (server *SocatServer) StartTunnel(ctx context.Context, cancel context.CancelFunc) (uint, error) {
	server.ctx = ctx
	server.cancel = cancel
	if runtime.GOOS != "linux" {
		return 0, fmt.Errorf("not support current os:%s, only support linux at now", runtime.GOOS)
	}
	_, err := exec.LookPath("socat")
	if err != nil {
		return 0, err
	}
	cmdStr := "socat tcp-l:%d,reuseaddr,bind=0.0.0.0,fork tcp-l:%d,reuseaddr,bind=0.0.0.0,retry=10"
	var Port uint = 0
	for _, port := range server.PortPool {
		if port != 0 {
			Port = port
			break
		}
	}
	if Port == 0 {
		return 0, fmt.Errorf("port pool overflow")
	}
	shellCmd := exec.CommandContext(ctx, fmt.Sprintf(cmdStr, Port, Port))
	shellCmd.Stdout = os.Stdout
	shellCmd.Stderr = os.Stderr
	server.socatTunnels[Port] = socatTunnel{
		ctx:      ctx,
		cancel:   cancel,
		shellCmd: shellCmd,
	}
	if err := server.socatTunnels[Port].shellCmd.Start(); err != nil {
		return 0, err
	}
	go func(cmd *exec.Cmd) {
		cmd.Process.Wait() // blocked until exited
	}(server.socatTunnels[Port].shellCmd)
	return (Port), nil
}

/*
*
* 停止隧道
*
 */
func (server *SocatServer) StopTunnel(port uint) error {
	if server.socatTunnels[port].shellCmd != nil {
		server.socatTunnels[port].shellCmd.Process.Kill()
		server.cancel()
		server.PortPool[port] = 0
		delete(server.socatTunnels, port)
		return nil
	}
	return fmt.Errorf("tunnel not exists")

}

/*
*
* 获取当前所有的隧道
*
 */
func (server *SocatServer) AllTunnel() map[uint]socatTunnel {
	return server.socatTunnels
}
