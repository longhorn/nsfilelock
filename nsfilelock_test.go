/*
  Use `make` to run the test. `go test` won't work because it should run inside
  a privileged container with host `/proc` bind-mount to `/host/proc`
*/

package nsfilelock

import (
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

const (
	HostNamespace = "/host/proc/1/ns"
	LockFile      = "/tmp/lock"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) BasicTest(c *C, ns string) {
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
	s.BasicTest(c, "")
}

func (s *TestSuite) TestBasicLockInHostNamespace(c *C) {
	s.BasicTest(c, HostNamespace)
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
		l1 = NewLock(HostNamespace, LockFile)
		err = l1.Lock()
		c.Assert(err, IsNil)
		time.Sleep(3 * time.Second)
		l1.Unlock()
	}()

	l2 = NewLock(HostNamespace, LockFile)
	err = l2.Lock()
	c.Assert(err, IsNil)
}
