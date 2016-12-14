package nsfilelock

import (
	"testing"
	//"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) TestBasicLock(c *C) {
	var (
		l1, l2 *NSFileLock
		err    error
	)

	l1 = NewLock("", "/tmp/lock")
	err = l1.Lock()
	c.Assert(err, IsNil)
	l1.Unlock()

	l1 = NewLock("/host/proc/1/ns", "/tmp/lock")
	err = l1.Lock()
	c.Assert(err, IsNil)

	l2 = NewLock("/host/proc/1/ns", "/tmp/lock")
	err = l2.Lock()
	c.Assert(err, ErrorMatches, "Timeout waiting for lock")

	l1.Unlock()

	l2 = NewLock("/host/proc/1/ns", "/tmp/lock")
	err = l2.Lock()
	c.Assert(err, IsNil)
	l2.Unlock()

	//TODO: Test unlock during waiting
	//TODO: Test lock released after process exit
}
