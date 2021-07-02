// Runs the http server to manage movies.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	// address for rest api
	baseURL = "https://localhost:5555/api/v0/"
)

// removes all special characters from a string
func replaceAllString(string1 string) string {
	reg, err := regexp.Compile(`[^a-zA-Z0-9!@#$%^&\s,.?_]+`)
	if err != nil {
		log.Fatal(err)
	}
	processedString := reg.ReplaceAllString(string1, "")
	return processedString
}

// the function to access the rest api
// requires the method and datapacket
// returns any courseinfo and error
func tapAPI(httpMethod string, jsonData dataPacket, baseURL1 string) (*dataPacket, error) {
	url := baseURL1
	jsonValue, _ := json.Marshal(jsonData)
	request, _ := http.NewRequest(httpMethod, url,
		bytes.NewBuffer(jsonValue))

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	var newDataPacket dataPacket

	if err != nil {
		logger1.logTrace("ERROR", "The HTTPS request failed with error: "+err.Error())
		return &newDataPacket, errors.New("https request failed with " + err.Error())

	} else {

		data1, err := ioutil.ReadAll(response.Body) //
		if err != nil {
			return &newDataPacket, errors.New("ioutil failed to read, error: " + err.Error())
		}

		json.Unmarshal(data1, &newDataPacket)
		response.Body.Close()

		if newDataPacket.ErrorMsg != "nil" {
			logger1.logTrace("ERROR", "Error encounted when reading datapacket: "+newDataPacket.ErrorMsg)
			return &newDataPacket, errors.New(newDataPacket.ErrorMsg)
		}

		return &newDataPacket, nil
	}
}

// check if user is logged in, else redirect
func checkLogged(res http.ResponseWriter, req *http.Request, userPersistInfo1 *userPersistInfo, id string) bool {
	if userPersistInfo1.Username == "" || userPersistInfo1.Username == "None" || userPersistInfo1.Username == "logout********" {
		sessionMgr.updatePersistInfo(id, "false", "None", "true1", "Please kindly log in to access page", "", "seelast", false)
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return false
	}
	return true
}

// a function for http handler, used to create any course.
// bring to a form for inputs
func createPost(res http.ResponseWriter, req *http.Request) {
	id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
	if !checkLogged(res, req, userPersistInfo1, id) {
		return
	}
	if req.Method == http.MethodPost {
		// get form values.
		postName := req.FormValue("PostName")
		postComment := req.FormValue("PostComment")
		postImg2 := req.FormValue("PostImg2")
		if strings.Contains(postImg2, "script") || strings.Contains(postImg2, "scr") { // returns error if the word script is found in link
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "Please fill in the form correctly, without special characters for the name and title", "", "seelast", false)
			http.Redirect(res, req, "/createpost", http.StatusSeeOther)
			return
		}

		postCondition := req.FormValue("PostCondition")
		postCat := req.FormValue("PostCat")
		postContactMeetInfo := req.FormValue("PostContactMeetInfo")

		for _, string1 := range []string{postName, postComment, postCondition, postCat, postContactMeetInfo} {

			if replaceAllString(string1) != string1 {

				fmt.Println("unable to create post, due to string:", string1)
				sessionMgr.updatePersistInfo(id, "false", "None", "true", "Please fill in the form correctly, without special characters for the name and title", "", "seelast", false)
				http.Redirect(res, req, "/createpost", http.StatusSeeOther)
				return
			}
		}
		// take inputs and put into map for api/server
		newPost := make(map[string]string)
		newPost["Username"] = userPersistInfo1.Username
		newPost["Name"] = postName
		newPost["ImageLink"] = postImg2
		timenow := time.Now().Unix()
		newPost["DatePosted"] = strconv.Itoa(int(timenow))
		newPost["CommentItem"] = postComment
		newPost["ConditionItem"] = postCondition
		newPost["Cat"] = postCat
		newPost["ContactMeetInfo"] = postContactMeetInfo
		newPost["Completion"] = "false"

		jsonData1 := dataPacket{
			Key:         key1(), // key to access rest api
			ErrorMsg:    "nil",
			InfoType:    "ItemListing",
			ResBool:     "false",
			RequestUser: userPersistInfo1.Username,
			DataInfo:    []map[string]string{newPost},
		}

		address := baseURL + "db/info"
		_, err5 := tapAPI("POST", jsonData1, address)

		if err5 != nil {
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "An error has occurred, please try again later", "", "seelast", false)
			http.Redirect(res, req, "/createpost", http.StatusSeeOther)
			return
		}

		logger1.logTrace("TRACE", "Created item: '"+postName+"', by user: '"+userPersistInfo1.Username+""+"' desc: '"+postComment+"'")
		sessionMgr.updatePersistInfo(id, "true", "You have created item: '"+postName+"'", "false", "None", "", "seelast", false)
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	sessionMgr.updatePersistInfo(id, "false", "None", "false", "None", "", "seelast", false)
	tplCreatePost.ExecuteTemplate(res, "createpost.gohtml", userPersistInfo1)
}

// a function for http handler, used to edit the listed post
// a form for inputs
func editPost(res http.ResponseWriter, req *http.Request) {
	id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
	if !checkLogged(res, req, userPersistInfo1, id) {
		return
	}
	params := mux.Vars(req)
	mapListing := make(map[string]string)
	mapListing["Name"] = params["id"]
	mapListing["ID"] = params["id"]

	jsonData1 := dataPacket{
		// key to access rest api
		Key:         key1(),
		ErrorMsg:    "nil",
		InfoType:    "ItemListing",
		ResBool:     "false",
		RequestUser: userPersistInfo1.Username,
		DataInfo:    []map[string]string{mapListing},
	}

	dataPacket1, err1 := tapAPI(http.MethodGet, jsonData1, baseURL+"db/info")

	dataInsert := struct { //struct for inserting into go template
		DataInfo        map[string]string
		UserPersistInfo userPersistInfo
	}{
		dataPacket1.DataInfo[0],
		*userPersistInfo1,
	}

	if err1 != nil || dataPacket1.ResBool == "false" || userPersistInfo1.Username != dataPacket1.DataInfo[0]["Username"] {
		sessionMgr.updatePersistInfo(id, "false", "None", "true1", "An error has occurred or you do not have access to this page", "", "seelast", false)
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	if req.Method == http.MethodPost {
		// get form values
		postName := req.FormValue("PostName")
		postComment := req.FormValue("PostComment")
		postImg2 := req.FormValue("PostImg2")
		if strings.Contains(postImg2, "script") {
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "Please fill in the form correctly, without special characters for the name and title", "", "seelast", false)
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return
		}
		postCondition := req.FormValue("PostCondition")
		postCat := req.FormValue("PostCat")
		postContactMeetInfo := req.FormValue("PostContactMeetInfo")

		// check if form values contains any special characters, if so refreshes the page
		for _, string1 := range []string{postName, postComment, postCondition, postCat, postContactMeetInfo} {
			if replaceAllString(string1) != string1 {
				fmt.Println("unable to create post, due to string:", string1)
				sessionMgr.updatePersistInfo(id, "false", "None", "true", "Please fill in the form correctly, without special characters for the name and title", "", "seelast", false)
				http.Redirect(res, req, "/", http.StatusSeeOther)
				return
			}
		}

		//put inputs into map and push it into the api
		newPost := make(map[string]string)
		newPost["ID"] = params["id"]
		newPost["Username"] = userPersistInfo1.Username
		newPost["Name"] = postName
		newPost["ImageLink"] = postImg2
		timenow := time.Now().Unix()
		newPost["DatePosted"] = strconv.Itoa(int(timenow))
		newPost["CommentItem"] = postComment
		newPost["ConditionItem"] = postCondition
		newPost["Cat"] = postCat
		newPost["ContactMeetInfo"] = postContactMeetInfo
		newPost["Completion"] = dataPacket1.DataInfo[0]["Completion"]
		for k, v := range newPost {
			fmt.Println(k, ", ", v)
		}

		jsonData1.DataInfo = []map[string]string{newPost}
		_, err5 := tapAPI(http.MethodPut, jsonData1, baseURL+"db/info") //communicate with api
		// if error feedback to user
		if err5 != nil {
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "An error has occurred, please try again later", "", "seelast", false)
			http.Redirect(res, req, "/editpost/"+params["id"], http.StatusSeeOther)
			return
		}

		logger1.logTrace("TRACE", "Created item: '"+postName+"', by user: '"+userPersistInfo1.Username+""+"' desc: '"+postComment+"'")
		sessionMgr.updatePersistInfo(id, "true", "You have edited item: '"+postName+"'", "false", "None", "", "seelast", false)
		http.Redirect(res, req, "/getpost/"+params["id"], http.StatusSeeOther)
		return
	}

	sessionMgr.updatePersistInfo(id, "false", "None", "false", "None", "", "seelast", false)
	tplEditPost.ExecuteTemplate(res, "editpost.gohtml", dataInsert)
}

// func sorts the incoming dataInfo by similarity (each map has "similarity")
// based on how similar it is to the searched term
func sortPost(dataArr []map[string]string, date1 string, cat1 string, sort1 string) ([]map[string]string, []int) {
	// fmt.Println("sort start:", date1, cat1)
	newSorted := []map[string]string{}
	sortArr := []int{}
	sortArrSim := []float32{}

	for i, map1 := range dataArr {

		timenow := time.Now().Unix()
		switch date1 { //calculate the cut off date,
		case "7days":
			timenow = timenow - (7 * 24 * 60 * 60)
		case "30days":
			timenow = timenow - (30 * 24 * 60 * 60)
		}

		dateVal, _ := strconv.Atoi(map1["DatePosted"])
		map1["DatePosted"] = time.Unix(int64(dateVal), 0).Format("02-01-2006")

		// adds index of map into array if the map meets the criteria, before sorting it
		if (timenow < int64(dateVal) || date1 == "All" || date1 == "") && (cat1 == map1["Cat"] || cat1 == "All" || cat1 == "") {
			sortArr = append(sortArr, i)
			simVal, _ := strconv.ParseFloat(map1["Similarity"], 32)
			sortArrSim = append(sortArrSim, float32(simVal))
		}

	}
	//sorts the similarity and provides an index
	_, sortArr2 := mergeSort(sortArrSim, sortArr)
	maxLen := len(sortArr)

	if sort1 == "asc" {

		for idx := 0; idx < maxLen; idx++ { //sorts results in ascending order
			newSorted = append(newSorted, dataArr[sortArr2[idx]])
			fmt.Println("data :", dataArr[sortArr2[idx]]["Name"], dataArr[sortArr2[idx]]["Similarity"])
		}

	} else {

		for idx := maxLen - 1; idx >= 0; idx-- { //sorts results in descending order
			newSorted = append(newSorted, dataArr[sortArr2[idx]])
			fmt.Println("data :", dataArr[sortArr2[idx]]["Name"], dataArr[sortArr2[idx]]["Similarity"])
		}

	}
	return newSorted, sortArr2
}

// a function for http handler, used to get any course.
// has a button to bring to a form to see the selected course
func seePostAll(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		postSearch := req.FormValue("PostSearch")
		postDate := req.FormValue("PostDate")
		postCat := req.FormValue("PostCat")
		postSort := req.FormValue("PostSort")
		http.Redirect(res, req, "/seepost?search1="+postSearch+"&date="+postDate+"&cat="+postCat+"&sort="+postSort, http.StatusSeeOther)
		return

	} else {

		searchParam := req.URL.Query().Get("search1")
		dateParam := req.URL.Query().Get("date")
		catParam := req.URL.Query().Get("cat")
		sortParam := req.URL.Query().Get("sort")

		// res.Write([]byte("<script>alert('Please login')</script>"))

		// checks the search parameters for special characters, normally it is auto generated, but can be edited on search bar
		// redirects if there is an error in the search parameters
		for _, params := range []string{searchParam, dateParam, catParam, sortParam} {

			if replaceAllString(params) != params {
				http.Redirect(res, req, "/seepost/1?search1=&date=All&cat=All&sort=desc", http.StatusSeeOther)
				return
			}
		}

		id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
		mapListing := make(map[string]string)
		mapListing["Name"] = searchParam // only for ItemListing
		fmt.Println("search: [", searchParam, dateParam, catParam, sortParam, "] url:", req.URL.RequestURI())

		jsonData1 := dataPacket{
			// key to access rest api
			Key:         key1(),
			ErrorMsg:    "nil",
			InfoType:    "ItemListing",
			ResBool:     "false",
			RequestUser: userPersistInfo1.Username,
			DataInfo:    []map[string]string{mapListing},
		}

		dataPacket1, err1 := tapAPI("GET", jsonData1, baseURL+"allinfo")
		dataInfoSorted, _ := sortPost(dataPacket1.DataInfo, dateParam, catParam, sortParam)

		if err1 != nil {
			sessionMgr.updatePersistInfo(id, "false", "None", "true1", "An error has occurred, please try again later", "", "seelast", false)
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return
		}

		// data required by the go template
		dataInsert := struct {
			DataInfo        []map[string]string
			UserPersistInfo userPersistInfo
		}{
			dataInfoSorted,
			*userPersistInfo1,
		}

		sessionMgr.updatePersistInfo(id, "false", "None", "false", "None", "", "seelast", false)
		tplSeePostAll.ExecuteTemplate(res, "seepostall.gohtml", dataInsert)
		return
	}
}

// a function for http handler, follow up from getCourse, zooms into the course
func getPostDetail(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req) // get post id
	id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)

	// reuqesting the information for the post, using the post id
	mapListing := make(map[string]string)
	mapListing["Name"] = params["id"]
	mapListing["ID"] = params["id"] // only for ItemListing
	jsonData1 := dataPacket{
		// key to access rest api
		Key:         key1(),
		ErrorMsg:    "nil",
		InfoType:    "ItemListing",
		ResBool:     "false",
		RequestUser: userPersistInfo1.Username,
		DataInfo:    []map[string]string{mapListing},
	}

	dataPacket1, err1 := tapAPI(http.MethodGet, jsonData1, baseURL+"db/info")

	if err1 != nil || dataPacket1.ResBool == "false" || len(dataPacket1.DataInfo) == 0 {
		//if post id does not exist return to search page
		sessionMgr.updatePersistInfo(id, "false", "None", "true", "The detail: "+params["id"]+" cannot be found, please try another course", "", "seelast", false)
		http.Redirect(res, req, "/seepost/1?search1=&date=All&cat=All&sort=desc", http.StatusSeeOther)
		return
	}

	// request for Comments for the post, sending the post id to api, if id cannot be found, redirect
	jsonData1.InfoType = "CommentItem"
	dataPacket2, err2 := tapAPI(http.MethodGet, jsonData1, baseURL+"allinfo")

	if err2 != nil || dataPacket2.ResBool == "false" {
		// if pos
		sessionMgr.updatePersistInfo(id, "false", "None", "true", "The detail: "+params["id"]+" cannot be found, please try another course", "", "seelast", false)
		http.Redirect(res, req, "/seepost/1?search1=&date=All&cat=All&sort=desc", http.StatusSeeOther)
		return
	}

	// send data of post and its comments to the template for rendering
	postData := dataPacket1.DataInfo[0]
	dateVal, _ := strconv.Atoi(postData["DatePosted"])
	postData["DatePosted"] = time.Unix(int64(dateVal), 0).Format("02-01-2006")
	dataInsert := struct {
		PostInfo        map[string]string
		PostCommentInfo []map[string]string
		UserPersistInfo userPersistInfo
		Owner           bool
	}{
		postData,
		dataPacket2.DataInfo,
		*userPersistInfo1,
		postData["Username"] == userPersistInfo1.Username,
	}

	// post request for adding a new comment for the post
	if req.Method == http.MethodPost {

		username1 := userPersistInfo1.Username
		if !checkLogged(res, req, userPersistInfo1, id) {
			return
		}

		// put inputs into map to be sent to api
		postComment := req.FormValue("PostComment")
		mapComment := make(map[string]string)
		mapComment["CommentItem"] = postComment
		mapComment["Username"] = username1
		mapComment["ForItem"] = params["id"]
		mapComment["Date"] = time.Now().Format("02-01-2006")

		jsonData1.DataInfo = []map[string]string{mapComment}
		dataPacket3, err3 := tapAPI(http.MethodPost, jsonData1, baseURL+"db/info")

		// if response is an error
		if err3 == nil && dataPacket3.ResBool == "true" {
			sessionMgr.updatePersistInfo(id, "true", "You have posted a comment", "false", "None", "", "seelast", false)
		} else {
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "an error has occurred while trying to post a comment", "", "seelast", false)
		}
		http.Redirect(res, req, "/getpost/"+params["id"], http.StatusSeeOther)
		return
	}

	sessionMgr.updatePersistInfo(id, "false", "None", "false", "None", "", "seelast", false) // clean up feedback
	tplGetPostDetail.ExecuteTemplate(res, "getpostdetail.gohtml", dataInsert)
}

// see the post for a particular user
// filters for their post/listing
func seePostUser(res http.ResponseWriter, req *http.Request) {
	postUsername := mux.Vars(req)["id"] // get post id
	id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
	mapListing := make(map[string]string)
	mapListing["Name"] = ""
	mapListing["filterUsername"] = postUsername //get all post for the username

	jsonData1 := dataPacket{
		// key to access rest api
		Key:         key1(),
		ErrorMsg:    "nil",
		InfoType:    "ItemListing",
		ResBool:     "false",
		RequestUser: userPersistInfo1.Username,
		DataInfo:    []map[string]string{mapListing},
	}

	dataPacket1, err1 := tapAPI("GET", jsonData1, baseURL+"allinfo")
	// dataInfoSorted, _ := sortPost(dataPacket1.DataInfo, "All", "All", "desc")

	if err1 != nil || dataPacket1.ErrorMsg == "false" {
		sessionMgr.updatePersistInfo(id, "false", "None", "true1", "An error has occurred, or user has no post", "", "seelast", false)
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	dataInsert := struct {
		DataInfo        []map[string]string
		UserPersistInfo userPersistInfo
		PostUsername    string
	}{
		dataPacket1.DataInfo,
		*userPersistInfo1,
		postUsername,
	}

	sessionMgr.updatePersistInfo(id, "false", "None", "false", "None", "", "seelast", false)
	tplSeePostUser.ExecuteTemplate(res, "seepostuser.gohtml", dataInsert)
}

// changes the post to completed, and return the rest of the values to api
func postComplete(res http.ResponseWriter, req *http.Request) {
	id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
	params := mux.Vars(req)

	mapListing := make(map[string]string)
	mapListing["ID"] = params["id"]

	jsonData1 := dataPacket{
		// key to access rest api
		Key:         key1(),
		ErrorMsg:    "nil",
		InfoType:    "ItemListing",
		ResBool:     "false",
		RequestUser: userPersistInfo1.Username,
		DataInfo:    []map[string]string{mapListing},
	}

	dataPacket1, err1 := tapAPI(http.MethodGet, jsonData1, baseURL+"db/info")

	if err1 != nil || dataPacket1.ResBool == "false" || userPersistInfo1.Username != dataPacket1.DataInfo[0]["Username"] {
		sessionMgr.updatePersistInfo(id, "false", "None", "true1", "An error has occurred or you do not have access to this page", "", "seelast", false)
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	// change completion value, returns the rest
	mapListing2 := dataPacket1.DataInfo[0]
	mapListing2["Completion"] = "true"
	fmt.Println(params, mapListing2)
	jsonData1.DataInfo = []map[string]string{mapListing2}

	//if error feedback to user and
	_, err5 := tapAPI("PUT", jsonData1, baseURL+"db/info")
	if err5 != nil {
		sessionMgr.updatePersistInfo(id, "false", "None", "true", "An error has occurred, please try again later", "", "seelast", false)
		http.Redirect(res, req, "/", http.StatusSeeOther)
		logger1.logTrace("TRACE", "error encountered while changing status of '"+mapListing2["Name"]+"' to completed ")
		return
	}

	sessionMgr.updatePersistInfo(id, "false", "None", "false", "None", "", "seelast", false)
	logger1.logTrace("TRACE", "Updated Item status for '"+mapListing2["Name"]+"' to completed ")
	sessionMgr.updatePersistInfo(id, "true", "'"+mapListing2["Name"]+"' is tagged as completed", "false", "None", "", "seelast", false)
	http.Redirect(res, req, "/", http.StatusSeeOther)
}

// a function for http handler, for the main page of the site.
func index(res http.ResponseWriter, req *http.Request) {
	fmt.Println("index page")
	id, _ := sessionMgr.getIdPersistInfo(res, req)
	sessionMgr.removePersistInfoError(id)
	_, persistInfo1 := sessionMgr.getIdPersistInfo(res, req)
	sessionMgr.updatePersistInfo(id, "false", "None", "false", "None", "", "seelast", false)
	if req.Method == http.MethodPost {
		postSearch := req.FormValue("search1")
		postCat := req.FormValue("cat")
		http.Redirect(res, req, "/seepost?search1="+postSearch+"&date=All&cat="+postCat+"&sort=desc", http.StatusSeeOther)
		return
	}

	tplIndex.ExecuteTemplate(res, "index.gohtml", persistInfo1)
}

// a function for http handler, gives you the information on user
func getUser(res http.ResponseWriter, req *http.Request) {
	idParam := req.URL.Query().Get("id")
	editParam := req.URL.Query().Get("edit")
	id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
	// reuqesting the information for the user, using the post id
	mapListing := make(map[string]string)
	mapListing["ID"] = idParam // get info for ItemListing based on ID

	jsonData1 := dataPacket{
		// key to access rest api
		Key:         key1(),
		ErrorMsg:    "nil",
		InfoType:    "UserInfo",
		ResBool:     "false",
		RequestUser: userPersistInfo1.Username,
		DataInfo:    []map[string]string{mapListing},
	}
	dataPacket1, err1 := tapAPI(http.MethodGet, jsonData1, baseURL+"db/info")
	// if error in fetching data
	if err1 != nil || dataPacket1.ResBool == "false" || len(dataPacket1.DataInfo) == 0 {
		//if user id does not exist return to index page
		sessionMgr.updatePersistInfo(id, "false", "None", "true", "The detail: "+idParam+" cannot be found, please try another course", "", "seelast", false)
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	// send data of post and its comments to the template for rendering
	userData := dataPacket1.DataInfo[0]
	dataInsert := struct {
		UserData        map[string]string
		UserPersistInfo userPersistInfo
		Owner           bool
		Edit            bool
	}{
		userData,
		*userPersistInfo1,
		userData["Username"] == userPersistInfo1.Username,
		editParam == "true",
	}

	// post request for adding a new comment for the post
	if req.Method == http.MethodPost {

		// checks if the owner of the user page is the one editing
		username1 := userPersistInfo1.Username
		if username1 == "" || username1 != userData["Username"] {
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "An error has occurred", "", "seelast", false)
			http.Redirect(res, req, "/user?id="+idParam+"&edit=false", http.StatusSeeOther)
			return
		}

		// put inputs into map to be sent to api
		commentItem := req.FormValue("CommentItem")
		mapComment := make(map[string]string)
		mapComment["CommentItem"] = commentItem
		mapComment["Username"] = username1
		mapComment["ID"] = idParam
		mapComment["Date"] = userData["Date"]
		mapComment["LastLogin"] = userData["Date"]

		jsonData1.DataInfo = []map[string]string{mapComment}
		dataPacket3, err3 := tapAPI(http.MethodPut, jsonData1, baseURL+"db/info")
		// if error in posting a comment
		if err3 == nil && dataPacket3.ResBool == "true" {
			sessionMgr.updatePersistInfo(id, "true", "You have posted a comment", "false", "None", "", "seelast", false)
		} else {
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "an error has occurred while trying to post a comment", "", "seelast", false)
		}
		http.Redirect(res, req, "/user?id="+idParam+"&edit=false", http.StatusSeeOther)
		return
	}
	sessionMgr.updatePersistInfo(id, "false", "None", "false", "None", "", "seelast", false)
	tplUpdateUser.ExecuteTemplate(res, "updateuser.gohtml", dataInsert)
}
