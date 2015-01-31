package main

import (
	"net"
	"flag"
	"os"
	"io"

	"github.com/zenhack/adb-proxy/common"
)

var (
	addr = flag.String("addr", "", "tcp <host:port> to connect to.")
	verbose = flag.Bool("v", false, "Display error messages on stderr. off by default.")
)

const (
	syncToken = "!!!!!!"
)


func die(status int, msg string, err error) {
	if *verbose {
		os.Stderr.Write([]byte(msg + err.Error()))
	}
	os.Exit(status)
}

func main() {
	flag.Parse()
	n, err := os.Stdout.Write([]byte(syncToken))
	if err != nil || n != len(syncToken) {
		die(1, "Writing sync token: ", err)
		os.Exit(1)
	}
	syncBuf := []byte{0}
	n, err = os.Stdin.Read(syncBuf)
	if err != nil || n != 1 || syncBuf[0] != '!' {
		die(2, "Reading sync token: ", err)
	}
	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		die(3, "Connecting to destination: ", err)
		os.Exit(3)
	}

	dec := common.NewDecoder(os.Stdin, common.ClientStart)
	enc := common.NewEncoder(os.Stdout, common.ServerStart)

	go io.Copy(conn, dec)
	io.Copy(enc, conn)
}
