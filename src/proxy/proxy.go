package proxy

import (
	"net"
	"strconv"

	"github.com/astaxie/beego/logs"
)

type Addr struct {
	//listen ip address
	IP string

	//listen port address
	Port int

	// Authentication status
	Auth bool

	// Authentication username
	User string

	// Authentication password
	Pass string
}

type Proxy struct {
	// local Listen ip and port address
	addr *Addr

	// connections handler function
	handle func(net.Conn, string, string)

	// listener
	listener net.Listener

	// Authentication username
	user string
	// Authentication password
	passwd string

	running bool
}

func NewProxy(lAddr *Addr, handle func(net.Conn, string, string)) (*Proxy, error) {
	p := new(Proxy)

	p.addr = lAddr
	p.handle = handle
	p.user = lAddr.User
	p.passwd = lAddr.Pass

	var err error
	netProto := "tcp"
	addr := net.JoinHostPort(p.addr.IP, strconv.Itoa(p.addr.Port))

	p.listener, err = net.Listen(netProto, addr)
	if nil != err {
		return nil, err
	}

	logs.Info("server", "NewServer", "Server running", 0,
		"netProto",
		netProto,
		"address",
		addr,
	)

	return p, nil
}

func (p *Proxy) Run() error {
	p.running = true

	// start accept connection
	for p.running {
		c, err := p.listener.Accept()
		if nil != err {
			logs.Error("server", "run", err.Error(), 0)
			continue
		}

		//
		go p.handle(c, p.user, p.passwd)
	}

	return nil
}

func (p *Proxy) Clone() {
	p.running = false
	if nil != p.listener {
		p.listener.Close()
	}
}
