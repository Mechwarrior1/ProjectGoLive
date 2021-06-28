// this file contains the http server functions for login, logout and sign out.
package main

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// a function for http handler, used for /logout, handles the logout portion.
func logout(res http.ResponseWriter, req *http.Request) {
	oldId, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
	if !alreadyLoggedIn(userPersistInfo1) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	// remove the cookie
	sessionMgr.deleteSession(oldId)
	sessionMgr.newCookieAndSet("logout********", res, req, "")
	http.Redirect(res, req, "/", http.StatusSeeOther)
	logger1.logTrace("TRACE", userPersistInfo1.Username+" has logged out")
}

// a function to check if current session has a username, returns true if session has a username.
func alreadyLoggedIn(userPersistInfo1 *userPersistInfo) bool {
	if userPersistInfo1.Username == "logout********" || userPersistInfo1.Username == "None" || userPersistInfo1.Username == "" {
		return false
	}
	_, ok := (*sessionMgr.mapSession)[userPersistInfo1.Username]
	return ok
}

// a function for http handler, used for /signup, handles the signup portion.
func signup(res http.ResponseWriter, req *http.Request) {
	id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
	if alreadyLoggedIn(userPersistInfo1) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	// process form submission.
	if req.Method == http.MethodPost {
		// get form values.
		username := req.FormValue("username")
		username1 := replaceAllString(username)
		if username != username1 {
			logger1.logTrace("TRACE", "someone tried to use "+username+" to sign up, but it contained special characters")
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "Please only use alphanumeric characters", "logout********", "None", false)
			http.Redirect(res, req, "/signup", http.StatusSeeOther)
			return
		}
		password := req.FormValue("password")
		if username != "" && username != "logout********" {
			// check if username exist/ taken.
			if ok := checkUsername(res, req, username); ok {
				logger1.logTrace("TRACE", "someone tried to use "+username+" to sign up, but it was used")
				sessionMgr.updatePersistInfo(id, "false", "None", "true", "Username already taken", "logout********", "None", false)
				http.Redirect(res, req, "/signup", http.StatusSeeOther)
				return
			}
			bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			if err != nil {
				sessionMgr.updatePersistInfo(id, "false", "None", "true", "There was an error, please try again", "logout********", "None", false)
				http.Redirect(res, req, "/signup", http.StatusSeeOther)
				return
			}
			currentTime := time.Now()
			lastLogin := currentTime.Format("2006-01-02 15:04:05 Monday")
			err5 := addUser(username, string(bPassword), "", lastLogin) // send user info to api
			if err5 != nil {
				sessionMgr.updatePersistInfo(id, "false", "None", "true", "There was an error, please try again", "logout********", "None", false)
				http.Redirect(res, req, "/signup", http.StatusSeeOther)
				return
			}
			sessionMgr.newCookieAndSet("", res, req, id) // create session.
			id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
			userPersistInfo1.LastLogin = lastLogin
			sessionMgr.logSession(id, userPersistInfo1)
		}
		sessionMgr.updatePersistInfo(id, "true", "You have successfully signed up!", "false", "None", userPersistInfo1.Username, userPersistInfo1.LastLogin, false)
		// redirect to main index.
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	tplSignup.ExecuteTemplate(res, "signup.gohtml", userPersistInfo1)
}

// a function for http handler, used for /login, handles the login portion.
func login(res http.ResponseWriter, req *http.Request) {
	id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
	if alreadyLoggedIn(userPersistInfo1) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	// process form submission.
	if req.Method == http.MethodPost {
		username := req.FormValue("username")
		password := req.FormValue("password")
		username1 := replaceAllString(username)
		if username != username1 {
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "Please only use alphanumeric characters", "logout********", "None", false)
			http.Redirect(res, req, "/signup", http.StatusSeeOther)
			return
		}
		// check if user exist with username.
		// Matching of password entered.
		ok1 := checkPW(username, password, userPersistInfo1.Username) /////////////////////////
		if !ok1 {
			logger1.logTrace("TRACE", "someone tried to login to: "+username+", but the password is wrong")
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "Username and/or password do not match", "logout********", "None", false)
			http.Redirect(res, req, "/login", http.StatusSeeOther)
			return
		}
		sessionMgr.checkLoggedIn(username)
		// create session.
		sessionMgr.updatePersistInfo(id, "true", "You have successfully logged in!", "false", "None", username, userPersistInfo1.LastLogin, false)
		sessionMgr.newCookieAndSet("", res, req, id)
		_, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
		sessionMgr.logSession(id, userPersistInfo1)
		logger1.logTrace("TRACE", username+" logged in")
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	tplLogin.ExecuteTemplate(res, "login.gohtml", userPersistInfo1)
}
