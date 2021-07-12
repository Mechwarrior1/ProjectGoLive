package main

import (
	"apiserver/mysql"
	"apiserver/word2vec"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // go mod init api_server.go
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var (
	dbHandler1 mysql.DBHandler
	s          http.Server
	embed      *word2vec.Embeddings
	key1       = mysql.AnonFunc()
)

// function for the rest api, respond with the slice of all courses
func allInfo(w http.ResponseWriter, r *http.Request) {
	// can only return listing results, commentUser and commentItem
	dataPacket1, err1 := readJSONBody(w, r) // read response JSON

	if err1 != nil {
		fmt.Println(err1)
		return
	}

	receiveInfoRaw := dataPacket1.DataInfo[0].(map[string]interface{}) // convert received data into map[string]string
	receiveInfo := make(map[string]string)

	for k, v := range receiveInfoRaw {
		receiveInfo[k] = fmt.Sprintf("%v", v)
	}

	tarDB := dataPacket1.InfoType
	tarID := receiveInfo["ID"]

	switch tarDB { // gets all post

	case "ItemListing":
		sendInfo, _ := dbHandler1.GetRecordlisting(tarDB, receiveInfo["Name"], receiveInfo["filterUsername"], embed)
		w.WriteHeader(http.StatusCreated)
		newResponse(w, r, sendInfo, "nil", "ItemListing", "true", "")
		return

	case "CommentUser": // gets all comments regarding a particular user id
		sendInfo, _ := dbHandler1.GetRecord(tarDB)
		newSendInfo := []interface{}{}

		for i := range sendInfo {

			temp1 := sendInfo[i].(mysql.CommentUser)
			if temp1.ForUsername == tarID {
				newSendInfo = append(newSendInfo, sendInfo[i])
			}
		}

		w.WriteHeader(http.StatusCreated)
		newResponse(w, r, newSendInfo, "nil", "ItemListing", "true", "")
		return

	case "CommentItem": // gets all comments regarding a particular item id
		sendInfo, _ := dbHandler1.GetRecord(tarDB)
		newSendInfo := []interface{}{}

		for i := range sendInfo {
			temp1 := sendInfo[i].(mysql.CommentItem)

			if temp1.ForItem == tarID {
				// fmt.Println(sendInfo[i])
				newSendInfo = append(newSendInfo, sendInfo[i])
			}
		}

		w.WriteHeader(http.StatusCreated)
		newResponse(w, r, newSendInfo, "nil", "ItemListing", "true", "")
		return

	default:
		w.WriteHeader(http.StatusNotFound)
		newErrorResponse(w, r, "404 - not found")
		return
	}
}

func pwCheck(w http.ResponseWriter, r *http.Request) { //works
	dataPacket1, err := readJSONBody(w, r) // read response JSON

	if err != nil { //err means username not found, ok to proceed
		return
	}

	receiveInfoRaw := dataPacket1.DataInfo[0].(map[string]interface{}) // convert received data into map[string]string
	receiveInfo := make(map[string]string)

	for k, v := range receiveInfoRaw {
		receiveInfo[k] = fmt.Sprintf("%v", v)
	}

	dbData, err1 := dbHandler1.GetSingleRecord("UserSecret", "WHERE Username = \"?\"", receiveInfo["Username"])

	if err1 != nil || len(dbData) == 0 { //err means username not found, ok to proceed
		w.WriteHeader(http.StatusNotFound)
		newErrorResponse(w, r, "404 - Invalid credentials")
		return
	}

	// update last login
	dbData2, err2 := dbHandler1.GetSingleRecord("UserInfo", "WHERE Username = \"?\"", receiveInfo["Username"])
	dbDataInfo := dbData2[0].(map[string]interface{})
	dbData1 := dbData[0].(mysql.UserSecret)
	err3 := bcrypt.CompareHashAndPassword([]byte(dbData1.Password), []byte(receiveInfo["Password"]))

	if err2 == nil && err3 == nil {
		w.WriteHeader(http.StatusCreated)
		//update lastLogin if there is no issues
		_ = dbHandler1.EditRecord("UserInfo", dbDataInfo)
		//return response with true if no issues
		newResponse(w, r, []interface{}{}, "nil", "UserInfo", "true", "")
		return
	}

	w.WriteHeader(http.StatusNotFound)
	newErrorResponse(w, r, "402 - Invalid credentials")
}

// change map[string]interface to map[string]string
func mapInterfaceToString(dataPacket1 *mysql.DataPacket) map[string]string {
	receiveInfoRaw := dataPacket1.DataInfo[0].(map[string]interface{}) // convert received data into map[string]string
	receiveInfo := make(map[string]string)

	for k, v := range receiveInfoRaw {
		receiveInfo[k] = fmt.Sprintf("%v", v)
	}

	return receiveInfo
}

// checks if the username in the sent info is currently in DB
// returns false if username is not taken
func usernameCheck(w http.ResponseWriter, r *http.Request) {
	dataPacket1, err1 := readJSONBody(w, r) // read response JSON
	if err1 != nil {                        //err means username not found, ok to proceed
		return
	}

	receiveInfo := mapInterfaceToString(dataPacket1) // convert received data into map[string]string
	allData, err := dbHandler1.GetSingleRecord("UserInfo", " WHERE Username = \"?\"", receiveInfo["Username"])

	if err != nil || len(allData) == 0 { //err means username not found, ok to proceed
		w.WriteHeader(http.StatusNotFound)
		newResponse(w, r, []interface{}{}, "username not found", "UserInfo", "false", "")
		return
	}

	w.WriteHeader(http.StatusCreated)
	newResponse(w, r, []interface{}{}, "nil", "UserInfo", "true", "")
}

// the function that writes the response back
func newResponse(w http.ResponseWriter, r *http.Request, dataInfo []interface{}, errorMsg string, infoType string, resBool string, requestUser string) {
	if errorMsg == "nil" {
		w.WriteHeader(http.StatusCreated)
	}
	//StatusOK 200
	//StatusAccepted 202
	//StatusCreated 201
	//StatusFound 302
	//StatusBadRequest 400
	//StatusForbidden 403
	//StatusInternalServerError 500

	var dataPacket1 mysql.DataPacket
	dataPacket1.DataInfo = dataInfo
	dataPacket1.ErrorMsg = errorMsg       // error msg if any
	dataPacket1.InfoType = infoType       // to access which db
	dataPacket1.ResBool = resBool         //
	dataPacket1.RequestUser = requestUser //request coming from which user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataPacket1)

}

// Response writer for when returning an error to a query
func newErrorResponse(w http.ResponseWriter, r *http.Request, errorMsg string) {
	var DataInfo1 []interface{}
	newResponse(w, r, DataInfo1, errorMsg, "error", "false", "")
}

// function to read the JSON on a request
func readJSONBody(w http.ResponseWriter, r *http.Request) (*mysql.DataPacket, error) {
	var newDataPacket mysql.DataPacket
	reqBody, err := ioutil.ReadAll(r.Body)

	if err == nil {
		// convert JSON to object
		json.Unmarshal(reqBody, &newDataPacket)
		if newDataPacket.Key != key1() {
			w.WriteHeader(http.StatusNotFound)
			newErrorResponse(w, r, "401 - Invalid key")
			return &newDataPacket, errors.New("incorrect api key supplied")
		}

		if len(newDataPacket.DataInfo) == 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			newErrorResponse(w, r, "422 - Please supply course information in JSON format")
			return &newDataPacket, errors.New("unable to read marshal json")
		}

		return &newDataPacket, nil
	}

	return &newDataPacket, errors.New("error while attempting to read body of request")
}

// checks if owner of the post is the same as the one requesting, for edits or
func checkUser(tarDB string, requestUser string, dataInfo []interface{}) bool {

	switch tarDB {
	case "UserInfo":
		dataInfo1 := dataInfo[0].(mysql.UserInfo)
		return dataInfo1.Username == requestUser

	case "ItemListing":
		dataInfo1 := dataInfo[0].(mysql.ItemListing)
		return dataInfo1.Username == requestUser

	case "CommentUser":
		dataInfo1 := dataInfo[0].(mysql.CommentUser)
		return dataInfo1.Username == requestUser

	case "CommentItem":
		dataInfo1 := dataInfo[0].(mysql.CommentItem)
		return dataInfo1.Username == requestUser

	case "UserSecret":
		dataInfo1 := dataInfo[0].(mysql.UserSecret)
		return dataInfo1.Username == requestUser

	default:
		fmt.Println("logger: " + tarDB + " is not a valid db")
		return false
	}

}

// main function for the rest api.
// GET request returns the specified post/user/comments.
// POST request will post the post/user/comments into the DB.
// PUT request will edit the specified post/user/comments.
func genInfo(w http.ResponseWriter, r *http.Request) {
	dataPacket1, err1 := readJSONBody(w, r) // read response JSON
	if err1 != nil {
		fmt.Println(err1)
		return
	}

	receiveInfoRaw := dataPacket1.DataInfo[0].(map[string]interface{}) // convert received data into map[string]string
	// receiveInfo := make(map[string]string)

	// for k, v := range receiveInfoRaw {
	// 	receiveInfo[k] = fmt.Sprintf("%v", v)
	// }

	tarItemID := receiveInfoRaw["ID"]
	tarDB := dataPacket1.InfoType
	if r.Method == "POST" {
		err2 := dbHandler1.InsertRecord(tarDB, receiveInfoRaw, "") // deletes if target is found
		if err2 == nil {
			w.WriteHeader(http.StatusCreated)
			newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
			return

		}

		fmt.Println("logger: insert, " + tarDB + ": " + tarDB + " db not found") //reach here only if it is not returned by the switch
		w.WriteHeader(http.StatusConflict)
		newErrorResponse(w, r, "Bad Request , item not found")
		return

	}

	// if request is not post, check if the ID exist before proceeding
	dbInfoSlice, err3 := dbHandler1.GetSingleRecord(tarDB, " WHERE ID = ?", tarItemID.(string))

	if err3 != nil || len(dbInfoSlice) == 0 {

		if tarDB == "UserInfo" {
			dbInfoSlice, err3 = dbHandler1.GetSingleRecord(tarDB, " WHERE Username = \"?\"", tarItemID.(string))
		}

		if err3 != nil || len(dbInfoSlice) == 0 {
			fmt.Println(dbInfoSlice, err3)
			fmt.Println("logger: error when looking up id " + tarItemID.(string) + " for DB " + tarDB + ", err:" + err3.Error())
			w.WriteHeader(http.StatusNotFound)
			newErrorResponse(w, r, "Bad Request , item not found")
			return
		}

	}

	// for userinfo, itemlisting, commentitem and commentuser only
	if r.Method == "GET" {

		if tarDB != "UserSecret" { //prevents any requeset to ask for user secrets
			w.WriteHeader(http.StatusCreated)
			newResponse(w, r, dbInfoSlice, "nil", tarDB, "true", "")
			return
		}

		w.WriteHeader(http.StatusNotFound)
		newErrorResponse(w, r, "Bad Request , item not found")
		return
	}

	if !checkUser(tarDB, dataPacket1.RequestUser, dbInfoSlice) { //checks if the user that requested deletion is the owner who posted the comment/listing
		fmt.Println("logger: " + dataPacket1.RequestUser + " tried to access an item not from user, item ID:" + tarItemID.(string) + ", DB: " + tarDB)
		w.WriteHeader(http.StatusNotFound)
		newErrorResponse(w, r, "Bad Request , item not found")
		return
	}
	// for deleting an entry
	if r.Method == "DELETE" {

		if tarDB == "ItemListing" || tarDB == "CommentUser" || tarDB == "CommentItem" { //only delete records for 3 items
			err2 := dbHandler1.DeleteRecord(tarDB, tarItemID.(string)) // attempt to delete record
			if err2 != nil {
				w.WriteHeader(http.StatusCreated)
				newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
				return
			}

		} else {
			//if delete did not occur
			fmt.Println("logger:  " + ": " + tarDB + " db not found or not in use for Delete func, err:")
			w.WriteHeader(http.StatusConflict)
			newErrorResponse(w, r, "Bad Request , item not found")
			return
		}
	}

	// PUT is for updating existing course
	if r.Method == "PUT" {
		var err2 error
		err2 = dbHandler1.EditRecord(tarDB, receiveInfoRaw) // deletes if target is found

		if err2 == nil {
			w.WriteHeader(http.StatusCreated)
			newResponse(w, r, []interface{}{}, "nil", "CommentUser", "true", "")
			return
		}

		fmt.Println("logger: edit, " + tarDB + ": " + tarDB + " db not found") //reach here only if it is not returned by the switch
		w.WriteHeader(http.StatusConflict)
		newErrorResponse(w, r, "Bad Request , item not found")
		return

	}
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	embed = word2vec.GetWord2Vec()
	dbHandler1 = mysql.OpenDB()
	defer dbHandler1.DB.Close()

	router := mux.NewRouter()
	router.HandleFunc("/api/v0/check", pwCheck)
	router.HandleFunc("/api/v0/allinfo", allInfo)
	router.HandleFunc("/api/v0/username", usernameCheck)
	router.HandleFunc("/api/v0/db/info", genInfo).Methods("GET", "PUT", "POST", "DELETE")
	fmt.Println("listening at port 5555")
	s = http.Server{Addr: ":5555", Handler: router}

	if err := s.ListenAndServeTLS("secure//cert.pem", "secure//key.pem"); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

}
