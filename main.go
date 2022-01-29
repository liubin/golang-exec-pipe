package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "demo" {
		externalCmd()
	} else if len(os.Args) == 2 && os.Args[1] == "self-pipe" {
		monitorSelfPipe()
	} else {
		monitor()
	}
}

func monitorSelfPipe() {

	cmd := exec.Command("./golang-exec-pipe", "demo")

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	cmd.Stdout = w

	stdoutReader := bufio.NewReader(r)

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	pid := cmd.Process.Pid
	fmt.Println("new process ID: ", pid)
	handleReader(stdoutReader)
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				fmt.Printf("wait exit status: %+v\n", err)
			} else {
				fmt.Printf("exit status: %+v\n", status)
			}
		}
		fmt.Printf("wait exit status: %+v\n", err)
	}
}

func monitor() {

	cmd := exec.Command("./golang-exec-pipe", "demo")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	defer stdout.Close()
	stdoutReader := bufio.NewReader(stdout)

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	pid := cmd.Process.Pid
	fmt.Println("new process ID: ", pid)
	handleReader(stdoutReader)
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				fmt.Printf("wait exit status: %+v\n", err)
			} else {
				fmt.Printf("exit status: %+v\n", status)
			}
		}
		fmt.Printf("wait exit status: %+v\n", err)
	}
}

func handleReader(reader *bufio.Reader) {
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("process exit: %+v", err)
			break
		}
		fmt.Print(str)
	}
}

func externalCmd() {
	i := 0
	for {
		i++
		time.Sleep(time.Second * 1)
		fmt.Fprintf(os.Stdout, "stdout %d\n", i)
		fmt.Fprintf(os.Stderr, "stderr %d\n", i)

	}
}
