package jwtsession

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt"
	uuid "github.com/satori/go.uuid"
)

// JwtWrapper wraps the signing key and the issuer
type (
	JwtWrapper struct {
		SecretKey         string
		Issuer            string
		ExpirationMinutes int64
	}

	// JwtClaim adds email as a claim to the token
	JwtClaim struct {
		Context JwtContext
		jwt.StandardClaims
	}

	//hold session information
	JwtContext struct {
		Success   string
		Msg       string
		Admin     string
		LastLogin string
		Username  string
		Uuid      string
	}
)

// GenerateToken generates a jwt token
func (j *JwtWrapper) GenerateToken(success string, msg string, admin string, lastLogin string, username string, uuid1 string) (signedToken string, claims *JwtClaim, err error) {
	if uuid1 == "" {
		uuid1 = uuid.NewV4().String()

	}
	jwtContext := JwtContext{
		success,
		msg,
		admin,
		lastLogin,
		username,
		uuid1}

	claims = &JwtClaim{
		Context: jwtContext,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Minute * time.Duration(j.ExpirationMinutes)).Unix(),
			Issuer:    j.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err = token.SignedString([]byte(j.SecretKey))
	if err != nil {
		return
	}

	return
}

//ValidateToken validates the jwt token
func (j *JwtWrapper) ValidateToken(signedToken string) (claims *JwtClaim, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(j.SecretKey), nil
		},
	)

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*JwtClaim)
	if !ok {
		err = errors.New("could not parse claims")
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("JWT is expired")
		return
	}

	return

}
