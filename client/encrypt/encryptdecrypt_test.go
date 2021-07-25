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
	encryptToFile("testkey.txt", testString)
	result := DecryptFromFile("testkey.txt")
	c.Assert(string(result), Equals, testString)

}
