/*
  Use `make` to run the test. `go test` won't work because it should run inside
  a privileged container with host `/proc` bind-mount to `/host/proc`
*/

package nsfilelock

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

const (
	HostNamespace    = "/host/proc/1/ns"
	LockFile         = "/tmp/lock"
	TestAssistBinary = "./bin/test_assist"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) basicTest(c *C, ns string) {
	var (
		l1, l2 *NSFileLock
		err    error
	)

	l1 = NewLock(ns, LockFile)
	err = l1.Lock()
	c.Assert(err, IsNil)

	l2 = NewLock(ns, LockFile)
	err = l2.Lock()
	c.Assert(err, ErrorMatches, "Timeout waiting for lock")

	l1.Unlock()

	l2 = NewLock(ns, LockFile)
	err = l2.Lock()
	c.Assert(err, IsNil)
	l2.Unlock()
}

func (s *TestSuite) TestBasicLock(c *C) {
	s.basicTest(c, "")
}

func (s *TestSuite) TestBasicLockInHostNamespace(c *C) {
	s.basicTest(c, HostNamespace)
}

func (s *TestSuite) TestBasicLockInInvalidNamespace(c *C) {
	lock := NewLock("/invalidns/", LockFile)
	err := lock.Lock()
	c.Assert(err, ErrorMatches, "Invalid namespace fd.*")
}

func (s *TestSuite) TestWaitForLock(c *C) {
	var (
		l1, l2 *NSFileLock
		err    error
	)

	l1 = NewLock(HostNamespace, LockFile)
	err = l1.Lock()
	c.Assert(err, IsNil)
	l1.Unlock()

	go func() {
		var err error
		l1 = NewLock(HostNamespace, LockFile)
		err = l1.Lock()
		c.Assert(err, IsNil)
		time.Sleep(3 * time.Second)
		l1.Unlock()
	}()

	//Let goroutine run first
	time.Sleep(1 * time.Second)

	l2 = NewLock(HostNamespace, LockFile)
	err = l2.Lock()
	c.Assert(err, IsNil)
}

func (s *TestSuite) lockProcess(c *C, ns string) {
	var cmd *exec.Cmd

	if ns == "" {
		cmd = exec.Command(TestAssistBinary, LockFile)
	} else {
		cmd = exec.Command(TestAssistBinary, LockFile, ns)
	}
	output, err := cmd.Output()
	c.Assert(strings.TrimSpace(string(output)), Equals, SuccessResponse)
	c.Assert(err, IsNil)
}

func (s *TestSuite) lockProcessTest(c *C, ns string) {
	// Execute twice, and make sure no lock process leftover so the second
	// time won't timeout
	s.lockProcess(c, ns)
	s.lockProcess(c, ns)

	l := NewLock(ns, LockFile)
	err := l.Lock()
	c.Assert(err, IsNil)

	l.Unlock()
}

func (s *TestSuite) TestAutomaticUnlock(c *C) {
	s.lockProcessTest(c, "")
}

func (s *TestSuite) TestAutomaticUnlockInHostNamespace(c *C) {
	s.lockProcessTest(c, HostNamespace)
}

func (s *TestSuite) TestLockWithTimeout(c *C) {
	var (
		l1, l2 *NSFileLock
		err    error
	)

	ns := HostNamespace
	ns = ""
	go func() {
		var err error
		l1 = NewLockWithTimeout(ns, LockFile, 0)
		err = l1.Lock()
		c.Assert(err, IsNil)
		time.Sleep(5 * time.Second)
		l1.Unlock()
	}()

	//Let goroutine run first
	time.Sleep(1 * time.Second)

	l2 = NewLockWithTimeout(ns, LockFile, 3*time.Second)
	err = l2.Lock()
	c.Assert(err, ErrorMatches, "Timeout waiting for lock")

	// 3 seconds passed, should unlock in 1 seconds
	l2 = NewLockWithTimeout(ns, LockFile, 5*time.Second)
	err = l2.Lock()
	c.Assert(err, IsNil)
	l2.Unlock()
}
