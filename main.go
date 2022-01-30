package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var quitCh chan bool

func main() {

	quitCh = make(chan bool)

	if len(os.Args) == 2 && os.Args[1] == "demo" {
		externalCmd()
	} else if len(os.Args) == 2 && os.Args[1] == "self-pipe" {
		monitorSelfPipe()
	} else if len(os.Args) == 2 && os.Args[1] == "combined" {
		monitorCombinedOutput()
	} else {
		monitor()
	}

	<-quitCh
}

// case 1: get stdout/stderr of sub-process
// by using cmd.StdoutPipe() and redirect stderr to stdout
func monitorCombinedOutput() {
	cmd := exec.Command("./golang-exec-pipe", "demo")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	cmd.Stderr = cmd.Stdout

	runAndWait(cmd, stdout)
}


// case 2: use self created pipe as stdout/stderr
// in this case the handleReader can't get sub process quit event
// because the reader is not closed by exec.Cmd
func monitorSelfPipe() {
	cmd := exec.Command("./golang-exec-pipe", "demo")

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	cmd.Stdout = w
	cmd.Stderr = w

	runAndWait(cmd, r)
}

func monitor() {
	cmd := exec.Command("./golang-exec-pipe", "demo")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	runAndWait(cmd, stdout)
}

func runAndWait(cmd *exec.Cmd, stdout io.ReadCloser) {
	stdoutReader := bufio.NewReader(stdout)

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	pid := cmd.Process.Pid
	fmt.Println("new sub process: ", pid)

	go kill(cmd)

	handleReader(stdoutReader)
	fmt.Printf("sub process %d quit\n", pid)

	wait(cmd)
}

func handleReader(reader *bufio.Reader) {
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("monitor process exit: %+v\n", err)
			break
		}
		fmt.Print(str)
	}
}

// kill sub process after 3 seconds
func kill(cmd *exec.Cmd) {
	t := time.NewTimer(3 * time.Second)
	<-t.C
	pid := cmd.Process.Pid
	cmd.Process.Kill()
	fmt.Printf("sub process %d killed\n", pid)

	// in case parent blocking in handleReader, sub process will be zombie
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("failed to find process %d: %s\n", pid, err.Error())
	} else {
		fmt.Printf("find process %+v\n", process)
	}

	// quitCh is used to safe time for get sub process status before main thread quit.
	close(quitCh)
}

// wait Process to quit
func wait(cmd *exec.Cmd) {
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				fmt.Printf("wait exit status: %+v\n", err)
			} else {
				fmt.Printf("exit status: %+v\n", status)
			}
		} else {
			fmt.Printf("wait exit status: %+v\n", err)
		}
	}
}

// demo sub process print to stdout and stderr
func externalCmd() {
	i := 0
	for {
		i++
		time.Sleep(time.Second * 1)
		fmt.Fprintf(os.Stdout, "stdout %d\n", i)
		fmt.Fprintf(os.Stderr, "stderr %d\n", i)
	}
}
