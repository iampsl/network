package netmodule

//Socket 套接字接口
type Socket interface {
	Write(...[]byte)
	Close()
}

//Session 数据包处理接口
type Session interface {
	OnRead([]byte) int
	OnError(error)
}

//Accept 接受新的连接
type Accept interface {
	OnConnect(s Socket) Session
}
