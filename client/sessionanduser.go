// this file manages the sessions and user information.
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	uuid "github.com/satori/go.uuid"
)

type (

	// persistInfo is mainly to let the server give user feedback, based on their input in different pages.
	// keeps tracks of the activity of the session.
	userPersistInfo struct {

		// gives a green highlighted msg on html if "true".
		Success string

		//the msg to be printed if Success is "true".
		SuccessMsg string

		// gives a red highlighted msg on html if "true".
		Error string

		// the msg to be printed if Error is "true".
		ErrorMsg string

		// the username for the session, None if not logged in.
		Username string

		// a string of the date and time of user's last login, to be printed on html.
		LastLogin string

		// a variable to keep tract of user's last active time.
		LastActive int64

		// boolean, true of this user is an admin.
		IsAdmin bool

		// // contains the id of the user's session, keeps the old id if he logs out.
		// lastSessionId string
	}
	userSecret struct {
		ID          string `json:"ID"`
		Username    string `json:"Username"`
		Password    string `json:"Password"`
		IsAdmin     string `json:"IsAdmin"`
		CommentItem string `json:"CommentItem"`
	}

	userInfo struct {
		ID          string `json:"ID"`
		Username    string `json:"Username"`
		LastLogin   string `json:"LastLogin"`
		DateJoin    string `json:"DateJoin"`
		CommentItem string `json:"CommentItem"`
	}

	itemListing struct {
		ID              string `json:"ID"`
		Username        string `json:"Username"`
		Name            string `json:"Name"`
		ImageLink       string `json:"ImageLink"`
		DatePosted      string `json:"DatePosted"`
		CommentItem     string `json:"CommentItem"`
		ConditionItem   string `json:"ConditionItem"`
		Cat             string `json:"Cat"`
		ContactMeetInfo string `json:"ContactMeetInfo"`
		Similarity      string `json:"Similarity"`
		Completion      string `json:"Completion"`
	}

	commentUser struct {
		ID          string `json:"ID"`
		Username    string `json:"Username"`
		ForUsername string `json:"ForUsername"`
		Date        string `json:"Date"`
		CommentItem string `json:"CommentItem"`
	}
	commentItem struct {
		ID          string `json:"ID"`
		Username    string `json:"Username"`
		ForItem     string `json:"ForItem"`
		Date        string `json:"Date"`
		CommentItem string `json:"CommentItem"`
	}

	dataPacket struct {
		// key to access rest api
		Key         string              `json:"Key"`
		ErrorMsg    string              `json:"ErrorMsg"`
		InfoType    string              `json:"InfoType"` // 5 types: userSecret, userInfo, itemListing, commentUser, commentItem
		ResBool     string              `json:"ResBool"`
		RequestUser string              `json:"RequestUser"`
		DataInfo    []map[string]string `json:"DataInfo"`
	}
	// a struct to handle all the server session and user information.
	sessionManager struct {
		mapPersistInfo *map[string]*userPersistInfo // takes in ID as string and maps to their session.
		mapSession     *map[string]*sessionStruct   // maps UUID to date (int64)
	}

	sessionStruct struct {
		ID       string
		TimeLast int64
	}

	logger struct {
		c1 chan string
		c2 chan string
	}
)

// function attached to logger struct, uses the saved channel variables and pass values to the channels
func (logger logger) logTrace(logType string, msg string) {
	logger.c1 <- logType
	logger.c2 <- msg
}

// the actual function incharge of logging
// opens 2 channels and returns both channels
// starts a goroutine that takes any string output from both channels and logs them into the logger file
// goroutine opens the logger file and defer the closure, waits for any input in a for loop
// for loop closes when it receives "close_goRoutine" on c1 channel
func loggerGo() (chan string, chan string) {
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

// a function to update a user's last login.
func (u *userPersistInfo) updateLastLogin() *userPersistInfo {
	currentTime := time.Now()
	lastLogin := currentTime.Format("2006-01-02 15:04:05 Monday")
	u.LastLogin = lastLogin
	return u
}

// sends user info to api
func addUser(username string, pwString string, commentItem string, lastLogin string) error {
	userSecret1 := make(map[string]string)
	userSecret1["Username"] = username
	userSecret1["Password"] = pwString
	userSecret1["CommentItem"] = commentItem
	jsonData1 := dataPacket{
		// key to access rest api
		Key:         "abce85da-b8b1-11eb-8529-0242ac130003",
		ErrorMsg:    "nil",
		InfoType:    "UserSecret",
		ResBool:     "false",
		RequestUser: username,
		DataInfo:    []map[string]string{userSecret1},
	}
	res, err1 := tapAPI(http.MethodPost, jsonData1, "https://127.0.0.1:5555/api/v0/db/info")
	if err1 != nil {
		return err1
	}
	fmt.Println(res)
	currentTime := time.Now()
	userInfo1 := make(map[string]string)
	userInfo1["Username"] = username
	userInfo1["LastLogin"] = lastLogin
	userInfo1["DateJoin"] = currentTime.Format("2006-01-02 Monday")
	userInfo1["CommentItem"] = commentItem

	jsonData1.InfoType = "UserInfo"
	jsonData1.DataInfo = []map[string]string{userInfo1}
	res2, err2 := tapAPI(http.MethodPost, jsonData1, "https://127.0.0.1:5555/api/v0/db/info")
	if err2 != nil {
		return err1
	}
	fmt.Println(res2)
	logger1.logTrace("TRACE", username+" is added/updated to system")
	if err1 == nil && err2 == nil {

	}
	return err1
}

// deletes the user struct that is mapped to the username from mapUsers.
// func (s *sessionManager) deleteUser(username string) {
// put in api request to backend server
// }

// logs the session in the sessionmanager.
// updates the user's lastlogin.
// attaches session id to user's lastsession.
// pushes above user updates to mapUsers.
// insert username to session id / persistInfo.
// insert username to list of logged in users (mapLoggedUsers).
func (s *sessionManager) logSession(id string, userPersistInfo1 *userPersistInfo) {
	userPersistInfo2 := userPersistInfo1.updateLastLogin() //update user with new id and login date.
	sessionStrt := &sessionStruct{id, userPersistInfo2.LastActive}
	(*s.mapSession)[userPersistInfo2.Username] = sessionStrt
	(*s.mapPersistInfo)[id] = userPersistInfo2
}

// checks if the user is logged in on another device.
// takes in a username string to check against the mapLoggedUsers.
// if the user is found to be logged in from another device.
// it deletes the link between the old sesion and user, logs out the old user, gives an error msg to the old user's session.
// the new logged in user will have control.
func (s *sessionManager) checkLoggedIn(username string) {
	_, ok := (*s.mapSession)[username]
	if ok {
		lastSessionStruct := (*s.mapSession)[username]
		s.deleteSession(lastSessionStruct.ID)
		userPersistInfo1 := userPersistInfo{
			Success:    "true",
			SuccessMsg: "you have been logged out due to being logged in from another device",
			Error:      "false",
			ErrorMsg:   "None",
			Username:   "None",
			LastLogin:  "None",
			LastActive: time.Now().Unix(),
			IsAdmin:    false,
		}
		(*s.mapPersistInfo)[lastSessionStruct.ID] = &userPersistInfo1
	}
}

// function deletes the session, based on the session id string.
func (s *sessionManager) deleteSession(id string) error {
	targetSession, ok1 := (*s.mapPersistInfo)[id]
	if ok1 {
		delete(*s.mapSession, targetSession.Username)
		delete(*s.mapPersistInfo, id)
	}
	return errors.New("Session not found for id: " + id)
}

// function runs as a go routine, runs an infinite for loop that check sessions once every 5 mins.
// it loop through all the sessions and checks its LastActive int variable.
// if it was last active was 30mins ago, it deletes the session.
func (s *sessionManager) pruneOldSessions() { // intended to run concurrently, go pruneOldSessions().
	for {
		time.Sleep(300 * time.Second) //checks and prune every 5 mins.
		timeNow := time.Now().Unix()
		for k, sessions := range *(s).mapPersistInfo { //k = uuid.
			if (timeNow - sessions.LastActive) > (60 * 30) {
				// if sessions' last active was 30mins (30 * 60 seconds) ago, delete session.
				logger1.logTrace("TRACE", sessions.Username+" session expired")
				s.deleteSession(k)
			}
		}
	}
}

// a function that returns the session id (uuid) from the browser cookie.
// generates a new session id (uuid) if no cookie is found.
func (s *sessionManager) getId(res http.ResponseWriter, req *http.Request) (string, error) { //set new cookie if cookie is not found.
	goRecycleCookie, err := req.Cookie("goRecycleCookie")
	if err != nil {
		newId := uuid.NewV4()
		return newId.String(), errors.New("id not found, generated new")
	}
	return goRecycleCookie.Value, nil
}

// a function that returns the id and session info (persistInfo) for the particular id.
// uses getId() to retrieve id first.
func (s *sessionManager) getIdPersistInfo(res http.ResponseWriter, req *http.Request) (string, *userPersistInfo) {
	id, err := s.getId(res, req)
	var userPersistInfo1 *userPersistInfo
	ok := false
	if err == nil {
		userPersistInfo1, ok = (*s.mapPersistInfo)[id]
	}
	if !ok {
		// uuid has expired, generate new
		userPersistInfo1 = &userPersistInfo{
			Success:    "false",
			SuccessMsg: "None",
			Error:      "false",
			ErrorMsg:   "None",
			Username:   "None",
			LastLogin:  "None",
			LastActive: time.Now().Unix(),
			IsAdmin:    false,
		}
		s.newCookieAndSet("None", res, req, id)
		(*s.mapPersistInfo)[id] = userPersistInfo1
	}
	fmt.Println(userPersistInfo1.Username)
	return id, userPersistInfo1
}

// a function to update the session information for the target id.
func (s *sessionManager) updatePersistInfo(id string, success string, successMsg string, errors string, errorsMsg string,
	username string, lastLogin string, IsAdmin bool) {
	userPersistInfo1 := &userPersistInfo{
		Success:    success,
		SuccessMsg: successMsg,
		Error:      errors,
		ErrorMsg:   errorsMsg,
		Username:   username,
		LastLogin:  lastLogin,
		LastActive: time.Now().Unix(),
		IsAdmin:    IsAdmin,
	}
	if lastLogin == "seelast" {
		if userPersistInfo2, ok := (*s.mapPersistInfo)[id]; ok {
			userPersistInfo1.Username = userPersistInfo2.Username
			userPersistInfo1.LastLogin = userPersistInfo2.LastLogin
			userPersistInfo1.IsAdmin = userPersistInfo2.IsAdmin
		}
	}
	(*s.mapPersistInfo)[id] = userPersistInfo1
}

// another function to edit the session info (persistInfo).
func (s *sessionManager) removePersistInfoError(id string) {
	userPersistInfos1, ok := (*s.mapPersistInfo)[id]
	if ok {
		if userPersistInfos1.Error == "true1" { //true1 errors are the ones we want to show.
			userPersistInfos1.Error = "true"
			(*s.mapPersistInfo)[id] = userPersistInfos1
		} else {
			userPersistInfos1.Error = "false"
			userPersistInfos1.ErrorMsg = "None"
			(*s.mapPersistInfo)[id] = userPersistInfos1
		}
	} else {
		fmt.Println("error, no cookie found")
	}
}

// a function to generate a new cookie, with the session id as cookie value.
func newCookie(expiry int, id string) *http.Cookie { //make a new cookie.
	goRecycleCookie := &http.Cookie{
		Name:   "goRecycleCookie",
		Value:  id,
		MaxAge: expiry,
	}
	return goRecycleCookie
}

// a function that sets the cookie generated from the newCookie() func.
func (s *sessionManager) newCookieAndSet(useCase string, res http.ResponseWriter, req *http.Request, id string) {
	if useCase == "logout********" {
		goRecycleCookie := newCookie(-1, id)
		http.SetCookie(res, goRecycleCookie)
	} else {
		goRecycleCookie := newCookie(1000000, id)
		http.SetCookie(res, goRecycleCookie)
	}
}

func mapInterfaceToString(dataPacket1 *dataPacket) map[string]string {
	receiveInfo := dataPacket1.DataInfo[0] //.(map[string]string) // convert received data into map[string]string
	// receiveInfo := make(map[string]string)
	// for k, v := range receiveInfoRaw {
	// 	receiveInfo[k] = fmt.Sprintf("%v", v)
	// }
	return receiveInfo
}

// a function that checks if username is taken
func checkUsername(res http.ResponseWriter, req *http.Request, username string) bool {
	baseURL := "https://127.0.0.1:5555/api/v0/username"
	userInfo1 := make(map[string]string)
	userInfo1["Username"] = username
	jsonData1 := dataPacket{
		// key to access rest api
		Key:         "abce85da-b8b1-11eb-8529-0242ac130003",
		ErrorMsg:    "nil",
		InfoType:    "UserInfo",
		ResBool:     "false",
		RequestUser: "",
		DataInfo:    []map[string]string{userInfo1},
	}
	dataInfo1, err1 := tapAPI(http.MethodGet, jsonData1, baseURL)
	// receiveInfo := mapInterfaceToString(dataInfo1)
	fmt.Println("checkUser: ", err1, res, dataInfo1)

	return dataInfo1.ResBool == "true"
}

func checkPW(username string, password string, reqUser string) bool {
	baseURL := "https://127.0.0.1:5555/api/v0/check"
	userSecret1 := make(map[string]string)
	userSecret1["Username"] = username
	userSecret1["Password"] = password

	jsonData1 := dataPacket{
		// key to access rest api
		Key:         "abce85da-b8b1-11eb-8529-0242ac130003",
		ErrorMsg:    "nil",
		InfoType:    "UserSecret",
		ResBool:     "false",
		RequestUser: reqUser,
		DataInfo:    []map[string]string{userSecret1},
	}
	dataInfo1, err1 := tapAPI(http.MethodGet, jsonData1, baseURL)
	// receiveInfo := mapInterfaceToString(dataInfo1)

	fmt.Println("checkUser: ", err1, dataInfo1)
	return dataInfo1.ResBool == "true"
}

// a function for http handler, used for /deletemovie, lets the admin delete sessions currently logged on the system.
// unable to delete own session.
// only accessible by admin (user.isAdmin == true).
// func (s *sessionManager) deleteSessionPage(res http.ResponseWriter, req *http.Request) {
// 	id, userPersistInfo1 := s.getIdPersistInfo(res, req)
// 	if !userPersistInfo1.IsAdmin {
// 		s.updatePersistInfo(id, "false", "None", "true1", "You do not have the rights to access this page", userPersistInfo1.Username, "seelast")
// 		http.Redirect(res, req, "/", http.StatusSeeOther)
// 		return
// 	}

// 	dataInsert := struct {
// 		MapPersistInfo map[string]userPersistInfo
// 		PersistInfo    *userPersistInfo
// 	}{
// 		*s.mapPersistInfo,
// 		userPersistInfo1,
// 	}
// 	if req.Method == http.MethodPost {
// 		if !userPersistInfo1.IsAdmin {
// 			s.updatePersistInfo(id, "false", "None", "true1", "You do not have the rights to access this page", userPersistInfo1.Username, "seelast")
// 			http.Redirect(res, req, "/", http.StatusSeeOther)
// 			return
// 		}
// 		req.ParseForm()                            // Required if you don't call r.FormValue().
// 		deleteUuid := req.Form["deleteSessionNum"] //req.FormValue("deleteMovieNum").
// 		_, ok := (*s.mapPersistInfo)[deleteUuid[0]]
// 		if id == deleteUuid[0] {
// 			s.updatePersistInfo(id, "false", "None", "true", "You cannot delete your own session: "+id, userPersistInfo1.Username, "seelast")
// 			http.Redirect(res, req, "/deletesession", http.StatusSeeOther)
// 			return
// 		}
// 		if !ok {
// 			s.updatePersistInfo(id, "false", "None", "true", "Please select a correct ID", userPersistInfo1.Username, "seelast")
// 			http.Redirect(res, req, "/deletesession", http.StatusSeeOther)
// 			return
// 		}
// 		err2 := s.deleteSession(deleteUuid[0])
// 		if err2 != nil {
// 			s.updatePersistInfo(id, "false", "None", "true", "Target session not found", userPersistInfo1.Username, "seelast")
// 			http.Redirect(res, req, "/deletesession", http.StatusSeeOther)
// 			return
// 		}
// 		s.updatePersistInfo(id, "true", "You have deleted "+deleteUuid[0], "false", "None", userPersistInfo1.Username, "seelast")
// 		http.Redirect(res, req, "/deletesession", http.StatusSeeOther)
// 	}
// 	tplDeleteSession.ExecuteTemplate(res, "deletesession.gohtml", dataInsert)
// }

// // a function for http handler, used for /deleteuser, lets the admin delete users that have signed up.
// // unable to delete any admin users.
// // only accessible by admin (user.isAdmin == true).
// func deleteUser(res http.ResponseWriter, req *http.Request) {
// 	user, id, persistInfo1 := getUser(res, req)
// 	if !user.IsAdmin {
// 		sessionMgr.updatePersistInfo(id, "false", "None", "true1", "You do not have the rights to access this page", user.Username, "seelast")
// 		http.Redirect(res, req, "/", http.StatusSeeOther)
// 		return
// 	}
// 	userList := map[string]bool{}
// 	for k, user1 := range *sessionMgr.mapUsers {
// 		userList[k] = user1.IsAdmin
// 	}
// 	dataInsert := struct {
// 		UserList    map[string]bool
// 		PersistInfo persistInfo
// 	}{
// 		userList,
// 		persistInfo1,
// 	}
// 	if req.Method == http.MethodPost {
// 		if !user.IsAdmin {
// 			sessionMgr.updatePersistInfo(id, "false", "None", "true1", "You do not have the rights to access this page", user.Username, "seelast")
// 			http.Redirect(res, req, "/", http.StatusSeeOther)
// 			return
// 		}
// 		req.ParseForm()
// 		deleteUser := req.Form["deleteUserNum"]
// 		if len(deleteUser) == 0 {
// 			logger1.logTrace("WARNING", "an error occured while attempting to delete a user, parseform returned an empty slice")
// 			sessionMgr.updatePersistInfo(id, "false", "None", "true", "An error seems to have occured", user.Username, "seelast")
// 			http.Redirect(res, req, "/deleteuser", http.StatusSeeOther)
// 			return
// 		}
// 		myUser, ok := sessionMgr.getUser(deleteUser[0])
// 		if !ok {
// 			sessionMgr.updatePersistInfo(id, "false", "None", "true", "an error occured while attempting to delete a user", user.Username, "seelast")
// 			http.Redirect(res, req, "/deleteuser", http.StatusSeeOther)
// 			return
// 		}
// 		if myUser.IsAdmin {
// 			sessionMgr.updatePersistInfo(id, "false", "None", "true", "You cannot delete an Admin", user.Username, "seelast")
// 			http.Redirect(res, req, "/deleteuser", http.StatusSeeOther)
// 			return
// 		}
// 		sessionMgr.deleteUser(myUser.Username) // delete the user from mapUsers.
// 		for id2, _ := range *sessionMgr.mapPersistInfo {
// 			if id2 == deleteUser[0] {
// 				sessionMgr.deleteSession(id2) //deletes the deleted user's session if he is logged in.
// 			}
// 		}
// 		sessionMgr.updatePersistInfo(id, "true", "You have deleted the user: "+deleteUser[0], "false", "None", user.Username, "seelast")
// 		http.Redirect(res, req, "/deleteuser", http.StatusSeeOther)
// 	}
// 	tplDeleteUser.ExecuteTemplate(res, "deleteuser.gohtml", dataInsert)
// }
