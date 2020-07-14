package netmodule

import (
	"log"
	"net"
)

//ListenAndAccept 接受新的连接，将阻塞
func ListenAndAccept(address string, a Accept) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		} else {
			psocket := newtcpsocket(tcpConn)
			session := a.OnConnect(psocket)
			if session == nil {
				psocket.Close()
			} else {
				go onSocket(psocket, session)
			}
		}
	}
}

func onSocket(psocket *tcpsocket, session Session) {
	defer psocket.Close()
	readbuffer := make([]byte, 1024)
	readsize := 0
	for {
		if readsize == len(readbuffer) {
			pnew := make([]byte, 2*len(readbuffer))
			copy(pnew, readbuffer)
			readbuffer = pnew
		}
		n, err := psocket.Read(readbuffer[readsize:])
		if err != nil {
			session.OnError(err)
			break
		}
		readsize += n
		procTotal := 0
		for {
			if psocket.IsClose() {
				procTotal = readsize
				break
			}
			proc := session.OnRead(readbuffer[procTotal:readsize])
			if proc == 0 {
				break
			}
			procTotal += proc
		}
		if procTotal > 0 {
			copy(readbuffer, readbuffer[procTotal:readsize])
			readsize -= procTotal
		}
	}
}
