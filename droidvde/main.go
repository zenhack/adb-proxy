package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"
)

var (
	bindir = flag.String("bindir", "/data/local/tmp/vde/bin", "Directory containing vde executables")
	socket = flag.String("socket", "/data/local/tmp/vde/socket", "Socket directory for the vde switch.")
	dns = flag.String("dns", "208.67.222.222", "DNS server to offer from dhcp.")
	addr = flag.String("addr", ":8282", "Address on which to listen for vde_plug connections.")
)

func handleConnection(conn net.Conn) {
	flag.Parse()
	defer conn.Close()
	cmd := exec.Command(*bindir + "/vde_plug", *socket)
	out, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	err = cmd.Start()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	go io.Copy(in, conn)
	go io.Copy(conn, out)
	cmd.Wait()
}

func autoRestart(job func() error) {
	for {
		job()
	}
}


func slirpvde() error {
	slirpvde := exec.Command(*bindir + "/slirpvde", "-dhcp", "-N", *dns, *socket)
	err := slirpvde.Start()
	if err != nil {
		time.Sleep(1 * time.Second)
		return err
	}
	return slirpvde.Wait()
}

func termWait(cmd *exec.Cmd) {
	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()
}

func main() {
	ln, err := net.Listen("tcp", ":8282")
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	vde_switch := exec.Command(*bindir + "/vde_switch", "-s", *socket)
	err = vde_switch.Start()
	if err != nil {
		fmt.Println("Error starting vde_switch: ", err)
		return
	}
	defer termWait(vde_switch)
	go autoRestart(slirpvde)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			break
		}
		go handleConnection(conn)
	}
}
