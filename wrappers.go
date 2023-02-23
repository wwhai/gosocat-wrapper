package socatwrapper

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"log"
)

type SocatServer struct {
	// 端口池，默认最多连接1000个客户端, 建议使用大数空闲端口段
	PortPool     [1000]uint           // 当数组的某处为0的时候表示是空闲的
	portBegin    uint                 // 端口开始范围
	portEnd      uint                 // 端口结束范围
	socatTunnels map[uint]socatTunnel // 正在打通的隧道进程
	ctx          context.Context
	cancel       context.CancelFunc
	out          io.Writer
	err          io.Writer
}

func NewSocatServer(b, e uint) *SocatServer {
	PortPool := [1000]uint{0}
	SocatServer := new(SocatServer)
	for i := 0; i < int(e-b); i++ {
		PortPool[int(i)] = uint(b + uint(i))
	}
	SocatServer.PortPool = PortPool
	SocatServer.portBegin = b
	SocatServer.portEnd = e
	SocatServer.socatTunnels = make(map[uint]socatTunnel)
	SocatServer.out = &logOutFilter{}
	SocatServer.err = &logErrFilter{}
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
	// cmdStr := "tcp-l:%d,reuseaddr,bind=0.0.0.0,fork tcp-l:%d,reuseaddr,bind=0.0.0.0,retry=10"
	c1 := "tcp-l:%d,reuseaddr,bind=0.0.0.0,fork"
	c2 := "tcp-l:%d,reuseaddr,bind=0.0.0.0,retry=10"
	shellCmd := exec.CommandContext(ctx, "socat", "-d", "-d", "-d",
		fmt.Sprintf(c1, Port), fmt.Sprintf(c2, Port+1))
	shellCmd.Stdout = server.out
	shellCmd.Stderr = server.err
	server.socatTunnels[Port] = socatTunnel{
		ctx:      ctx,
		cancel:   cancel,
		shellCmd: shellCmd,
	}
	if err := server.socatTunnels[Port].shellCmd.Start(); err != nil {
		return 0, err
	}
	log.Println("Start:", server.socatTunnels[Port].shellCmd.String())
	go func(cmd *exec.Cmd) {
		cmd.Process.Wait() // blocked until exited
	}(server.socatTunnels[Port].shellCmd)
	// Port + 1 表示一对端口的第二个作为桥接端口用
	// 客户端最终连接的是这个桥接端口
	return (Port + 1), nil
}

/*
*
* 停止隧道
*
 */
func (server *SocatServer) StopTunnel(port uint) error {
	// port - 1 : 表示服务端口
	ListenPort := port - 1
	if server.socatTunnels[ListenPort].shellCmd != nil {
		server.socatTunnels[ListenPort].shellCmd.Process.Kill()
		server.socatTunnels[ListenPort].shellCmd.Process.Signal(os.Kill)
		server.cancel()
		for i, v := range server.PortPool {
			if v == ListenPort {
				server.PortPool[i] = 0
				break
			}
		}
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

/*
*
* 获取当前通道的状态
*
 */
func (server *SocatServer) State() {

}
