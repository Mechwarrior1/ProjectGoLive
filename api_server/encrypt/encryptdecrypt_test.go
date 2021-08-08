package encrypt

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestEncrypt(c *C) {
	testString := "test"
	EncryptToFile("testkey.txt", testString, "testkey.xml")
	result := DecryptFromFile("testkey.txt", "testkey.xml")
	c.Assert(string(result), Equals, testString)

}
