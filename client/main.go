package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"

	"github.com/zenhack/adb-proxy/common"
)

var (
	addr = flag.String("addr", "", "tcp <host:port> to connect to.")
	laddr = flag.String("laddr", ":7000",  "tcp address to listen on.")
)

func check(ctx string, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, ctx, err)
		os.Exit(1)
	}
}

func doSync(r io.Reader) error {
	buf := []byte{0}
	count := 0
	for count < 6 {
		_, err := io.ReadAtLeast(r, buf, 1)
		if err != nil {
			return err
		}
		if buf[0] == '!' {
			count++
		} else {
			count = 0
		}
	}
	return nil
}

func main() {
	flag.Parse()
	cmd := exec.Command("adb", "shell")
	stdin, err := cmd.StdinPipe()
	check("StdinPipe: ", err)
	stdout, err := cmd.StdoutPipe()
	check("StdoutPipe: ", err)
	err = cmd.Start()
	check("Start: ", err)

	stdin.Write([]byte(
		"cd /data/local/tmp\n" +
		"exec /data/local/tmp/server -addr " + *addr + " 2> err.log\n"))
	check("doSync: ", doSync(stdout))

	ln, err := net.Listen("tcp", *laddr)
	check("Listen: ", err)

	conn, err := ln.Accept()
	check("Accept: ", err)

	_, err = stdin.Write([]byte{'!'})
	check("final sync: ", err)

	dec := common.NewDecoder(stdout, common.ServerStart)
	enc := common.NewEncoder(stdin, common.ClientStart)

	go common.LoudCopy(enc, conn)
	common.LoudCopy(conn, dec)
}
