package main

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestCos(c *C) {
	testString := "teststring"
	encryptToFile("secure/testkey.txt", testString)
	result := decryptFromFile("secure/testkey.txt")
	c.Assert(string(result), Equals, testString)

}
