package main

import (
	"log"
	"mynet/netmodule"
)

//Mgr 管理连接
type Mgr struct {
}

//OnConnect 新建连接
func (m *Mgr) OnConnect(s netmodule.Socket) netmodule.Session {
	return nil
}

func main() {
	var mgr Mgr
	if err := netmodule.ListenAndAccept("127.0.0.1:8088", &mgr); err != nil {
		log.Fatalln(err)
	}
}
