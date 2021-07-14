package main

import (
	"apiserver/mysql"
	"apiserver/word2vec"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // go mod init api_server.go
	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
)

var (
	dbHandler1 mysql.DBHandler
	s          http.Server
	embed      *word2vec.Embeddings
	key1       = mysql.AnonFunc()
)

// function for the rest api, respond with the slice of all courses
func allInfo(c echo.Context) error {
	// can only return listing results, commentUser and commentItem
	dataPacket1, err1 := readJSONBody(c) // read response JSON

	if err1 != nil {
		fmt.Println(err1)
		return newErrorResponse(c, "Bad Request", http.StatusBadRequest)
	}

	receiveInfo := dataPacket1["DataInfo"].(map[string]string)

	tarDB := dataPacket1["InfoType"].(string)
	tarID := receiveInfo["ID"]

	switch tarDB { // gets all post

	case "ItemListing":

		sendInfo, _ := dbHandler1.GetRecordlisting(tarDB, receiveInfo["Name"], receiveInfo["filterUsername"], embed)
		return newResponse(c, sendInfo, "nil", "ItemListing", "true", "", http.StatusOK)

	// case "CommentUser": // gets all comments regarding a particular user id
	// 	sendInfo, _ := dbHandler1.GetRecord(tarDB)
	// 	newSendInfo := []interface{}{}

	// 	for i := range sendInfo {
	// 		temp1 := sendInfo[i].(mysql.CommentUser)

	// 		if temp1.ForUsername == tarID {
	// 			newSendInfo = append(newSendInfo, sendInfo[i])
	// 		}
	// 	}

	// 	w.WriteHeader(http.StatusCreated)
	// 	newResponse(c, newSendInfo, "nil", "ItemListing", "true", "")
	// 	return

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

		return newResponse(c, newSendInfo, "nil", "ItemListing", "true", "", http.StatusOK)

	default:

		return newErrorResponse(c, "404", http.StatusNotFound)
	}
}

func pwCheck(c echo.Context) error { //works
	dataPacket1, err := readJSONBody(c) // read response JSON

	if err != nil { //err means username not found, ok to proceed
		return newErrorResponse(c, "Bad Request", http.StatusBadRequest)
	}

	receiveInfo := mapInterfaceToString(dataPacket1) // convert received data into map[string]string
	dbData, err1 := dbHandler1.GetSingleRecord("UserSecret", "WHERE Username = ?", receiveInfo["Username"])

	if err1 != nil || len(dbData) == 0 { //err means username not found, ok to proceed

		return newErrorResponse(c, "", http.StatusBadRequest)
	}

	// update last login
	dbData1 := dbData[0].(mysql.UserSecret)
	err3 := bcrypt.CompareHashAndPassword([]byte(dbData1.Password), []byte(receiveInfo["Password"]))

	if err3 == nil {
		//update lastLogin if there is no issues
		dbHandler1.EditRecord("UserInfo", receiveInfo)
		//return response with true if no issues
		return newResponse(c, []interface{}{}, "nil", "UserInfo", "true", "", http.StatusOK)
	}

	return newErrorResponse(c, "", http.StatusBadRequest)
}

// change map[string]interface to map[string]string
func mapInterfaceToString(dataPacket1 map[string]interface{}) map[string]string {

	fmt.Println(dataPacket1["DataInfo"])

	receiveInfoRaw := dataPacket1["DataInfo"].([]interface{})[0] // convert received data into map[string]string
	receiveInfo := make(map[string]string)

	for k, v := range receiveInfoRaw.(map[string]interface{}) {
		receiveInfo[k] = fmt.Sprintf("%v", v)
	}

	return receiveInfo
}

// checks if the username in the sent info is currently in DB
// returns false if username is not taken
func usernameCheck(c echo.Context) error {
	dataPacket1, err1 := readJSONBody(c) // read response JSON
	if err1 != nil {                     //err means username not found, ok to proceed
		return newErrorResponse(c, "Bad Request", http.StatusBadRequest)
	}

	receiveInfo := mapInterfaceToString(dataPacket1)
	allData, err := dbHandler1.GetSingleRecord("UserInfo", " WHERE Username = ?", receiveInfo["Username"])

	if err != nil || len(allData) == 0 { //err means username not found, ok to proceed

		return newResponse(c, []interface{}{}, "username not found", "UserInfo", "false", "", http.StatusOK)
	}

	return newResponse(c, []interface{}{}, "nil", "UserInfo", "true", "", http.StatusOK)
}

// the function that writes the response back
func newResponse(c echo.Context, dataInfo []interface{}, errorMsg string, infoType string, resBool string, requestUser string, httpStatus int) error {
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

	fmt.Println(dataPacket1)
	return c.JSON(httpStatus, dataPacket1) // encode to json and send
}

// Response writer for when returning an error to a query
func newErrorResponse(c echo.Context, errorMsg string, httpStatus int) error {
	var DataInfo1 []interface{}
	return newResponse(c, DataInfo1, errorMsg, "error", "false", "", httpStatus)
}

// function to read the JSON on a request
func readJSONBody(c echo.Context) (map[string]interface{}, error) {
	// decode JSON body
	json_map := make(map[string]interface{})
	err1 := json.NewDecoder(c.Request().Body).Decode(&json_map)
	if err1 == nil {

		if json_map["Key"] != key1() {
			newErrorResponse(c, "403", http.StatusForbidden)
			return json_map, errors.New("incorrect api key supplied")
		}
		return json_map, nil
	}

	return json_map, errors.New("error while attempting to read body of request")
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
func genInfoPost(c echo.Context) error {
	dataPacket1, err1 := readJSONBody(c) // read response JSON
	if err1 != nil {
		fmt.Println(err1)
		return newErrorResponse(c, "Bad Request", http.StatusBadRequest)
	}

	receiveInfoRaw := mapInterfaceToString(dataPacket1) // convert received data into map[string]string
	tarDB := dataPacket1["InfoType"].(string)

	err2 := dbHandler1.InsertRecord(tarDB, receiveInfoRaw, "") // deletes if target is found
	if err2 == nil {

		return newResponse(c, []interface{}{}, "nil", "userInfo", "true", "", http.StatusOK)

	}

	fmt.Println("logger: insert, " + tarDB + ": " + tarDB + " db not found") //reach here only if it is not returned by the switch

	return newErrorResponse(c, "Bad Request , item not found", http.StatusBadRequest)

}

func getItem(c echo.Context, tarDB string, tarItemID string) ([]interface{}, error) {
	dbInfoSlice, err3 := dbHandler1.GetSingleRecord(tarDB, " WHERE ID = ?", tarItemID)

	if err3 != nil || len(dbInfoSlice) == 0 {

		if tarDB == "UserInfo" {
			dbInfoSlice, err3 = dbHandler1.GetSingleRecord(tarDB, " WHERE Username = ?", tarItemID)
		}

		if err3 != nil || len(dbInfoSlice) == 0 {

			fmt.Println("logger: error when looking up id " + tarItemID + " for DB " + tarDB + ", err:" + err3.Error())
			newErrorResponse(c, "Bad Request , item not found", http.StatusBadRequest)
			return []interface{}{}, errors.New("403")
		}

	}
	return dbInfoSlice, err3
}

func genInfoGet(c echo.Context) error {
	dataPacket1, err1 := readJSONBody(c) // read response JSON
	if err1 != nil {
		fmt.Println(err1)
		return newErrorResponse(c, "Bad Request", http.StatusBadRequest)
	}

	receiveInfoRaw := mapInterfaceToString(dataPacket1) // convert received data into map[string]string

	tarItemID := receiveInfoRaw["ID"]
	tarDB := dataPacket1["InfoType"].(string)

	dbInfoSlice, err3 := getItem(c, tarDB, tarItemID)

	// for userinfo, itemlisting, commentitem and commentuser only

	if tarDB != "UserSecret" && err3 == nil { //prevents any requeset to ask for user secrets
		return newResponse(c, dbInfoSlice, "nil", tarDB, "true", "", http.StatusCreated)
	}

	return newErrorResponse(c, "Bad Request , item not found", http.StatusBadRequest)
}

func genInfoDelete(c echo.Context) error {

	dataPacket1, err1 := readJSONBody(c) // read response JSON
	if err1 != nil {
		fmt.Println(err1)
		return newErrorResponse(c, "Bad Request", http.StatusBadRequest)
	}

	receiveInfoRaw := mapInterfaceToString(dataPacket1) // convert received data into map[string]string

	tarItemID := receiveInfoRaw["ID"]
	tarDB := dataPacket1["InfoType"].(string)

	dbInfoSlice, err3 := getItem(c, tarDB, tarItemID)

	if !checkUser(tarDB, dataPacket1["RequestUser"].(string), dbInfoSlice) || err3 != nil {

		return newErrorResponse(c, "Bad Request", http.StatusBadRequest)
	}

	// for deleting an entry

	if tarDB == "ItemListing" || tarDB == "CommentUser" || tarDB == "CommentItem" { //only delete records for 3 items
		err2 := dbHandler1.DeleteRecord(tarDB, tarItemID) // attempt to delete record
		if err2 != nil {

			return newResponse(c, []interface{}{}, "nil", "userInfo", "true", "", http.StatusOK)
		}

	}
	//if delete did not occur
	fmt.Println("logger:  " + ": " + tarDB + " db not found or not in use for Delete func, err:")
	return newErrorResponse(c, "Bad Request", http.StatusBadRequest)

}

// PUT is for updating existing course
func genInfoPut(c echo.Context) error {

	dataPacket1, err1 := readJSONBody(c) // read response JSON
	if err1 != nil {
		fmt.Println(err1)
		return newErrorResponse(c, "Bad Request , item not found", http.StatusBadRequest)
	}

	receiveInfoRaw := mapInterfaceToString(dataPacket1) // convert received data into map[string]string
	tarDB := dataPacket1["InfoType"].(string)

	// dbInfoSlice, err3 := getItem(c, tarDB, tarItemID)

	err2 := dbHandler1.EditRecord(tarDB, receiveInfoRaw) // deletes if target is found

	if err2 == nil {

		return newResponse(c, []interface{}{}, "nil", "CommentUser", "true", "", http.StatusCreated)
	}

	fmt.Println("logger: edit, " + tarDB + ": " + tarDB + " db not found") //reach here only if it is not returned by the switch

	return newErrorResponse(c, "Bad Request , item not found", http.StatusBadRequest)

}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	embed = word2vec.GetWord2Vec()
	dbHandler1 = mysql.OpenDB()
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.GET("/api/v0/check", pwCheck)
	e.GET("/api/v0/allinfo", allInfo)
	e.GET("/api/v0/username", usernameCheck)
	e.POST("/api/v0/db/info", genInfoPost)
	e.GET("/api/v0/db/info", genInfoGet)
	e.DELETE("/api/v0/db/info", genInfoDelete)
	e.PUT("/api/v0/db/info", genInfoPut)
	fmt.Println("listening at port 5555")
	s = http.Server{Addr: ":5555", Handler: e}

	if err := s.ListenAndServeTLS("secure//cert.pem", "secure//key.pem"); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

}
