package request

import (
	"io"
	"net"
	"strconv"

	"github.com/astaxie/beego/logs"
)

func HandleAuthRequest(c net.Conn, user, passwd string) {
	if nil == c {
		return
	}
	defer c.Close()

	var b [1024]byte
	n, err := c.Read(b[:])
	if nil != err {
		logs.Error(err)
		return
	}

	//sock5
	if b[0] == 0x05 {
		// response proxy ack, auth username/passwd
		c.Write([]byte{0x05, 0x02})

		// get request body
		n, err = c.Read(b[:])

		var u, p string
		// get auth username/passwd
		uLen := int(b[1])
		u = string(b[2 : uLen+2])
		pLen := int(b[uLen+2])
		p = string(b[uLen+3 : uLen+pLen+3])

		if u != user || p != passwd {
			// reponse auth failu
			c.Write([]byte{0x01, 0x01})
			return
		}

		// reponse autu success
		c.Write([]byte{0x01, 0x00})

		n, err = c.Read(b[:])
		var host, port string
		switch b[3] {
		case 0x01: //IP V4
			host = net.IPv4(b[4], b[5], b[6], b[7]).String()
		case 0x03: //域名
			host = string(b[5 : n-2]) //b[4]表示域名的长度
		case 0x04: //IP V6
			host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
		}
		port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))

		addr := net.JoinHostPort(host, port)
		netProto := "tcp"

		server, err := net.Dial(netProto, addr)
		if nil != err {
			logs.Error(err)
			return
		}
		defer server.Close()

		// 响应客户端连接成功
		c.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		go io.Copy(server, c)
		io.Copy(c, server)
	}
}
