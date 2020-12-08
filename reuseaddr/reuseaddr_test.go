package reuseaddr

import (
	"context"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestReUseAddr(t *testing.T) {
	listenConfig := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if err != nil {
					return
				}
				err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if err != nil {
					return
				}
			})
			return err
		},
		KeepAlive: 0,
	}

	go func() {
		server1 := http.ServeMux{}
		server1.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte("server1"))
		})

		conn, err := listenConfig.Listen(context.Background(), "tcp", "127.0.0.1:12009")
		if err != nil {
			log.Fatal(err)
		}
		err = http.Serve(conn, &server1)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		server2 := http.ServeMux{}
		server2.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte("server2"))
		})

		time.Sleep(time.Second)

		conn, err := listenConfig.Listen(context.Background(), "tcp", "127.0.0.1:12009")
		if err != nil {
			log.Fatal(err)
		}
		err = http.Serve(conn, &server2)
		if err != nil {
			log.Fatal(err)
		}
	}()

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGABRT, syscall.SIGUSR1, syscall.SIGKILL)
	<-signals
}
