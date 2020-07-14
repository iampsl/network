package netmodule

import (
	"net"
	"sync"
	"sync/atomic"
)

//tcpsocket 对net.Conn 的包装
type tcpsocket struct {
	conn      net.Conn   //TCP底层连接
	buffers   [2]*buffer //双发送缓存
	sendIndex uint       //发送缓存索引
	notify    chan int   //通知通道
	isclose   uint32     //指示socket是否关闭

	m          sync.Mutex //锁
	bclose     bool       //是否关闭
	writeIndex uint       //插入缓存索引
}

//newtcpsocket 创建一个tcpsocket
func newtcpsocket(c net.Conn) *tcpsocket {
	if c == nil {
		//c为nil,抛出异常
		panic("c is nil")
	}
	//初始化结构体
	var psocket = new(tcpsocket)
	psocket.conn = c
	psocket.buffers[0] = new(buffer)
	psocket.buffers[1] = new(buffer)
	psocket.sendIndex = 0
	psocket.notify = make(chan int, 1)
	psocket.isclose = 0
	psocket.bclose = false
	psocket.writeIndex = 1
	//启动发送协程
	go psocket._dosend()
	return psocket
}

func (my *tcpsocket) _dosend() {
	writeErr := false
	for {
		_, ok := <-my.notify
		if !ok {
			return
		}
		my.m.Lock()
		my.writeIndex = my.sendIndex
		my.m.Unlock()
		my.sendIndex = (my.sendIndex + 1) % 2
		if !writeErr {
			var sendSplice = my.buffers[my.sendIndex].Data()
			for len(sendSplice) > 0 {
				n, err := my.conn.Write(sendSplice)
				if err != nil {
					writeErr = true
					break
				}
				sendSplice = sendSplice[n:]
			}
		}
		my.buffers[my.sendIndex].Clear()
	}
}

//Read 读数据
func (my *tcpsocket) Read(b []byte) (n int, err error) {
	return my.conn.Read(b)
}

//WriteBytes 写数据
func (my *tcpsocket) Write(b ...[]byte) {
	my.m.Lock()
	if my.bclose {
		my.m.Unlock()
		return
	}
	dataLen := my.buffers[my.writeIndex].Len()
	writeLen := 0
	for i := 0; i < len(b); i++ {
		writeLen += len(b[i])
		my.buffers[my.writeIndex].Append(b[i])
	}
	if dataLen == 0 && writeLen != 0 {
		my.notify <- 0
	}
	my.m.Unlock()
}

//Close 关闭一个tcpsocket, 释放系统资源
func (my *tcpsocket) Close() {
	my.m.Lock()
	if my.bclose {
		my.m.Unlock()
		return
	}
	my.bclose = true
	my.conn.Close()
	close(my.notify)
	my.m.Unlock()
	atomic.StoreUint32(&(my.isclose), 1)
}

//IsClose 判断tcpsocket是否关闭
func (my *tcpsocket) IsClose() bool {
	val := atomic.LoadUint32(&(my.isclose))
	if val > 0 {
		return true
	}
	return false
}
