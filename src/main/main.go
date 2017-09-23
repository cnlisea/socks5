package main

import (
	"flag"
	"fmt"
	"ltproxy/src/proxy"
	"ltproxy/src/request"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/astaxie/beego/logs"
)

var logLevel *string = flag.String("log-level", "", "log level [debug|info|warn|error], default error")
var version *bool = flag.Bool("v", false, "the version of ltproxy")
var bindHost *string = flag.String("h", "127.0.0.1", "ltproxy listener host")
var bindPort *int = flag.Int("P", 19926, "ltproxy listener port")
var authUser *string = flag.String("u", "admin", "ltproxy Authentication username")
var authPass *string = flag.String("p", "test", "ltproxy Authentication password")

const banner string = `
ltproxy v1.0.1
`

func main() {
	logs.Info(banner)
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	if *version {
		return
	}

	if 0 == len(*bindHost) || 0 >= *bindPort {
		fmt.Println("must use a right listen host and port")
		return
	}

	// logs settings
	logs.EnableFuncCallDepth(true)
	// logs output filename and logs level default
	logs.SetLogger(logs.AdapterFile, `{"filename":"logs/error.log", "level":6}`)
	if "" != *logLevel {
		setLogLevel(*logLevel)
	}

	lAddr := &proxy.Addr{
		IP:   *bindHost,
		Port: *bindPort,
		Auth: true,
		User: *authUser,
		Pass: *authPass,
	}

	p, err := proxy.NewProxy(lAddr, request.HandleAuthRequest)
	if nil != err {
		logs.Error("New Proxy error:", err)
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGPIPE,
	)

	go func() {
		for {
			sig := <-sc
			if syscall.SIGINT == sig || syscall.SIGTERM == sig || syscall.SIGQUIT == sig {
				logs.Info("main", "main", "Got signal", 0, "signal", sig)
				p.Clone()
			} else if syscall.SIGPIPE == sig {
				logs.Info("main", "main", "Ignore broken pipe signal", 0)
			}
		}
	}()

	p.Run()
}

func setLogLevel(level string) {
	// set default level info
	var l int = logs.LevelInformational
	switch strings.ToLower(level) {
	case "debug":
		l = logs.LevelDebug
	case "info":
		l = logs.LevelInformational
	case "warn":
		l = logs.LevelWarning
	case "error":
		l = logs.LevelError
	}
	logs.SetLevel(l)
}
