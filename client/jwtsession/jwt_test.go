package jwtsession

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	jwtWrapper := &JwtWrapper{
		"key",
		"GoRecycle",
		10,
	}

	generatedToken, _, err := jwtWrapper.GenerateToken("success", "msg", "false", "lastlogin", "username", "uuid")
	assert.NoError(t, err)

	os.Setenv("testToken", generatedToken)
}

func TestValidateToken(t *testing.T) {
	encodedToken := os.Getenv("testToken")

	jwtWrapper := &JwtWrapper{
		"key",
		"GoRecycle",
		10,
	}

	claims, err := jwtWrapper.ValidateToken(encodedToken)
	assert.NoError(t, err)

	assert.Equal(t, "success", claims.Context.Success)
	assert.Equal(t, "msg", claims.Context.Msg)
	assert.Equal(t, "lastlogin", claims.Context.LastLogin)
	assert.Equal(t, "username", claims.Context.Username)
	assert.Equal(t, "uuid", claims.Context.Uuid)
	assert.Equal(t, "GoRecycle", claims.StandardClaims.Issuer)
}
