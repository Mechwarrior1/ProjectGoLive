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
	reg, err := regexp.Compile(`[^a-zA-Z0-9!@#$%^&\s,.?]+`)
	if err != nil {
		log.Fatal(err)
	}
	processedString := reg.ReplaceAllString(string1, "")
	return processedString
}

// a function to access rest api and get all the courses
// returns all the courses available
// func getAllPost() ([]courseInfo, error) {

// 	url := baseURL
// 	response, err := http.Get(url)
// 	if err != nil {
// 		logger1.logTrace("ERROR", "The HTTPS request failed with error: "+err.Error())
// 		return []courseInfo{}, errors.New("https request failed with " + err.Error())
// 	} else {
// 		data1, _ := ioutil.ReadAll(response.Body)
// 		var newDataPacket dataPacket
// 		json.Unmarshal(data1, &newDataPacket)
// 		response.Body.Close()
// 		return newDataPacket.CourseInfoSlice, nil
// 	}
// }

// the function to access the rest api
// requires the method and datapacket
// returns any courseinfo and error
func tapAPI(httpMethod string, jsonData dataPacket, baseURL1 string) (*dataPacket, error) {
	// if baseURL1 == "" {
	// 	baseURL1 = baseURL + "/info"
	// }
	url := baseURL1
	jsonValue, _ := json.Marshal(jsonData)
	request, _ := http.NewRequest(httpMethod, url,
		bytes.NewBuffer(jsonValue))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	var newDataPacket dataPacket
	if err != nil {
		// logger1.logTrace("ERROR", "The HTTPS request failed with error: "+err.Error())
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

// a function for http handler, used to create any course.
// bring to a form for inputs
func createPost(res http.ResponseWriter, req *http.Request) {
	id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
	// if userPersistInfo1.Username == "" || userPersistInfo1.Username == "None" || userPersistInfo1.Username == "" {
	// 	sessionMgr.updatePersistInfo(id, "false", "None", "true", "Please kindly log in to post", "", "seelast", false)
	// 	http.Redirect(res, req, "/", http.StatusSeeOther)
	// 	return
	// }
	if req.Method == http.MethodPost {
		// get form values.
		postName := req.FormValue("PostName")
		postComment := req.FormValue("PostComment")
		postImg2 := req.FormValue("PostImg2")
		if strings.Contains(postImg2, "script") {
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
		newPost := make(map[string]string)
		newPost["Username"] = "admin" //userPersistInfo1.Username
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
			// key to access rest api
			Key:         key1(),
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
	tplCreatePost.ExecuteTemplate(res, "createpost.gohtml", userPersistInfo1)
}

// a function for http handler, used to create any course.
// bring to a form for inputs
func editPost(res http.ResponseWriter, req *http.Request) {
	id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
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

	dataInsert := struct {
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
		// get form values.
		postName := req.FormValue("PostName")
		postComment := req.FormValue("PostComment")
		postImg2 := req.FormValue("PostImg2")
		if strings.Contains(postImg2, "script") {
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
		newPost := make(map[string]string)
		newPost["Username"] = "admin" //userPersistInfo1.Username
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
			// key to access rest api
			Key:         key1(),
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
	tplEditPost.ExecuteTemplate(res, "editpost.gohtml", dataInsert)
}

func sortPost(dataArr []map[string]string, date1 string, cat1 string) ([]map[string]string, []int, error) {
	// fmt.Println("sort start:", date1, cat1)
	newSorted := []map[string]string{}
	sortArr := []int{}
	sortArrSim := []float32{}
	for i, map1 := range dataArr {
		timenow := time.Now().Unix()
		switch date1 { //calculate the cut off date
		case "7days":
			timenow = timenow - (7 * 24 * 60 * 60)
		case "30days":
			timenow = timenow - (30 * 24 * 60 * 60)
		}
		dateVal, _ := strconv.Atoi(map1["DatePosted"])
		map1["DatePosted"] = time.Unix(int64(dateVal), 0).Format("02-01-2006")

		if (timenow < int64(dateVal) || date1 == "All" || date1 == "") && (cat1 == map1["Cat"] || cat1 == "All" || cat1 == "") {
			// fmt.Println("added ", map1["Name"], timenow < int64(dateVal), cat1)
			sortArr = append(sortArr, i)
			simVal, _ := strconv.ParseFloat(map1["Similarity"], 32)
			sortArrSim = append(sortArrSim, float32(simVal))
			// fmt.Println("data0 :", simVal, map1["Name"], map1["Similarity"], sortArrSim)
		} else {
			// fmt.Println("skipped ", map1["Name"], timenow < int64(dateVal), cat1, map1["DatePosted"])
		}
	}
	_, sortArr2 := mergeSort(sortArrSim, sortArr)
	maxLen := len(sortArr)
	// fmt.Println(aa, sortArr, sortArr2)
	// for idx := 0; idx < maxLen; idx++ { //prints results in ascending order
	// 	newSorted = append(newSorted, dataArr[sortArr2[idx]])
	// 	fmt.Println("data :", dataArr[sortArr2[idx]]["Name"], dataArr[sortArr2[idx]]["Similarity"])
	// }
	for idx := maxLen - 1; idx >= 0; idx-- { //prints results in descending order
		newSorted = append(newSorted, dataArr[sortArr2[idx]])
		fmt.Println("data :", dataArr[sortArr2[idx]]["Name"], dataArr[sortArr2[idx]]["Similarity"])
	}
	return newSorted, sortArr2, nil
}

// a function for http handler, used to get any course.
// has a button to bring to a form to see the selected course
func seePostAll(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		postSearch := req.FormValue("PostSearch")
		postDate := req.FormValue("PostDate")
		postCat := req.FormValue("PostCat")
		http.Redirect(res, req, "/seepost?search1="+postSearch+"&date="+postDate+"&cat="+postCat, http.StatusSeeOther)
		return
	} else {
		// searchParams := mux.Vars(req)["search"]
		searchParam := req.URL.Query().Get("search1")
		dateParam := req.URL.Query().Get("date")
		catParam := req.URL.Query().Get("cat")
		// res.Write([]byte("<script>alert('Please login')</script>"))
		for _, params := range []string{searchParam, dateParam, catParam} {
			if replaceAllString(params) != params {
				http.Redirect(res, req, "/seepost/1?search1=&date=All&cat=All", http.StatusSeeOther)
				return
			}
		}
		_, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
		mapListing := make(map[string]string)
		mapListing["Name"] = searchParam // only for ItemListing
		fmt.Println("search: [", searchParam, dateParam, catParam, "] url:", req.URL.RequestURI())
		jsonData1 := dataPacket{
			// key to access rest api
			Key:         key1(),
			ErrorMsg:    "nil",
			InfoType:    "ItemListing",
			ResBool:     "false",
			RequestUser: userPersistInfo1.Username,
			DataInfo:    []map[string]string{mapListing},
		}

		dataPacket1, _ := tapAPI("GET", jsonData1, "https://127.0.0.1:5555/api/v0/allinfo")
		dataInfoSorted, _, _ := sortPost(dataPacket1.DataInfo, dateParam, catParam)
		// if err2 != nil {
		// 	sessionMgr.updatePersistInfo(id, "false", "None", "true1", "An error has occurred, please try again later", "", "seelast", false)
		// 	http.Redirect(res, req, "/", http.StatusSeeOther)
		// 	return
		// }
		dataInsert := struct {
			DataInfo        []map[string]string
			UserPersistInfo userPersistInfo
		}{
			dataInfoSorted,
			*userPersistInfo1,
		}
		// if err1 != nil {
		// 	///put logger
		// 	fmt.Println(err1)
		// 	dataInsert.DataInfo = []map[string]string{}
		// }

		tplSeePostAll.ExecuteTemplate(res, "seepostall.gohtml", dataInsert)
		return
	}
}

// a function for http handler, follow up from getCourse, zooms into the course
func getPostDetail(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	id, userPersistInfo1 := sessionMgr.getIdPersistInfo(res, req)
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
	if err1 != nil || dataPacket1.ResBool == "false" {
		sessionMgr.updatePersistInfo(id, "false", "None", "true", "The detail: "+params["courseid"]+" cannot be found, please try another course", "", "seelast", false)
		http.Redirect(res, req, "/seepost/1?search1=&date=All&cat=All", http.StatusSeeOther)
		return
	}

	jsonData1.InfoType = "CommentItem"
	dataPacket2, err2 := tapAPI(http.MethodGet, jsonData1, baseURL+"allinfo")
	if err2 != nil || dataPacket2.ResBool == "false" {
		sessionMgr.updatePersistInfo(id, "false", "None", "true", "The detail: "+params["courseid"]+" cannot be found, please try another course", "", "seelast", false)
		http.Redirect(res, req, "/seepost/1?search1=&date=All&cat=All", http.StatusSeeOther)
		return
	}
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
	if req.Method == http.MethodPost {
		username1 := userPersistInfo1.Username
		if username1 == "" || username1 == "None" {
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "you need to login to post a comment", "", "seelast", false)
			http.Redirect(res, req, "/getpost/"+params["id"], http.StatusSeeOther)
			return
		}
		postComment := req.FormValue("PostComment")
		mapComment := make(map[string]string)
		mapComment["CommentItem"] = postComment
		mapComment["Username"] = username1
		mapComment["ForItem"] = params["id"]
		mapComment["Date"] = time.Now().Format("02-01-2006")
		// fmt.Println(mapComment)
		jsonData1.DataInfo = []map[string]string{mapComment}
		dataPacket3, err3 := tapAPI(http.MethodPost, jsonData1, baseURL+"db/info")
		if err3 == nil && dataPacket3.ResBool == "true" {
			sessionMgr.updatePersistInfo(id, "true", "You have posted a comment", "false", "None", "", "seelast", false)
		} else {
			sessionMgr.updatePersistInfo(id, "false", "None", "true", "an error has occurred while trying to post a comment", "", "seelast", false)
		}
		http.Redirect(res, req, "/getpost/"+params["id"], http.StatusSeeOther)
		return
	}
	tplGetPostDetail.ExecuteTemplate(res, "getpostdetail.gohtml", dataInsert)
}

// a function for http handler, used to update any course.
// has a button to bring to a form to update
// func updateCourse(res http.ResponseWriter, req *http.Request) {
// 	id, persistInfo1 := sessionMgr.getIdPersistInfo(res, req)
// 	allCourse, err := getAllCourse()
// 	if err != nil {
// 		sessionMgr.updatePersistInfo(id, "false", "None", "true1", "An error seems to have occured: "+err.Error())
// 		http.Redirect(res, req, "/", http.StatusSeeOther)
// 		return
// 	}
// 	dataInsert := struct {
// 		CourseInfo  []courseInfo
// 		PersistInfo persistInfo
// 	}{
// 		allCourse,
// 		persistInfo1,
// 	}
// 	tplUpdateCourse.ExecuteTemplate(res, "updatecourse.gohtml", dataInsert)
// }

// a function for http handler, follow up from updateCourse, has a form to update the item
// func updateCourseForm(res http.ResponseWriter, req *http.Request) {
// 	params := mux.Vars(req)
// 	id, persistInfo1 := sessionMgr.getIdPersistInfo(res, req)
// 	courseInfo1 := courseInfo{
// 		"",
// 		params["courseid"],
// 		"",
// 		"",
// 	}
// 	jsonData1 := dataPacket{Key: key1(),
// 		ErrorMsg:        "nil",
// 		CourseInfoSlice: []courseInfo{courseInfo1},
// 	}
// 	allCourse, err2 := tapAPI(http.MethodGet, jsonData1, "")
// 	if err2 != nil {
// 		sessionMgr.updatePersistInfo(id, "false", "None", "true", "The course: "+params["courseid"]+" cannot be found, please try another course")
// 		http.Redirect(res, req, "/updatecourse", http.StatusSeeOther)
// 		return
// 	}
// 	dataInsert := struct {
// 		CourseInfo  courseInfo
// 		PersistInfo persistInfo
// 	}{
// 		allCourse[0],
// 		persistInfo1,
// 	}
// 	if req.Method == http.MethodPost {
// 		// get form values.
// 		courseName := req.FormValue("CourseName")
// 		courseTitle := req.FormValue("CourseTitle")
// 		courseDesc := req.FormValue("CourseDesc")
// 		for _, string1 := range []string{courseName, courseTitle} {
// 			if replaceAllString(string1) != string1 {
// 				sessionMgr.updatePersistInfo(id, "false", "None", "true", "Please fill in the form correctly, without special characters for the name and title")
// 				http.Redirect(res, req, "/updatecourse/"+params["courseid"], http.StatusSeeOther)
// 				return
// 			}
// 		}
// 		// create packet to send
// 		courseInfo1 = courseInfo{
// 			allCourse[0].ID,
// 			courseName,
// 			courseTitle,
// 			courseDesc,
// 		}
// 		jsonData1.CourseInfoSlice = []courseInfo{courseInfo1}
// 		_, err3 := tapAPI(http.MethodPut, jsonData1, "")
// 		if err3 != nil {
// 			sessionMgr.updatePersistInfo(id, "false", "None", "true", err3.Error())
// 			http.Redirect(res, req, "/updatecourse", http.StatusSeeOther)
// 			return
// 		}
// 		logger1.logTrace("TRACE", "Updated course: '"+courseName+"', title: '"+courseTitle+"', desc: '"+courseDesc+"'")
// 		sessionMgr.updatePersistInfo(id, "true", "You have updated course: '"+courseName+"', title: '"+courseTitle+"', desc: '"+courseDesc+"'", "false", "None")
// 		http.Redirect(res, req, "/", http.StatusSeeOther)
// 		return
// 	}
// 	tplUpdateCourseForm.ExecuteTemplate(res, "updatecourseform.gohtml", dataInsert)
// }

// http://myhost/country?code=DE
// you'd do:
// code := r.URL.Query().Get("code")
// <h4>1) <a href="/createcourse">Create a New Course</a></h4>

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
		http.Redirect(res, req, "/seepost?search1="+postSearch+"&date=All&cat="+postCat, http.StatusSeeOther)
		return
	}
	tplIndex.ExecuteTemplate(res, "index.gohtml", persistInfo1)
}

// a function for http handler, used to delete any course.
// func deleteCourse(res http.ResponseWriter, req *http.Request) {
// 	id, persistInfo1 := sessionMgr.getIdPersistInfo(res, req)
// 	allCourse, err := getAllCourse()
// 	if err != nil {
// 		sessionMgr.updatePersistInfo(id, "false", "None", "true1", "An error seems to have occured: "+err.Error())
// 		http.Redirect(res, req, "/", http.StatusSeeOther)
// 		return
// 	}
// 	dataInsert := struct {
// 		CourseInfo  []courseInfo
// 		PersistInfo persistInfo
// 	}{
// 		allCourse,
// 		persistInfo1,
// 	}
// 	if req.Method == http.MethodPost {
// 		// get chosen item
// 		req.ParseForm()
// 		deleteCourse1 := req.Form["deleteCourse"]
// 		// create packet to send
// 		courseInfo1 := courseInfo{
// 			"",
// 			deleteCourse1[0],
// 			"",
// 			"",
// 		}
// 		jsonData1 := dataPacket{Key: key1(),
// 			ErrorMsg:        "nil",
// 			CourseInfoSlice: []courseInfo{courseInfo1},
// 		}
// 		_, err1 := tapAPI(http.MethodDelete, jsonData1, "")
// 		if err1 != nil {
// 			sessionMgr.updatePersistInfo(id, "false", "None", "true1", "An error seems to have occured: "+err1.Error())
// 			http.Redirect(res, req, "/deletecourse", http.StatusSeeOther)
// 			return
// 		}
// 		logger1.logTrace("TRACE", "Deleted course: '"+deleteCourse1[0])
// 		sessionMgr.updatePersistInfo(id, "true", "You have deleted the course: "+deleteCourse1[0], "false", "None")
// 		http.Redirect(res, req, "/deletecourse", http.StatusSeeOther)
// 		return
// 	}
// 	tplDeleteCourse.ExecuteTemplate(res, "deletecourse.gohtml", dataInsert)
// }

// // a function for http handler, used for /shutdown, it shutdown the server, close the logger file and send the command to api service as well.
// func shutdown(res http.ResponseWriter, req *http.Request) {
// 	id, persistInfo1 := sessionMgr.getIdPersistInfo(res, req)
// 	if req.Method == http.MethodPost {
// 		shutdownVal := req.FormValue("shutdownVal")
// 		if shutdownVal == "shutdown" {
// 			func() {
// 				// create packet to send
// 				courseInfo1 := courseInfo{
// 					"",
// 					"shutdown",
// 					"",
// 					"",
// 				}
// 				jsonData1 := dataPacket{Key: key1(),
// 					ErrorMsg:        "nil",
// 					CourseInfoSlice: []courseInfo{courseInfo1},
// 				}
// 				_, err2 := tapAPI(http.MethodGet, jsonData1, "https://localhost:5555/api/v0/shutdown")
// 				if err2 != nil {
// 					logger1.logTrace("TRACE", "Server attempted to shutdown, error: "+err2.Error())
// 					sessionMgr.updatePersistInfo(id, "false", "None", "true", "An error occurred: "+err2.Error())
// 					http.Redirect(res, req, "/shutdown", http.StatusSeeOther)
// 				} else {
// 					logger1.logTrace("TRACE", "Server shutting down")
// 					sessionMgr.updatePersistInfo(id, "true", "The server is shutting down...", "false", "None")
// 					http.Redirect(res, req, "/", http.StatusSeeOther)
// 					logger1.logTrace("close_goRoutine", "") // close file.
// 					go func() {
// 						time.Sleep(1 * time.Second)
// 						if err := s.Shutdown(context.Background()); err != nil {
// 							log.Fatal(err)
// 						}
// 					}()
// 				}
// 			}()
// 		} else {
// 			sessionMgr.updatePersistInfo(id, "false", "None", "true", "Please type in shutdown correctly")
// 			http.Redirect(res, req, "/shutdown", http.StatusSeeOther)
// 			return
// 		}
// 	}
// 	tplShutdown.ExecuteTemplate(res, "shutdown.gohtml", persistInfo1)
// }
