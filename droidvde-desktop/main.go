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
	bindir = flag.String("bindir", "/data/local/tmp/vde/bin", "Directory with vde executables.")
)

func termWait(cmd *exec.Cmd) {
	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()
}

func mainLoop() (err error) {
	droidvde := exec.Command("adb", "shell", *bindir + "/droidvde")
	vde_plug := exec.Command("vde_plug", "/tmp/vde0")
	exec.Command("adb", "wait-for-device").Run()
	exec.Command("adb", "forward", "tcp:8282", "tcp:8282").Run()
	err = droidvde.Start()
	if err != nil {
		fmt.Println("Error launching droidvde: ", err)
		return
	}
	defer termWait(droidvde)
	time.Sleep(1 * time.Second)
	conn, err := net.Dial("tcp", ":8282")
	if err != nil {
		fmt.Println("Error connecting: ", err)
		return
	}
	in, err := vde_plug.StdinPipe()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	out, err := vde_plug.StdoutPipe()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	err = vde_plug.Start()
	if err != nil {
		fmt.Println("Error launching vde_plug: ", err)
		return
	}
	defer termWait(vde_plug)
	go io.Copy(in, conn)
	io.Copy(conn, out)
	return nil
}

func main() {
	flag.Parse()
	for {
		err := mainLoop()
		if err != nil {
			os.Exit(1)
		}
	}
}
