package controller

import (
	"apiserver/mysql"
	"apiserver/word2vec"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
)

// function for the rest api, respond with the slice of all courses
func GetAllListing(c echo.Context, dbHandler1 *mysql.DBHandler, embed *word2vec.Embeddings) error {
	// can only return listing results, commentUser and commentItem
	itemName := c.QueryParam("name")
	filterUsername := c.QueryParam("filter")
	fmt.Println(itemName, filterUsername)
	sendInfo, _ := dbHandler1.GetRecordlisting("ItemListing", itemName, filterUsername, embed)
	return newResponse(c, sendInfo, "nil", "ItemListing", "true", "", http.StatusOK)
}

func GetAllComment(c echo.Context, dbHandler1 *mysql.DBHandler, embed *word2vec.Embeddings) error {
	// gets all comments regarding a particular item id
	itemID := c.Param("id")
	sendInfo, _ := dbHandler1.GetRecord("CommentItem")
	newSendInfo := []interface{}{}

	for i := range sendInfo {
		temp1 := sendInfo[i].(mysql.CommentItem)

		if temp1.ForItem == itemID {
			// fmt.Println(sendInfo[i])
			newSendInfo = append(newSendInfo, sendInfo[i])
		}
	}

	return newResponse(c, newSendInfo, "nil", "ItemListing", "true", "", http.StatusOK)
}

func PwCheck(c echo.Context, dbHandler1 *mysql.DBHandler) error { //works
	dataPacket1, err := readJSONBody(c, dbHandler1) // read response JSON
	fmt.Println(dataPacket1)

	if err != nil { //err means username not found, ok to proceed
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}

	receiveInfo := mapInterfaceToString(dataPacket1) // convert received data into map[string]string
	dbData, err1 := dbHandler1.GetSingleRecord("UserSecret", "WHERE Username = ?", receiveInfo["Username"])
	dbData2, err2 := dbHandler1.GetSingleRecord("UserInfo", "WHERE Username = ?", receiveInfo["Username"])

	if err1 != nil || err2 != nil || len(dbData) == 0 {
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}

	// update last login
	dbData1 := dbData[0].(mysql.UserSecret)
	dbData3 := dbData2[0].(mysql.UserInfo)
	err3 := bcrypt.CompareHashAndPassword([]byte(dbData1.Password), []byte(receiveInfo["Password"]))
	fmt.Println(err3)
	if err3 == nil {
		//update lastLogin if there is no issues

		receiveInfo["CommentItem"] = dbData3.CommentItem
		dbHandler1.EditRecord("UserInfo", receiveInfo)

		payload := make(map[string]string)
		payload["LastLogin"] = dbData3.LastLogin
		payload["IsAdmin"] = dbData1.IsAdmin
		if dbData1.IsAdmin == "" {
			fmt.Println("there seems to be an error with the sql request for user admin and lastlogin info")
			fmt.Println(dbData1, dbData3)
		}

		//return response with true if no issues
		return newResponse(c, []interface{}{payload}, "nil", "ItemListing", "true", "", http.StatusOK)
	}

	return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
}

// change map[string]interface to map[string]string
func mapInterfaceToString(dataPacket1 map[string]interface{}) map[string]string {

	receiveInfoRaw := dataPacket1["DataInfo"].([]interface{})[0] // convert received data into map[string]string
	receiveInfo := make(map[string]string)

	for k, v := range receiveInfoRaw.(map[string]interface{}) {
		receiveInfo[k] = fmt.Sprintf("%v", v)
	}

	return receiveInfo
}

// checks if the username in the sent info is currently in DB
// returns false if username is not taken
func CheckUsername(c echo.Context, dbHandler1 *mysql.DBHandler) error {
	username := c.Param("username")
	fmt.Println(c.Path())

	//check received input
	if username == "" {
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}

	allData, err := dbHandler1.GetSingleRecord("UserInfo", " WHERE Username = ?", username)

	if err != nil || len(allData) == 0 { //err means username not found, ok to proceed
		return newResponseSimple(c, "username not found", "false", http.StatusOK)
	}

	return newResponseSimple(c, "username found", "true", http.StatusOK)
}

// the function that writes the response back
// for when you need to return arrays of entries
func newResponse(c echo.Context, dataInfo []interface{}, errorMsg string, infoType string, resBool string, requestUser string, httpStatus int) error {
	var responseJson mysql.DataPacket
	responseJson.DataInfo = dataInfo
	responseJson.ErrorMsg = errorMsg       // error msg if any
	responseJson.InfoType = infoType       // to access which db
	responseJson.ResBool = resBool         //
	responseJson.RequestUser = requestUser //request coming from which user

	// fmt.Println(responseJson)
	fmt.Println(responseJson)
	return c.JSON(httpStatus, responseJson) // encode to json and send
}

// the function that writes the response back
func newResponseSimple(c echo.Context, msg string, resBool string, httpStatus int) error {
	responseJson := mysql.DataPacketSimple{
		msg,
		resBool,
	}
	fmt.Println(responseJson)
	return c.JSON(httpStatus, responseJson) // encode to json and send
}

// function to read the JSON on a request
func readJSONBody(c echo.Context, dbHandler1 *mysql.DBHandler) (map[string]interface{}, error) {
	// decode JSON body
	fmt.Println("1, \n", c.Request())
	json_map := make(map[string]interface{})
	err1 := json.NewDecoder(c.Request().Body).Decode(&json_map)
	if err1 == nil {

		if json_map["Key"] != dbHandler1.ApiKey {
			newResponseSimple(c, "Forbidden", "false", http.StatusForbidden)
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

// post sign up
func Signup(c echo.Context, dbHandler *mysql.DBHandler) error {
	dataPacket1, err1 := readJSONBody(c, dbHandler) // read response JSON
	if err1 != nil {
		fmt.Println(err1)
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}

	receiveInfo := mapInterfaceToString(dataPacket1) // convert received data into map[string]string
	fmt.Println(receiveInfo)

	// get current max number
	maxIDString := "0"
	maxID, err1 := dbHandler.GetMaxID("UserSecret")

	if err1 == nil {
		maxIDString = fmt.Sprintf("%06d", maxID+1)
	}

	// begin ticket
	tx, err := dbHandler.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	//first statement
	stmt, err1 := tx.Prepare("INSERT INTO my_db.UserSecret VALUES (?, ?, ?, ?, ?)")
	if err1 != nil {
		err = err1
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}

	_, err = stmt.Exec(maxIDString, receiveInfo["Username"], receiveInfo["Password"], receiveInfo["IsAdmin"], receiveInfo["CommentItem"])
	stmt.Close()

	//2nd statement
	stmt, err2 := tx.Prepare("INSERT INTO my_db.UserInfo VALUES (?, ?, ?, ?, ?)")
	if err2 != nil {
		err = err2
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}
	fmt.Println(err)
	_, err = stmt.Exec(maxIDString, receiveInfo["Username"], receiveInfo["LastLogin"], receiveInfo["DateJoin"], receiveInfo["CommentItem"])
	stmt.Close()

	//check error
	if err != nil {
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}

	return newResponseSimple(c, "nil", "true", http.StatusOK)

}

// main function for the rest api.
// GET request returns the specified post/user/comments.
// POST request will post the post/user/comments into the DB.
// PUT request will edit the specified post/user/comments.
func GenInfoPost(c echo.Context, dbHandler1 *mysql.DBHandler) error {
	dataPacket1, err1 := readJSONBody(c, dbHandler1) // read response JSON
	if err1 != nil {
		fmt.Println(err1)
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}

	receiveInfoRaw := mapInterfaceToString(dataPacket1) // convert received data into map[string]string
	tarDB := dataPacket1["InfoType"].(string)

	err2 := dbHandler1.InsertRecord(tarDB, receiveInfoRaw, "") // deletes if target is found

	if err2 == nil {
		return newResponseSimple(c, "nil", "true", http.StatusBadRequest)
	}

	fmt.Println("logger: insert " + tarDB + " db not found") //reach here only if it is not returned by the switch

	return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)

}

func getItem(c echo.Context, tarDB string, tarItemID string, dbHandler1 *mysql.DBHandler) ([]interface{}, error) {
	dbInfoSlice, err3 := dbHandler1.GetSingleRecord(tarDB, " WHERE ID = ?", tarItemID)

	if err3 != nil || len(dbInfoSlice) == 0 {

		if tarDB == "UserInfo" {
			dbInfoSlice, err3 = dbHandler1.GetSingleRecord(tarDB, " WHERE Username = ?", tarItemID)
		}

		if err3 != nil || len(dbInfoSlice) == 0 {
			// fmt.Println("logger: error when looking up id " + tarItemID + " for DB " + tarDB + ", err:" + err3.Error())

			newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
			return []interface{}{}, errors.New("403")
		}

	}
	return dbInfoSlice, err3
}

func GenInfoGet(c echo.Context, dbHandler1 *mysql.DBHandler) error {
	itemID := c.QueryParam("id")
	itemDB := c.QueryParam("db")

	dbInfoSlice, err3 := getItem(c, itemDB, itemID, dbHandler1)

	// for userinfo, itemlisting, commentitem and commentuser only
	if itemDB != "UserSecret" && err3 == nil { //prevents any requeset to ask for user secrets
		return newResponse(c, dbInfoSlice, "nil", itemDB, "true", "false", http.StatusCreated)
	}

	return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
}

func Completed(c echo.Context, dbHandler *mysql.DBHandler) error {
	itemID := c.Param("id")
	dataPacket1, err1 := readJSONBody(c, dbHandler)

	username, err := dbHandler.GetUsername("ItemListing", itemID)

	// check if the username is the same
	if username != dataPacket1["RequestUser"] && err1 != nil {
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}

	tx, err := dbHandler.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	//first statement
	stmt, err1 := tx.Prepare("UPDATE my_db.ItemListing SET Completion =? WHERE ID=?")
	if err1 != nil {
		err = err1
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}

	_, err = stmt.Exec("true", itemID)
	stmt.Close()

	//check error
	if err != nil {
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}

	return newResponseSimple(c, "nil", "true", http.StatusOK)
}

/// delete is not implemented currently, might change to soft delete instead
// func GenInfoDelete(c echo.Context, dbHandler1 *mysql.DBHandler) error {

// 	itemID := c.QueryParam("id")
// 	itemDB := c.QueryParam("db")
// 	apiKey := c.QueryParam("key")
// 	if dbHandler1.Apikey != apiKey{
// 				return newErrorResponse(c, "Bad Request", http.StatusBadRequest)
// 	}

// dbInfoSlice, err3 := getItem(c, tarDB, tarItemID, dbHandler1)

// if !checkUser(tarDB, dataPacket1["RequestUser"].(string), dbInfoSlice) || err3 != nil {

// 	return newErrorResponse(c, "Bad Request", http.StatusBadRequest)
// }

// for deleting an entry

// 	if tarDB == "ItemListing" || tarDB == "CommentUser" || tarDB == "CommentItem" { //only delete records for 3 items
// 		err2 := dbHandler1.DeleteRecord(tarDB, tarItemID) // attempt to delete record
// 		if err2 != nil {

// 			return newResponse(c, []interface{}{}, "nil", "userInfo", "true", "", http.StatusOK)
// 		}

// 	}
// 	//if delete did not occur
// 	fmt.Println("logger:  " + ": " + tarDB + " db not found or not in use for Delete func, err:")
// 	return newErrorResponse(c, "Bad Request", http.StatusBadRequest)

// }

// PUT is for updating existing course
func GenInfoPut(c echo.Context, dbHandler1 *mysql.DBHandler) error {

	dataPacket1, err1 := readJSONBody(c, dbHandler1) // read response JSON
	if err1 != nil {
		fmt.Println(err1)
		return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)
	}

	receiveInfoRaw := mapInterfaceToString(dataPacket1) // convert received data into map[string]string

	err2 := dbHandler1.EditRecord(dataPacket1["InfoType"].(string), receiveInfoRaw) // deletes if target is found

	if err2 == nil {
		return newResponseSimple(c, "nil", "true", http.StatusBadRequest)
	}

	return newResponseSimple(c, "Bad Request", "false", http.StatusBadRequest)

}
