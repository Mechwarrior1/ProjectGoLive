package main

// cd C:\Projects\Go\src\ProjectGoLive\ProjectGoLive\api_server\
// go run rest_api.go mysql_api.go encryptdecrypt.go readword2vec.go
import (
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
	dbHandler1 dbHandler
	s          http.Server
	embed      *Embeddings
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
		sendInfo, _ := dbHandler1.getRecordlisting(tarDB, receiveInfo["Name"], receiveInfo["filterUsername"])
		w.WriteHeader(http.StatusCreated)
		newResponse(w, r, sendInfo, "nil", "ItemListing", "true", "")
		return

	case "CommentUser": // gets all comments regarding a particular user id
		sendInfo, _ := dbHandler1.getRecord(tarDB)
		newSendInfo := []interface{}{}

		for i := range sendInfo {

			temp1 := sendInfo[i].(commentUser)
			if temp1.ForUsername == tarID {
				newSendInfo = append(newSendInfo, sendInfo[i])
			}
		}

		w.WriteHeader(http.StatusCreated)
		newResponse(w, r, newSendInfo, "nil", "ItemListing", "true", "")
		return

	case "CommentItem": // gets all comments regarding a particular item id
		sendInfo, _ := dbHandler1.getRecord(tarDB)
		newSendInfo := []interface{}{}

		for i := range sendInfo {
			temp1 := sendInfo[i].(commentItem)

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

	dbData, err1 := dbHandler1.getSingleRecord("UserSecret", "WHERE Username = \""+receiveInfo["Username"]+"\"")

	if err1 != nil || len(dbData) == 0 { //err means username not found, ok to proceed
		w.WriteHeader(http.StatusNotFound)
		newErrorResponse(w, r, "404 - Invalid credentials")
		return
	}

	// update last login
	dbData2, err2 := dbHandler1.getSingleRecord("UserInfo", "WHERE Username = \""+receiveInfo["Username"]+"\"")
	dbDataInfo := dbData2[0].(userInfo)
	dbData1 := dbData[0].(userSecret)
	err3 := bcrypt.CompareHashAndPassword([]byte(dbData1.Password), []byte(receiveInfo["Password"]))

	if err2 == nil && err3 == nil {
		w.WriteHeader(http.StatusCreated)
		//update lastLogin if there is no issues
		_ = dbHandler1.editRecord("UserInfo", dbDataInfo.Username, receiveInfo["LastLogin"], dbDataInfo.DateJoin, dbDataInfo.CommentItem, dbDataInfo.ID)
		//return response with true if no issues
		newResponse(w, r, []interface{}{}, "nil", "UserInfo", "true", "")
		return
	}

	w.WriteHeader(http.StatusNotFound)
	newErrorResponse(w, r, "402 - Invalid credentials")
}

// change map[string]interface to map[string]string
func mapInterfaceToString(dataPacket1 *dataPacket) map[string]string {
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
	allData, err := dbHandler1.getSingleRecord("UserInfo", " WHERE Username = \""+receiveInfo["Username"]+"\"")

	if err != nil || len(allData) == 0 { //err means username not found, ok to proceed
		w.WriteHeader(http.StatusNotFound)
		newResponse(w, r, []interface{}{}, "username not found", "UserInfo", "false", "")
		return
	}

	w.WriteHeader(http.StatusCreated)
	newResponse(w, r, []interface{}{}, "nil", "UserInfo", "true", "")
}

// // function to shutdown the api service
// func shutdown(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == "GET" {
// 		dataPacket1, err1 := readJSONBody(w, r)
// 		if err1 != nil {
// 			w.WriteHeader(http.StatusNotFound)
// 			newErrorResponse(w, r, err1.Error())
// 			return
// 		}
// 		if dataPacket1.DataInfo[0].Username == "shutdown" {
// 			newResponse(w, r, []courseInfo{}, "nil", "UserInfo", "true", "")
// 			fmt.Println("Shutting down...")
// 			go func() {
// 				time.Sleep(1 * time.Second)
// 				if err3 := s.Shutdown(context.Background()); err3 != nil {
// 					fmt.Println(err3)
// 				}
// 			}()
// 		}
// 	} else {
// 		newErrorResponse(w, r, "method is wrong")
// 	}
// }

// the function that writes the response back
// needs the []courseInfo
func newResponse(w http.ResponseWriter, r *http.Request, dataInfo []interface{}, errorMsg string, infoType string, resBool string, requestUser string) {
	if errorMsg == "nil" {
		w.WriteHeader(http.StatusCreated)
	}

	var dataPacket1 dataPacket
	dataPacket1.DataInfo = dataInfo
	dataPacket1.ErrorMsg = errorMsg       // error msg if any
	dataPacket1.InfoType = infoType       // to access which db
	dataPacket1.ResBool = resBool         //
	dataPacket1.RequestUser = requestUser //request coming from which user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataPacket1)

}

// Response writer for when query is an error
func newErrorResponse(w http.ResponseWriter, r *http.Request, errorMsg string) {
	var DataInfo1 []interface{}
	newResponse(w, r, DataInfo1, errorMsg, "error", "false", "")
}

// function to read the JSON on a request
func readJSONBody(w http.ResponseWriter, r *http.Request) (*dataPacket, error) {
	var newDataPacket dataPacket
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

func checkUser(tarDB string, requestUser string, dataInfo []interface{}) bool {

	switch tarDB {
	case "UserInfo":
		dataInfo1 := dataInfo[0].(userInfo)
		return dataInfo1.Username == requestUser

	case "ItemListing":
		dataInfo1 := dataInfo[0].(itemListing)
		return dataInfo1.Username == requestUser

	case "CommentUser":
		dataInfo1 := dataInfo[0].(commentUser)
		return dataInfo1.Username == requestUser

	case "CommentItem":
		dataInfo1 := dataInfo[0].(commentItem)
		return dataInfo1.Username == requestUser

	case "UserSecret":
		dataInfo1 := dataInfo[0].(userSecret)
		return dataInfo1.Username == requestUser

	default:
		fmt.Println("logger: " + tarDB + " is not a valid db")
		return false
	}

}

// main function for the rest api.
// GET request returns the specified course.
// DELETE request deletes the specified course.
// POST request will post the course into the DB.
// PUT request will edit the specified course.
func genInfo(w http.ResponseWriter, r *http.Request) {
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

	tarItemID := receiveInfo["ID"]
	tarDB := dataPacket1.InfoType
	if r.Method == "POST" {
		// check if course exists; add only if course does not exist
		maxID, err4 := dbHandler1.getMaxID(tarDB)
		maxIDString := fmt.Sprintf("%06d", maxID+1) //get current max ID in DB and increment by 1
		var err2 error
		switch tarDB {

		case "UserSecret":
			err2 = dbHandler1.insertRecord(tarDB, maxIDString, receiveInfo["Username"], receiveInfo["Password"], receiveInfo["IsAdmin"], receiveInfo["CommentItem"]) // deletes if target is found

			if err2 == nil {
				w.WriteHeader(http.StatusCreated)
				newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
				return
			}

		case "CommentUser":
			err2 = dbHandler1.insertRecord(tarDB, maxIDString, receiveInfo["Username"], receiveInfo["ForUsername"], receiveInfo["Date"], receiveInfo["CommentItem"]) // deletes if target is found

			if err2 == nil {
				w.WriteHeader(http.StatusCreated)
				newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
				return
			}

		case "CommentItem":
			err2 = dbHandler1.insertRecord(tarDB, maxIDString, receiveInfo["Username"], receiveInfo["ForItem"], receiveInfo["Date"], receiveInfo["CommentItem"])

			if err2 == nil {
				w.WriteHeader(http.StatusCreated)
				newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
				return
			}

		case "UserInfo":
			err2 = dbHandler1.insertRecord(tarDB, maxIDString, receiveInfo["Username"], receiveInfo["LastLogin"], receiveInfo["DateJoin"], receiveInfo["CommentItem"])

			if err2 == nil {
				w.WriteHeader(http.StatusCreated)
				newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
				return
			}

		case "ItemListing":
			// dbInfoSlice := dbInfoSlice.(itemListing)
			err2 = dbHandler1.insertRecord(tarDB,
				maxIDString,
				receiveInfo["Username"],
				receiveInfo["Name"],
				receiveInfo["ImageLink"],
				receiveInfo["DatePosted"],
				receiveInfo["CommentItem"],
				receiveInfo["ConditionItem"],
				receiveInfo["Cat"],
				receiveInfo["ContactMeetInfo"],
				receiveInfo["Completion"])

			if err2 == nil {
				w.WriteHeader(http.StatusCreated)
				newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
				return
			}

		default:
			fmt.Println("logger: insert, " + tarDB + ": " + tarDB + " db not found") //reach here only if it is not returned by the switch
			w.WriteHeader(http.StatusConflict)
			newErrorResponse(w, r, "404 - not found")
			return
		}

		if err2 != nil {
			dbHandler1.deleteRecord(tarDB, maxIDString) // if an error occured while posting to db, attempt to delete the entry
		}

		fmt.Println("logger: edit, " + tarDB + ":" + err2.Error()) //reach here only if it is not returned by the switch
		w.WriteHeader(http.StatusConflict)
		newErrorResponse(w, r, err4.Error())
		return
	}

	// if request is not post, check
	dbInfoSlice, err3 := dbHandler1.getSingleRecord(tarDB, " WHERE ID = "+tarItemID)
	if err3 != nil || len(dbInfoSlice) == 0 {

		if tarDB == "UserInfo" {
			dbInfoSlice, err3 = dbHandler1.getSingleRecord(tarDB, " WHERE Username = \""+tarItemID+"\"")
		}

		if err3 != nil || len(dbInfoSlice) == 0 {

			fmt.Println("logger: error when looking up id " + tarItemID + " for DB " + tarDB + ", err:" + err3.Error())
			w.WriteHeader(http.StatusNotFound)
			newErrorResponse(w, r, "404 - No item found")
			return
		}

	}

	// for userinfo, itemlisting, commentitem and commentuser only
	if r.Method == "GET" {

		if tarDB != "UserSecret" {
			w.WriteHeader(http.StatusCreated)
			newResponse(w, r, dbInfoSlice, "nil", tarDB, "true", "")
			return
		}

		w.WriteHeader(http.StatusNotFound)
		newErrorResponse(w, r, "404 - No item found")
		return
	}

	if !checkUser(tarDB, dataPacket1.RequestUser, dbInfoSlice) { //checks if the user that requested deletion is the owner who posted the comment/listing
		fmt.Println("logger: " + dataPacket1.RequestUser + " tried to access an item not from user, item ID:" + receiveInfo["ID"] + ", DB: " + tarDB)
		w.WriteHeader(http.StatusNotFound)
		newErrorResponse(w, r, "404 - Item not found")
		return
	}

	//delete commentitem, commentuser, itemlisting only cannot delete user info
	if r.Method == "DELETE" {
		err2 := dbHandler1.deleteRecord(tarDB, tarItemID) // deletes if target is found

		if err2 != nil && (tarDB == "ItemListing" || tarDB == "CommentUser" || tarDB == "CommentItem") {
			// newResponse(w, r, allCourses, "username not found", "userInfo", "false", "")
			w.WriteHeader(http.StatusCreated)
			newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
			return

		} else {
			fmt.Println("logger:  " + ": " + tarDB + " db not found or not in use for Delete func, err:") //reach here only if it is not returned by the switch
			w.WriteHeader(http.StatusConflict)
			newErrorResponse(w, r, "404 - not found")
			return
		}
	}

	// PUT is updating existing course
	if r.Method == "PUT" {

		var err2 error
		switch tarDB {
		case "CommentUser":
			dbInfoSlice1 := dbInfoSlice[0].(commentUser)
			err2 = dbHandler1.editRecord(tarDB, dbInfoSlice1.Username, dbInfoSlice1.ForUsername, dbInfoSlice1.Date, receiveInfo["CommentItem"], receiveInfo["ID"]) // deletes if target is found

			if err2 == nil {
				w.WriteHeader(http.StatusCreated)
				newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
				return
			}

		case "CommentItem":
			dbInfoSlice1 := dbInfoSlice[0].(commentItem)
			err2 = dbHandler1.editRecord(tarDB, dbInfoSlice1.Username, dbInfoSlice1.ForItem, dbInfoSlice1.Date, receiveInfo["CommentItem"], receiveInfo["ID"])

			if err2 == nil {
				w.WriteHeader(http.StatusCreated)
				newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
				return
			}

		case "UserInfo":
			dbInfoSlice1 := dbInfoSlice[0].(userInfo)
			err2 = dbHandler1.editRecord(tarDB, dbInfoSlice1.Username, dbInfoSlice1.LastLogin, dbInfoSlice1.DateJoin, receiveInfo["CommentItem"], receiveInfo["ID"])

			if err2 == nil {
				w.WriteHeader(http.StatusCreated)
				newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
				return
			}

		case "ItemListing":
			dbInfoSlice1 := dbInfoSlice[0].(itemListing)
			//( &data1.Username, &data1.Name, &data1.ImageLink, &data1.DatePosted, &data1.CommentItem, &data1.ConditionItem, &data1.Cat, &data1.ContactMeetInfo, &data1.ID,)
			err2 = dbHandler1.editRecord(tarDB,
				dbInfoSlice1.Username,
				dbInfoSlice1.Name,
				receiveInfo["ImageLink"],
				dbInfoSlice1.DatePosted,
				receiveInfo["CommentItem"],
				receiveInfo["ConditionItem"],
				receiveInfo["Cat"],
				receiveInfo["ContactMeetInfo"],
				receiveInfo["Completion"],
				dbInfoSlice1.ID)

			if err2 == nil {
				w.WriteHeader(http.StatusCreated)
				newResponse(w, r, []interface{}{}, "nil", "userInfo", "true", "")
				return
			}

		default:
			fmt.Println("logger: edit, " + tarDB + ": " + tarDB + " db not found") //reach here only if it is not returned by the switch
			w.WriteHeader(http.StatusConflict)
			newErrorResponse(w, r, "404 "+tarDB+"- not found")
			return
		}

		fmt.Println("logger: edit, " + tarDB + ":" + err2.Error()) //reach here only if it is not returned by the switch
		w.WriteHeader(http.StatusConflict)
		newErrorResponse(w, r, err2.Error())
		return

	}
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	embed = getWord2Vec()
	dbHandler1 = openDB()
	defer dbHandler1.DB.Close()

	router := mux.NewRouter()
	router.HandleFunc("/api/v0/check", pwCheck)
	router.HandleFunc("/api/v0/allinfo", allInfo)
	router.HandleFunc("/api/v0/username", usernameCheck)
	// router.HandleFunc("/api/v0/shutdown", shutdown)
	router.HandleFunc("/api/v0/db/info", genInfo).Methods("GET", "PUT", "POST", "DELETE")
	fmt.Println("listening at port 5555")
	s = http.Server{Addr: ":5555", Handler: router}

	if err := s.ListenAndServeTLS("secure//cert.pem", "secure//key.pem"); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// fmt.Println(dbHandler1.getRecord("UserSecret"))  //works
	// fmt.Println(dbHandler1.getRecord("UserInfo"))    //works
	// fmt.Println(dbHandler1.getRecord("ItemListing")) //works
	// fmt.Println(dbHandler1.getRecord("CommentUser")) //works
	// fmt.Println(dbHandler1.getRecord("CommentItem")) //works

	// dbHandler1.insertRecord("UserSecret", "000002", "admin2", "admin2", "false", "not a real admin")
	// dbHandler1.insertRecord("UserInfo", "000002", "admin2", "19/6/2021", "19/6/2021", "not a real admin")
	// dbHandler1.insertRecord("ItemListing",
	// "000002",
	// "admin2",
	// "float",
	// "https://cdn.education.com/files/86601_86700/86647/float.jpg",
	// "56345345",
	// "giving away a float",
	// "well used",
	// "misc",
	// "call me at 12345678, or meet at city hall mrt",
	// "false") //works
	// &data1.ID, &data1.Username, &data1.Name, &data1.ImageLink, &data1.DatePosted, &data1.CommentItem, &data1.ConditionItem, &data1.Cat, &data1.ContactMeetInfo
	// dbHandler1.insertRecord("CommentUser", "000002", "admin2", "admin", "21/6/2021", "make me an admin")      //works      //&data1.ID, &data1.Username, &data1.ForUsername, &data1.Date, &data1.CommentItem
	// dbHandler1.insertRecord("CommentItem", "000002", "admin2", "000001", "21/6/2021", "i love this game too") //works //&data1.ID, &data1.Username, &data1.ForItem, &data1.Date, &data1.CommentItem

	// editRecord(dbTable string, values ...interface{})
	// dbHandler1.editRecord("UserSecret", "admin3", "admin2", "false", "not a real admin", "000002")      //works
	// dbHandler1.editRecord("UserInfo", "admin3", "19/6/2021", "19/6/2021", "not a real admin", "000002")       //works
	// dbHandler1.editRecord("ItemListing", "000002", "admin3", "float", "https://cdn.education.com/files/86601_86700/86647/float.jpg", "56345345", "giving away a float", "well used", "misc", "call me at 12345678, or meet at city hall mrt","false", "000002") //works
	// //&data1.ID, &data1.Username, &data1.Name, &data1.ImageLink, &data1.DatePosted, &data1.CommentItem, &data1.ConditionItem, &data1.Cat, &data1.ContactMeetInfo
	// dbHandler1.editRecord("CommentUser", "admin3", "admin", "21/6/2021", "make me an admin", "000002")      //works      //&data1.ID, &data1.Username, &data1.ForUsername, &data1.Date, &data1.CommentItem
	// dbHandler1.editRecord("CommentItem", "admin3", "000001", "21/6/2021", "i love this game too", "000002") //works

	// getMaxID(dbTable string)
	// fmt.Println(dbHandler1.getMaxID("UserSecret"))
	// fmt.Println(dbHandler1.getMaxID("UserInfo"))
	// fmt.Println(dbHandler1.getMaxID("ItemListing"))
	// fmt.Println(dbHandler1.getMaxID("CommentUser"))
	// fmt.Println(dbHandler1.getMaxID("CommentItem"))

	// dbHandler1.deleteRecord("UserSecret", "000002")  //works
	// dbHandler1.deleteRecord("UserInfo", "000002")    //works
	// dbHandler1.deleteRecord("ItemListing", "000002") //works
	// dbHandler1.deleteRecord("CommentUser", "000002") //works
	// dbHandler1.deleteRecord("CommentItem", "000002") //works

	// fmt.Println(dbHandler1.getRecord("UserSecret"))  //works
	// fmt.Println(dbHandler1.getRecord("UserInfo"))    //works
	// fmt.Println(dbHandler1.getRecord("ItemListing")) //works
	// fmt.Println(dbHandler1.getRecord("CommentUser")) //works
	// fmt.Println(dbHandler1.getRecord("CommentItem")) //works
}

/*
	// so using interface to call the function to return pointers doesnt seem to work
	aa1 := userSecret{}
	var data1 genData = aa1
	// aa, _, _, _, ee := data1.returnParameter()
	results, _ := dbHandler1.DB.Query("Select * FROM my_db." + "UserInfo")
	fmt.Println(results.Scan(data1.returnParameter()))
	fmt.Println(aa1, data1)
	out, _ := json.Marshal(aa1)
	fmt.Println(string(out))

*/
