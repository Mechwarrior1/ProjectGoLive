// this file manages the sessions and user information.
package session

import (
	"client/jwtsession"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
)

type (

	// // persistInfo is mainly to let the server give user feedback, based on their input in different pages.
	// // keeps tracks of the activity of the session.
	// userPersistInfo struct {

	// 	// gives a green highlighted msg on html if "true".
	// 	Success string

	// 	//the msg to be printed if Success is "true".
	// 	Msg string

	// 	// the username for the session, None if not logged in.
	// 	Username string

	// 	// a string of the date and time of user's last login, to be printed on html.
	// 	LastLogin string

	// 	// a variable to keep tract of user's last active time.
	// 	LastActive int64

	// 	// boolean, true of this user is an admin.
	// 	IsAdmin bool
	// }

	// a struct to manage jwt blacklist
	Session struct {
		MapSession *map[string]SessionStruct // maps UUID to date (int64)
		ApiKey     string
	}

	SessionStruct struct {
		Uuid       string
		LastActive int64
	}

	logger struct {
		c1 chan string
		c2 chan string
	}
)

// function attached to logger struct, uses the saved channel variables and pass values to the channels
func (logger logger) LogTrace(logType string, msg string) {
	logger.c1 <- logType
	logger.c2 <- msg
}

// the actual function incharge of logging
// opens 2 channels and returns both channels
// starts a goroutine that takes any string output from both channels and logs them into the logger file
// goroutine opens the logger file and defer the closure, waits for any input in a for loop
// for loop closes when it receives "close_goRoutine" on c1 channel
func LoggerGo() (chan string, chan string) {
	c1 := make(chan string)
	c2 := make(chan string)
	var logType string
	var msg string
	go func(c1 chan string, c2 chan string) {
		f, err := os.OpenFile("secure//logger.log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		for {
			logType = <-c1
			msg = <-c2
			Trace := log.New(f,
				logType+": ",
				log.Ldate|log.Ltime|log.Lshortfile)
			if logType == "close_goRoutine" {
				Trace := log.New(f,
					"TRACE: ",
					log.Ldate|log.Ltime|log.Lshortfile)
				Trace.Println("Closing logger go routine")
				break
			}
			Trace.Println(msg)
		}
	}(c1, c2)

	return c1, c2
}

// // a function to update a user's last login.
// func (u *userPersistInfo) updateLastLogin() *userPersistInfo {
// 	currentTime := time.Now()
// 	lastLogin := currentTime.Format("02-01-2006 15:04 Monday")
// 	u.LastLogin = lastLogin
// 	return u
// }

// logs the session in the sessionmanager.
func (s *Session) CheckSession(c echo.Context, jwtClaim *jwtsession.JwtClaim, jwtWrapper *jwtsession.JwtWrapper) *jwtsession.JwtClaim {
	if jwtClaim.Context.Username == "" {
		return jwtClaim
	}

	sessionStruct := (*s.MapSession)[jwtClaim.Context.Username] // check if previous session is around

	if sessionStruct.Uuid != jwtClaim.Context.Uuid { //if the jwt uuid id different from stored uuid, immediately invalidates current jwt

		fmt.Println(*s.MapSession)
		fmt.Println(jwtClaim.Context.Username, ",old uuid: ", sessionStruct.Uuid, ", your uuid: ", jwtClaim.Context.Uuid)

		newJwt, claims, _ := jwtWrapper.GenerateToken("error", "you have been logged out", "false", "", "", uuid.NewV4().String())

		fmt.Println("new mapsession2, ", claims.Context.Username, (*s.MapSession))
		NewCookie(c, 3, newJwt)

		return claims // return new claims for user, since old session got terminated
	}

	(*s.MapSession)[jwtClaim.Context.Username] = SessionStruct{jwtClaim.Context.Uuid, jwtClaim.StandardClaims.ExpiresAt}
	fmt.Println("new mapsession, ", (*s.MapSession))
	return jwtClaim
}

// function deletes the session, based on the session id string.
func (s *Session) DeleteSession(username string) {
	// fmt.Println("deleting session for ", username)
	delete(*s.MapSession, username)
}

// function runs as a go routine, runs an infinite for loop that check sessions once every 5 mins.
// it loop through all the sessions and checks its LastActive int variable.
// if it was last active was 30mins ago, it deletes the session.
func (s *Session) PruneOldSessions() { // intended to run concurrently, go pruneOldSessions().
	for {
		time.Sleep(300 * time.Second) //checks and prune every 5 mins.
		timeNow := time.Now().Unix()
		for username, struct1 := range *(s).MapSession { //k = uuid.
			if (timeNow - struct1.LastActive) > (10 * 60) {
				// if sessions' last active was 30mins (30 * 60 seconds) ago, delete session.
				// logger1.logTrace("TRACE", sessions.Username+" session expired")
				s.DeleteSession(username)
			}
		}
	}
}

// a function that gets Jwt from cookie.
func (s *Session) GetCookieJwt(c echo.Context, jwtWrapper *jwtsession.JwtWrapper) (*jwtsession.JwtClaim, error) { //set new cookie if cookie is not found.
	goRecycleCookie, err := c.Cookie("goRecycleCookie")

	if err != nil {
		//success string, msg string, admin string, lastLogin string, username string, uuid string
		newJwt, claims, err := jwtWrapper.GenerateToken("", "", "false", "", "", uuid.NewV4().String())
		if err != nil {
			return nil, err
		}

		NewCookie(c, 10*60, newJwt)
		return claims, errors.New("id not found, generated new")

	} else {
		claims, err := jwtWrapper.ValidateToken(goRecycleCookie.Value)
		if err != nil {
			return nil, err
		}
		claims = s.CheckSession(c, claims, jwtWrapper)

		return claims, nil
	}

}

// a function to generate a new cookie, with the session id as cookie value.
func NewCookie(c echo.Context, expiry int, id string) { //make a new cookie.
	goRecycleCookie := &http.Cookie{
		Name:   "goRecycleCookie",
		Value:  id,
		MaxAge: expiry,
		Path:   "/",
	}
	c.SetCookie(goRecycleCookie)
}

func ExpCookie(c echo.Context) { //make a new cookie.
	NewCookie(c, -1, "")
}

func UpdateJwtLong(success string, msg string, admin string, lastLogin string, username string, jwtContext *jwtsession.JwtContext, c echo.Context, jwtWrapper *jwtsession.JwtWrapper, session *Session) {
	//success string, msg string, admin string, lastLogin string, username string, uuid string
	newJwt, jwtClaim, err := jwtWrapper.GenerateToken(success, msg, admin, lastLogin, username, jwtContext.Uuid)

	if err != nil {
		return
	}
	//update new uuid to session
	(*session.MapSession)[jwtClaim.Context.Username] = SessionStruct{jwtClaim.Context.Uuid, jwtClaim.StandardClaims.ExpiresAt}

	NewCookie(c, 10*60, newJwt)
}

func UpdateJwt(success string, msg string, jwtContext *jwtsession.JwtContext, c echo.Context, jwtWrapper *jwtsession.JwtWrapper) {
	//success string, msg string, admin string, lastLogin string, username string, uuid string
	newJwt, _, _ := jwtWrapper.GenerateToken(success, msg, jwtContext.Admin, jwtContext.LastLogin, jwtContext.Username, jwtContext.Uuid)
	NewCookie(c, 10*60, newJwt)
}
