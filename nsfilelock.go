package nsfilelock

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	MountNamespaceFD   = "mnt"
	MaximumMessageSize = 255
)

var (
	LockCheckInterval = 100 * time.Millisecond
	Timeout           = 15 * time.Second
)

type NSFileLock struct {
	Namespace string
	FilePath  string

	done chan struct{}
}

func NewLock(ns string, filepath string) *NSFileLock {
	if ns == "" {
		ns = "/proc/1/ns/"
	}
	return &NSFileLock{
		Namespace: ns,
		FilePath:  filepath,
	}
}

func (l *NSFileLock) Lock() error {
	successResp := "locked"
	resp := ""

	l.done = make(chan struct{})
	result := make(chan string)
	timeout := make(chan struct{})

	nsFD := filepath.Join(l.Namespace, MountNamespaceFD)
	if _, err := os.Stat(nsFD); err != nil {
		return fmt.Errorf("Invalid namespace fd: %s", nsFD)
	}

	lockCmd := fmt.Sprintf("\"\"exec 314>%s; flock 314; echo %s; exec sleep 65535\"\"",
		l.FilePath, successResp)
	cmd := exec.Command("nsenter", "--mount="+nsFD, "bash", "-c", lockCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		var err error
		buf := make([]byte, MaximumMessageSize)
		n := 0
		for n == 0 {
			n, err = stdout.Read(buf)
			if err != nil {
				result <- err.Error()
			}
		}
		result <- strings.Trim(string(buf), "\n\x00")
	}()

	go func() {
		var err error
		buf := make([]byte, MaximumMessageSize)
		n := 0
		for n == 0 {
			n, err = stderr.Read(buf)
			if err != nil {
				result <- err.Error()
			}
		}
		result <- strings.Trim(string(buf), "\n\x00")
	}()

	go func() {
		time.Sleep(Timeout)
		timeout <- struct{}{}
	}()

	select {
	case resp = <-result:
		if resp != successResp {
			return fmt.Errorf("Failed to lock, response: %s", resp)
		}
	case <-timeout:
		syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
		return fmt.Errorf("Timeout waiting for lock")
	}

	// Wait for unlock
	go func() {
		select {
		case <-l.done:
			syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
			return
		}
	}()
	return nil
}

func (l *NSFileLock) Unlock() {
	l.done <- struct{}{}
}
