package common

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"runtime/debug"
	"strings"
	"syscall"
)

func Execute(commandLine string) (stdout *string, stderr *string, exit int) {
	stdout = nil
	stderr = nil
	exit = 0

	commands := strings.Split(commandLine, " ")
	cmd := exec.Command(commands[0], commands[1:]...)
	var bufOut bytes.Buffer
	var bufErr bytes.Buffer
	cmd.Stdout = &bufOut
	cmd.Stderr = &bufErr

	err := cmd.Run()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exit = status.ExitStatus()
			}
		} else {
			debug.PrintStack()
			log.Fatalf("cmd.Wait: %s, %v", commandLine, err)
		}
	}

	sout := bufOut.String()
	stdout = &sout

	serr := bufErr.String()
	stderr = &serr
	fmt.Printf("Command execute: %s, exitcode: %d\n", commandLine, exit)
	return
}

func IsBinaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
