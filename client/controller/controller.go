// Runs the http server to manage movies.
package controller

import (
	"bytes"
	"client/encrypt"
	"client/jwtsession"
	"client/session"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type (
	dataPacket struct {
		// key to access rest api
		Key         string      `json:"Key"`
		ErrorMsg    string      `json:"ErrorMsg"`
		InfoType    string      `json:"InfoType"` // 5 types: userSecret, userInfo, itemListing, commentUser, commentItem
		ResBool     string      `json:"ResBool"`
		RequestUser string      `json:"RequestUser"`
		DataInfo    interface{} `json:"DataInfo"`
	}

	SearchSession struct {
		DateCreated int64
		IdArr       []interface{}
	}

	DataPacketSimple struct {
		Msg     string `json:"Msg"`
		ResBool string `json:"ResBool"`
	}

	Template struct {
		templates *template.Template
	}

	HTTPClient interface {
		Do(req *http.Request) (*http.Response, error)
	}
)

const (
	// address for rest api
	baseURL = "https://localhost:5555/api/v0/"
)

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// removes all special characters from a string
func replaceAllString(string1 string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9!@#$%^&\s,.?_]+`)

	processedString := reg.ReplaceAllString(string1, "")
	return processedString
}

// helps to check the jwt token, see if user is logged in
// takes in echo.context, jwtwrapper, sessionmgr and checkUsername flag
// returns jwt claims and error if any
func checkLoggedIn(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session, checkUsername bool) (*jwtsession.JwtClaim, *jwtsession.JwtContext, error) {
	jwtClaim, err := sessionMgr.GetCookieJwt(c, jwtWrapper)
	jwtContext := &jwtClaim.Context

	//checks if username is not empty, if the flag to check is triggered
	if jwtContext.Username == "" && checkUsername {
		session.UpdateJwt("error", "You have to be logged in to use this service", jwtContext, c, jwtWrapper)
		return jwtClaim, jwtContext, errors.New("you have to be logged in to use this service")
	}

	if err != nil {
		session.UpdateJwt("error", "An error has occurred", jwtContext, c, jwtWrapper)
		fmt.Println("jwt error: ", err.Error())
		return jwtClaim, jwtContext, errors.New("jwt error")
	}
	return jwtClaim, jwtContext, err
}

// the function to access the rest api
// requires the method and datapacket
// returns any courseinfo and error
func tapApi(httpMethod string, jsonData interface{}, url string, sessionMgr *session.Session) (*map[string]interface{}, error) {
	url = baseURL + url
	var request *http.Request
	if jsonData != nil {
		jsonValue, _ := json.Marshal(jsonData)
		jsonValueMarshal := bytes.NewBuffer(jsonValue)
		request, _ = http.NewRequest(httpMethod, url, jsonValueMarshal)
	} else {
		request, _ = http.NewRequest(httpMethod, url, nil)
	}

	request.Header.Set("Content-Type", "application/json")
	// client := &http.Client{}
	response, err := sessionMgr.Client.Do(request)
	mapInterface := make(map[string]interface{})
	if err != nil {
		fmt.Println("tapapi failed with error:", err.Error())
		return &mapInterface, errors.New("https request failed with " + err.Error())

	} else {

		data1, err := ioutil.ReadAll(response.Body) //

		if err != nil {
			return &mapInterface, errors.New("ioutil failed to read, error: " + err.Error())
		}

		json.Unmarshal(data1, &mapInterface)
		response.Body.Close()

		if mapInterface["ErrorMsg"] != "nil" {
			return &mapInterface, errors.New("error")
		}

		// if data, ok := mapInterface["DataInfo"]; ok && len(data.([]interface{}) == 1{
		// 	dataTemp := data.([]interface{})[0].(map[string]interface{})
		// 	mapInterface["DataInfo"] = dataTemp
		// }

		return &mapInterface, nil
	}
}

// takes inputs from forms and send them to reset api
func CreatePost_POST(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, false)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// get form values and check
	form, _ := c.FormParams()
	if len(form["PostName"]) == 0 {
		session.UpdateJwt("error", "an error has occurred", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/")
	}

	if strings.Contains(form["PostImg2"][0], "script") {
		session.UpdateJwt("error", "Please fill in the form correctly, without special characters for the name and title", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// take inputs and put into map for api/server
	newPost := make(map[string]interface{})

	newPost["Name"] = form["PostName"][0]
	newPost["CommentItem"] = form["PostComment"][0]
	newPost["ConditionItem"] = form["PostCondition"][0]
	newPost["Cat"] = form["PostCat"][0]
	newPost["ContactMeetInfo"] = form["PostContactMeetInfo"][0]

	// check if form values contains any special characters, if so refreshes the page
	for _, string1 := range newPost {
		if replaceAllString(string1.(string)) != string1 {
			session.UpdateJwt("error", "Please fill in the form correctly, without special characters for the name and title", jwtContext, c, jwtWrapper)
			return c.Redirect(http.StatusSeeOther, "/")
		}
	}

	newPost["ImageLink"] = form["PostImg2"][0]
	newPost["Username"] = jwtContext.Username
	timenow := time.Now().Unix()
	newPost["DatePosted"] = strconv.Itoa(int(timenow))
	newPost["Completion"] = "false"

	jsonData1 := dataPacket{
		Key:         sessionMgr.ApiKey, // key to access rest api
		ErrorMsg:    "nil",
		InfoType:    "ItemListing",
		ResBool:     "false",
		RequestUser: jwtContext.Username,
		DataInfo:    []map[string]interface{}{newPost},
	}

	// communicate with api, with json payload
	_, err5 := tapApi("POST", jsonData1, "db/info", sessionMgr)

	if err5 != nil {

		session.UpdateJwt("error", "An error has occurred, please try again later", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/createpost")
	} else {
		// logger1.logTrace("TRACE", "Created item: '"+postName+"', by user: '"+jwtContext.Username+""+"' desc: '"+postComment+"'")

		session.UpdateJwt("ok", "You have created item: '"+newPost["Name"].(string)+"'", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusFound, "/createpost")
	}

	session.UpdateJwt("", "", jwtContext, c, jwtWrapper)
	return c.Render(http.StatusOK, "createpost.gohtml", jwtContext)
}

//renders the go html for get request, see post
func CreatePost_GET(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, true)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}
	session.UpdateJwt("", "", jwtContext, c, jwtWrapper)
	return c.Render(http.StatusOK, "createpost.gohtml", jwtContext)
}

// takes inputs from forms and sends it to rest api
func EditPost_POST(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {

	jwtClaim, err := sessionMgr.GetCookieJwt(c, jwtWrapper)
	jwtContext := &jwtClaim.Context

	if jwtContext.Username == "" || err != nil {
		session.UpdateJwt("error", "You have to be logged in to use this service", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// get form values and check
	form, err := c.FormParams()

	if strings.Contains(form["PostImg2"][0], "script") || err != nil {
		session.UpdateJwt("error", "Please fill in the form correctly, without special characters for the name and title", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/")
	}

	//put inputs into map and push it into the api
	newPost := make(map[string]interface{})
	newPost["Name"] = form["PostName"][0]
	newPost["CommentItem"] = form["PostComment"][0]
	newPost["ConditionItem"] = form["PostCondition"][0]
	newPost["Cat"] = form["PostCat"][0]
	newPost["ContactMeetInfo"] = form["PostContactMeetInfo"][0]

	// check if form values contains any special characters, if so refreshes the page
	for _, string1 := range newPost {
		if replaceAllString(string1.(string)) != string1 {
			session.UpdateJwt("error", "Please fill in the form correctly, without special characters for the name and title", jwtContext, c, jwtWrapper)
			return c.Redirect(http.StatusSeeOther, "/")
		}
	}

	newPost["ImageLink"] = form["PostImg2"][0]
	newPost["ID"] = c.Param("id")
	newPost["Username"] = jwtContext.Username

	jsonData1 := dataPacket{
		// key to access rest api
		Key:         sessionMgr.ApiKey,
		ErrorMsg:    "",
		InfoType:    "ItemListing",
		ResBool:     "",
		RequestUser: jwtContext.Username,
		DataInfo:    []map[string]interface{}{newPost},
	}

	_, err5 := tapApi(http.MethodPut, jsonData1, "db/info", sessionMgr) //communicate with api
	// if error feedback to user
	if err5 != nil {
		session.UpdateJwt("error", "An error has occurred, please try again later", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/editpost/"+c.Param("id"))
	} else {

		// logger1.logTrace("TRACE", "Created item: '"+postName+"', by user: '"+jwtContext.Username+""+"' desc: '"+postComment+"'")
		session.UpdateJwt("ok", "You have edited item: '"+newPost["Name"].(string)+"'", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusFound, "/editpost/"+c.Param("id"))
	}

	session.UpdateJwt("", "", jwtContext, c, jwtWrapper)
	return c.Redirect(http.StatusSeeOther, "/getpost/"+c.Param("id"))
}

// query rest api for information and populate the rendered html form
func EditPost_GET(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, true)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	//params from url are used in rest api query
	q := make(url.Values)
	q.Set("id", c.Param("id"))
	q.Set("db", "ItemListing")

	//query rest api

	fmt.Println("db/info?id="+c.Param("id")+"&db=ItemListing", "db/info?"+q.Encode())
	dataPacket1, err1 := tapApi(http.MethodGet, "", "db/info?id="+c.Param("id")+"&db=ItemListing", sessionMgr)
	fmt.Println(dataPacket1, err1)

	//check returned information, redirects if error or information is nil
	if err1 != nil || len((*dataPacket1)["DataInfo"].([]interface{})) == 0 {
		session.UpdateJwt("error", "An error has occurred, please try again later", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// checks if the user editing is the owner of the post
	// redirects if not owner
	if jwtContext.Username != (*dataPacket1)["DataInfo"].([]interface{})[0].(map[string]interface{})["Username"].(string) {
		session.UpdateJwt("error", "you cannot edit a post by others", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/")
	}

	dataInsert := struct { //struct info for go template
		DataInfo        interface{}
		UserPersistInfo jwtsession.JwtContext
	}{
		(*dataPacket1)["DataInfo"].([]interface{})[0],
		*jwtContext,
	}

	session.UpdateJwt("", "", jwtContext, c, jwtWrapper)
	return c.Render(http.StatusOK, "editpost.gohtml", dataInsert)
}

// func sorts the incoming dataInfo by similarity (each map entry has "similarity")
// based on how similar it is to the searched term
func SortPost(dataArr []interface{}, date1 string, cat1 string, sort1 string) ([]interface{}, []int) {
	// fmt.Println("sort start:", date1, cat1)
	newSorted := []interface{}{}
	sortArr := []int{}
	sortArrSim := []float64{}

	for i, mapData := range dataArr {
		map1 := mapData.(map[string]interface{})
		timenow := time.Now().Unix()
		switch date1 { //calculate the cut off date,
		case "7days":
			timenow = timenow - (7 * 24 * 60 * 60)
		case "30days":
			timenow = timenow - (30 * 24 * 60 * 60)
		default:
			timenow = timenow - (180 * 24 * 60 * 60)
		}

		dateVal, _ := strconv.Atoi(map1["DatePosted"].(string))
		map1["DatePosted"] = time.Unix(int64(dateVal), 0).Format("02-01-2006")

		// adds index of map into array if the map meets the criteria, before sorting it
		if (timenow < int64(dateVal) || date1 == "All" || date1 == "") && (cat1 == map1["Cat"] || cat1 == "All" || cat1 == "") {
			sortArr = append(sortArr, i)
			// simVal, _ := strconv.ParseFloat(map1["Similarity"], 32)
			sortArrSim = append(sortArrSim, map1["Similarity"].(float64))
		}

	}
	//sorts the similarity and provides an index
	_, sortArr2 := encrypt.MergeSort(sortArrSim, sortArr)
	maxLen := len(sortArr)

	if sort1 == "asc" {
		for idx := 0; idx < maxLen; idx++ { //sorts results in ascending order
			newSorted = append(newSorted, dataArr[sortArr2[idx]])
			// fmt.Println("data :", dataArr[sortArr2[idx]].(map[string]interface{})["ID"], ", ", dataArr[sortArr2[idx]].(map[string]interface{})["Similarity"])
		}

	} else {
		for idx := maxLen - 1; idx >= 0; idx-- { //sorts results in descending order
			newSorted = append(newSorted, dataArr[sortArr2[idx]])

		}
	}

	return newSorted, sortArr2
}

// takes search parameters and puts it into url to be redirected into
func SeePostAll_POST(c echo.Context) error {
	form, _ := c.FormParams()
	postSearch := form["PostSearch"][0]
	postDate := form["PostDate"][0]
	postCat := form["PostCat"][0]
	postSort := form["PostSort"][0]
	return c.Redirect(http.StatusSeeOther, "/seepost?search="+postSearch+"&date="+postDate+"&cat="+postCat+"&sort="+postSort)
}

//
func SeePostAll_GET(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session, searchSession map[string]SearchSession) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, false)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// adjust this variable for entries per page
	entriesPerPage := 5

	searchParam := c.QueryParam("search")
	dateParam := c.QueryParam("date")
	catParam := c.QueryParam("cat")
	sortParam := c.QueryParam("sort")
	page := c.QueryParam("pg")
	userParam := c.QueryParam("user")
	searchSessionId := c.QueryParam("sesid")

	// checks the search parameters for special characters, normally it is auto generated, but can be edited on search bar
	// redirects if there is an error in the search parameters
	for _, params := range []string{searchParam, dateParam, catParam, sortParam, page} {
		if replaceAllString(params) != params {

			session.UpdateJwt("error", "please try searching without any special characters", jwtContext, c, jwtWrapper)
			return c.Redirect(http.StatusSeeOther, "/seepost/1?search=&date=All&cat=All&sort=desc")
		}
	}

	searchSessionStruct, ok := searchSession[searchSessionId]

	//create new session if old is not found
	if searchSessionId == "" || !ok {

		//encode variables into url for api
		q := make(url.Values)
		q.Set("name", searchParam)
		q.Set("cat", catParam)
		q.Set("date", dateParam)
		q.Set("filter", userParam)
		fmt.Println(searchParam, catParam, dateParam, userParam)
		dataPacket1, err1 := tapApi("GET", "", "index?"+q.Encode(), sessionMgr)

		if err1 != nil {
			session.UpdateJwt("error", "An error has occurred, please try again later", jwtContext, c, jwtWrapper)
			return c.Redirect(http.StatusSeeOther, "/")
		}
		searchSessionId = uuid.NewV4().String()
		//add the array of ID into search session
		returnInfo := (*dataPacket1)["DataInfo"].([]interface{})
		searchSessionStruct = SearchSession{
			time.Now().Unix(),
			returnInfo}
		searchSession[searchSessionId] = searchSessionStruct

		//redirects to a new search with search session id, set page to 1
		session.UpdateJwt("", "", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusFound, "/seepost?sesid="+searchSessionId+"&pg=1")
	}

	dataInsert := struct {
		DataInfo         []interface{}
		UserPersistInfo  jwtsession.JwtContext
		PaginationString string
		PaginationBool   bool
	}{
		[]interface{}{},
		*jwtContext,
		"0",
		false,
	}
	if len(searchSessionStruct.IdArr) != 0 {
		//if session is found
		IdArr := searchSessionStruct.IdArr //the array of id associated with the session eg [000001, 0000002]
		lenArr := len(IdArr)               //len of the array
		pageInt, _ := strconv.Atoi(page)   //the current page the user is viewing

		//get max page for pagniation
		maxPage := (lenArr - (lenArr % entriesPerPage)) / entriesPerPage
		if (lenArr % entriesPerPage) > 0 {
			maxPage += 1
		}

		var IdArrDisplay []interface{}
		//limit max value of page
		if maxPage < pageInt {
			// there is no result beyond the max page
			// render a no result page
			session.UpdateJwt("", "", jwtContext, c, jwtWrapper)
			return c.Render(http.StatusOK, "seepostall.gohtml", dataInsert)

		} else if maxPage == pageInt {
			// eg entriesPerPage is 5, arr has 6 entries,
			// should output [5:6]
			IdArrDisplay = IdArr[(maxPage-1)*entriesPerPage:]
		} else {
			// eg entriesPerPage is 5, arr has 6 entries, pageInt is 1
			// should output [0:5]
			IdArrDisplay = IdArr[(pageInt-1)*entriesPerPage : pageInt*entriesPerPage]
		}

		jsonData1 := dataPacket{
			Key:         sessionMgr.ApiKey,
			ErrorMsg:    "nil",
			InfoType:    "ItemListing",
			ResBool:     "false",
			RequestUser: jwtContext.Username,
			DataInfo:    []interface{}{IdArrDisplay},
		}

		dataPacket1, err1 := tapApi("GET", jsonData1, "listing", sessionMgr)
		// fmt.Println((*dataPacket1)["DataInfo"].([]interface{})[0])

		if err1 != nil || (*dataPacket1)["DataInfo"] == nil {
			session.UpdateJwt("error", "An error has occurred, please try again later (no result returned)", jwtContext, c, jwtWrapper)
			return c.Redirect(http.StatusSeeOther, "/")
		}

		// data required by the go template
		dataInsert.PaginationBool = true
		dataInsert.PaginationString = pagination(pageInt, maxPage, searchSessionId)
		dataInsert.DataInfo = (*dataPacket1)["DataInfo"].([]interface{})
	}
	// res.Write([]byte("<script>alert('Please login')</script>"))

	session.UpdateJwt("", "", jwtContext, c, jwtWrapper)
	return c.Render(http.StatusOK, "seepostall.gohtml", dataInsert)
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func pagination(pageInt int, maxPage int, searchSessionId string) string {

	HTMLString := ""
	if maxPage < 8 {
		for i := 1; i <= maxPage; i++ {
			i2 := strconv.Itoa(i)
			if i == pageInt {
				HTMLString += "<li class='page-item active'><a class='page-link no-border' href=#>" + i2 + "</a></li>"
			} else {
				HTMLString += "<li class='page-item'><a class='page-link no-border' href='seepost?sesid=" + searchSessionId + "&pg=" + i2 + "'>" + i2 + "</a></li>"
			}

		}
	} else { //renders first page, current page +-2 and last page
		if pageInt > 3 {
			HTMLString += "<li class='page-item'><a class='page-link no-border' href='seepost?sesid=" + searchSessionId + "&pg=1'>1</a></li>"
		}
		for i := max(pageInt-2, 1); i <= min(maxPage, pageInt+2); i++ {
			i2 := strconv.Itoa(i)
			if i == pageInt {
				HTMLString += "<li class='page-item active'><a class='page-link no-border' href=#>" + i2 + "</a></li> ..."
			} else {
				HTMLString += "<li class='page-item'><a class='page-link no-border' href='seepost?sesid=" + searchSessionId + "&pg=" + i2 + "'>" + i2 + "</a></li>"
			}

		}
		if pageInt < 3 {
			HTMLString += "... <li class='page-item'><a class='page-link no-border' href='seepost?sesid=" + searchSessionId + "&pg=" + strconv.Itoa(maxPage) + "'>" + strconv.Itoa(maxPage) + "</a></li>"
		}

	}

	return HTMLString
}

// handler function follow up from getCourse, zooms into the course
func GetPostDetail_GET(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, false)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	dataPacket1, err1 := tapApi(http.MethodGet, "", "db/info?id="+c.Param("id")+"&db=ItemListing", sessionMgr)
	fmt.Println(err1, (*dataPacket1)["ResBool"], (*dataPacket1)["DataInfo"])

	if err1 != nil || (*dataPacket1)["ResBool"] == "false" { // || len((*dataPacket1)["DataInfo"].([]interface{})) == 0
		//if post id does not exist return to search page

		session.UpdateJwt("error", "The detail: "+c.Param("id")+" cannot be found, please try another course", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/seepost/1?search=&date=All&cat=All&sort=desc")
	}

	// request for Comments for the post, sending the post id to api, if id cannot be found, redirect

	dataPacket2, _ := tapApi(http.MethodGet, "", "comment/"+c.Param("id"), sessionMgr)

	// send data of post and its comments to the template for rendering
	postData := (*dataPacket1)["DataInfo"].([]interface{})[0].(map[string]interface{})
	dateVal, _ := strconv.Atoi(postData["DatePosted"].(string))
	postData["DatePosted"] = time.Unix(int64(dateVal), 0).Format("02-01-2006")

	dataInsert := struct {
		PostInfo        interface{}
		PostCommentInfo []interface{}
		UserPersistInfo jwtsession.JwtContext
		Owner           bool
	}{
		postData,
		(*dataPacket2)["DataInfo"].([]interface{}),
		*jwtContext,
		postData["Username"] == jwtContext.Username,
	}
	session.UpdateJwt("", "", jwtContext, c, jwtWrapper)
	return c.Render(http.StatusOK, "getpostdetail.gohtml", dataInsert)
}

// post request for adding a new comment for the post
func GetPostDetail_POST(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, true)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// prepare payload to api
	form, _ := c.FormParams()
	postComment := form["PostComment"][0]
	mapComment := make(map[string]interface{})
	mapComment["CommentItem"] = postComment
	mapComment["Username"] = jwtContext.Username

	mapComment["ForItem"] = c.Param("id")
	mapComment["Date"] = time.Now().Format("02-01-2006")

	jsonData1 := dataPacket{
		// key to access rest api
		Key:         sessionMgr.ApiKey,
		ErrorMsg:    "nil",
		InfoType:    "CommentItem",
		ResBool:     "false",
		RequestUser: jwtContext.Username,
		DataInfo:    []map[string]interface{}{mapComment},
	}

	jsonData1.DataInfo = []map[string]interface{}{mapComment}
	dataPacket3, err3 := tapApi(http.MethodPost, jsonData1, "db/info", sessionMgr)

	// if response is an error
	if err3 == nil && (*dataPacket3)["ResBool"].(string) == "true" {
		session.UpdateJwt("ok", "You have posted a comment", jwtContext, c, jwtWrapper)
	} else {
		session.UpdateJwt("error", "an error has occurred while trying to post a comment", jwtContext, c, jwtWrapper)
	}

	c.Redirect(http.StatusSeeOther, "/getpost/"+c.Param("id"))

	return c.Render(http.StatusOK, "getpostdetail.gohtml", nil)
}

///* obsolete, user search transferred to main search function
// see the post for a particular user
// filters for their post/listing
// func SeePostUser(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
// 	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, false)
// 	if err != nil {
// 		return c.Redirect(http.StatusSeeOther, "/")
// 	}

// 	postUsername := c.Param("id") // get post id

// 	dataPacket1, err1 := tapApi("GET", "", "listing/?name=&filter="+postUsername, sessionMgr)
// 	// dataInfoSorted, _ := sortPost(dataPacket1.DataInfo, "All", "All", "desc")

// 	if err1 != nil || (*dataPacket1)["ErrorMsg"] == "false" {
// 		session.UpdateJwt("error", "An error has occurred, or user has no post", jwtContext, c, jwtWrapper)
// 		return c.Redirect(http.StatusSeeOther, "/"+c.Param("id"))
// 	}

// 	dataInsert := struct {
// 		DataInfo        []interface{}
// 		UserPersistInfo jwtsession.JwtContext
// 		PostUsername    string
// 	}{
// 		(*dataPacket1)["DataInfo"].([]interface{}),
// 		*jwtContext,
// 		postUsername,
// 	}

// 	session.UpdateJwt("", "", jwtContext, c, jwtWrapper)
// 	return c.Render(http.StatusOK, "seepostuser.gohtml", dataInsert)
// }

// changes the post to completed, and return the rest of the values to api
func PostComplete(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	// checks if user is logged in
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, false)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// payload to api
	jsonData1 := dataPacket{
		// key to access rest api
		Key:         sessionMgr.ApiKey,
		ErrorMsg:    "",
		InfoType:    "ItemListing",
		ResBool:     "",
		RequestUser: jwtContext.Username,
		DataInfo:    []map[string]interface{}{},
	}

	// send to api
	// if an error is returned, it means the current user is not the owner
	dataPacket1, err5 := tapApi("PUT", jsonData1, "db/completed/"+c.Param("id"), sessionMgr)

	if (*dataPacket1)["ErrorMsg"].(string) == "Not owner" {
		session.UpdateJwt("error", "you cannot edit a post by others", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/")
	}

	if err5 != nil {
		session.UpdateJwt("error", "An error has occurred, please try again later", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// logger1.logTrace("TRACE", "Updated Item status for '"+mapListing2["Name"].(string)+"' to completed ")
	session.UpdateJwt("ok", "'"+c.Param("id")+"' is tagged as completed", jwtContext, c, jwtWrapper)
	return c.Redirect(http.StatusSeeOther, "/")
}

// handler function, for the index page
func Index_GET(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, false)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	session.UpdateJwt("", "", jwtContext, c, jwtWrapper)
	return c.Render(http.StatusOK, "index.gohtml", jwtContext)
}

// handler function, for the index page
// when posting, takes form params and redirect to search page
func Index_POST(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	form, _ := c.FormParams()

	postSearch := form["search"][0]
	postCat := form["cat"][0]
	url := "/seepost?search=" + postSearch + "&date=All&cat=" + postCat + "&sort=desc"
	return c.Redirect(http.StatusSeeOther, url)

}

// handler function gives you the information on user
func GetUser_GET(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, false)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	idParam := c.QueryParam("id")
	editParam := c.QueryParam("edit")

	// reuqesting the information for the user, using the post id

	dataPacket1, err1 := tapApi(http.MethodGet, "", "db/info?id="+idParam+"&db=UserInfo", sessionMgr)
	fmt.Println(err1)
	// if error in fetching data
	if err1 != nil || (*dataPacket1)["ResBool"] == "false" || len((*dataPacket1)["DataInfo"].([]interface{})) == 0 {
		//if user id does not exist return to index page
		session.UpdateJwt("error", "The detail: "+idParam+" cannot be found, please try another course", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// send data of post and its comments to the template for rendering
	userData := (*dataPacket1)["DataInfo"].([]interface{})[0]
	dataInsert := struct {
		UserData        interface{}
		UserPersistInfo jwtsession.JwtContext
		Owner           bool
		Edit            bool
	}{
		userData,
		*jwtContext,
		userData.(map[string]interface{})["Username"] == jwtContext.Username,
		editParam == "true",
	}
	session.UpdateJwt("", "", jwtContext, c, jwtWrapper)
	return c.Render(http.StatusOK, "updateuser.gohtml", dataInsert)
}

// handler function gives you the information on user
func GetUser_POST(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, true)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	idParam := c.QueryParam("id")

	// put inputs into map , for the payload to api
	form, _ := c.FormParams()
	commentItem := form["CommentItem"][0]
	mapComment := make(map[string]interface{})
	mapComment["CommentItem"] = commentItem
	mapComment["ID"] = idParam

	jsonData1 := dataPacket{
		// key to access rest api
		Key:         sessionMgr.ApiKey,
		ErrorMsg:    "nil",
		InfoType:    "UserInfo",
		ResBool:     "false",
		RequestUser: jwtContext.Username,
		DataInfo:    []map[string]interface{}{mapComment},
	}

	dataPacket3, err3 := tapApi(http.MethodPut, jsonData1, "db/info", sessionMgr)

	// if error in posting a comment
	if err3 == nil && (*dataPacket3)["ResBool"].(string) == "true" {
		session.UpdateJwt("ok", "You have posted a comment", jwtContext, c, jwtWrapper)
	} else {
		session.UpdateJwt("error", "an error has occurred while trying to post a comment", jwtContext, c, jwtWrapper)
	}

	return c.Redirect(http.StatusSeeOther, "/user?id="+idParam+"&edit=false")

}

// a function that checks if username is taken
func CheckUsername(username string, sessionMgr *session.Session) bool { //u

	//query the api
	dataInfo1, _ := tapApi(http.MethodGet, nil, "username/"+username, sessionMgr)

	return (*dataInfo1)["ResBool"].(string) == "true"
}

// handler function for signing up new users
func Signup_GET(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, false)
	session.UpdateJwt("", "", jwtContext, c, jwtWrapper)

	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}
	return c.Render(http.StatusOK, "signup.gohtml", jwtContext)
}

// handler function for signing up new users, post request with form
// does some checks on the string
// checks if username is taken on server
// before sending the new user info to be registered
func Signup_POST(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, false)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// process form submission.
	// get form values.
	form, _ := c.FormParams()
	username := form["username"][0]
	username1 := replaceAllString(username)
	password := form["password"][0]

	// checks username for non standard characters
	if username != username1 || username == "" {
		session.UpdateJwt("error", "there was something wrong with your username, please only use alphanumeric characters", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/signup")
	}

	if username != "" {
		// check if username exist/ taken.
		if ok := CheckUsername(username, sessionMgr); ok {
			session.UpdateJwt("error", "Username already taken", jwtContext, c, jwtWrapper)
			return c.Redirect(http.StatusSeeOther, "/signup")
		}

		// encrypt password and prepare payload
		bPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
		currentTime := time.Now()
		lastLogin := currentTime.Format("06-01-2006 15:04 Monday")
		err5 := AddUser(username, string(bPassword), "", lastLogin, sessionMgr) // send user info to api

		if err5 != nil {
			session.UpdateJwt("error", "There was an error, please try again", jwtContext, c, jwtWrapper)
			return c.Redirect(http.StatusSeeOther, "/signup")
		}

		session.UpdateJwt("ok", "You have successfully signed up!", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	return c.Render(http.StatusOK, "signup.gohtml", jwtContext)
}

// sends user info to api to be added into mysql
func AddUser(username string, pwString string, commentItem string, lastLogin string, sessionMgr *session.Session) error {

	// prepare payload to api, to register the new user on server
	userSecret1 := make(map[string]interface{})
	userSecret1["Username"] = username
	userSecret1["Password"] = pwString
	userSecret1["CommentItem"] = commentItem
	userSecret1["IsAdmin"] = "false"

	currentTime := time.Now()
	userSecret1["DateJoin"] = currentTime.Format("02-01-2006 Monday")
	userSecret1["LastLogin"] = currentTime.Format("02-01-2006 Monday")

	jsonData1 := dataPacket{
		// key to access rest api
		Key:         sessionMgr.ApiKey,
		ErrorMsg:    "",
		InfoType:    "UserSecret",
		ResBool:     "",
		RequestUser: username,
		DataInfo:    []map[string]interface{}{userSecret1},
	}

	//send to api
	_, err1 := tapApi(http.MethodPost, jsonData1, "db/signup", sessionMgr)

	//check if it returned an error
	if err1 != nil {
		return err1
	}

	return err1
}

// handler function, checks if user is logged in
func Login_GET(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, false)
	session.UpdateJwt("", "", jwtContext, c, jwtWrapper)

	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}
	return c.Render(http.StatusOK, "login.gohtml", jwtContext)
}

// handler function, to login registered users
func Login_POST(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, false)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// process form submission.
	form, _ := c.FormParams()
	username := form["username"][0]
	password := form["password"][0]

	if username != replaceAllString(username) {
		session.UpdateJwt("error", "Please only use alphanumeric characters", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// check if user exist with username.
	// Matching of password entered.
	ok1, isAdmin, lastLogin := CheckPW(username, password, sessionMgr) /////////////////////////
	if !ok1 {
		// logger1.logTrace("TRACE", "someone tried to login to: "+username+", but the password is wrong")
		session.UpdateJwt("error", "Username and/or password do not match", jwtContext, c, jwtWrapper)
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// create session
	//success string, msg string, admin string, lastLogin string, username string, jwtContext *jwtsession.JwtContext, c echo.Context, jwtWrapper *jwtsession.JwtWrapper, session *Session)
	session.UpdateJwtLong("ok", "You have successfully logged in!", strconv.FormatBool(isAdmin), lastLogin, username, jwtContext, c, jwtWrapper, sessionMgr) // call to userinfo api for lastlogin
	c.Redirect(http.StatusSeeOther, "/")

	return c.Render(http.StatusOK, "login.gohtml", jwtContext)
}

// talks api, to check username and password
// api returns true or false
func CheckPW(username string, password string, sessionMgr *session.Session) (bool, bool, string) {

	//prepare payload to api, with user information to check
	userSecret1 := make(map[string]interface{})
	userSecret1["Username"] = username
	userSecret1["Password"] = password
	lastLogin := time.Now().Format("02-01-2006 15:04 Monday")
	userSecret1["LastLogin"] = lastLogin
	jsonData1 := dataPacket{
		// key to access rest api
		Key:         sessionMgr.ApiKey,
		ErrorMsg:    "",
		InfoType:    "UserSecret",
		ResBool:     "",
		RequestUser: "",
		DataInfo:    []map[string]interface{}{userSecret1},
	}

	//send payload to api
	dataInfo1, err1 := tapApi(http.MethodGet, jsonData1, "check", sessionMgr)

	if err1 != nil {
		return false, false, "error"
	}

	//maps interface into map[string]interface
	mapData := (*dataInfo1)["DataInfo"].([]interface{})[0].(map[string]interface{})

	fmt.Println("checkUser: ", err1, dataInfo1)
	return (*dataInfo1)["ResBool"].(string) == "true", mapData["IsAdmin"].(string) == "true", mapData["LastLogin"].(string)
}

// handler function to log out a logged in user
func Logout(c echo.Context, jwtWrapper *jwtsession.JwtWrapper, sessionMgr *session.Session) error {
	// checks if user is logged in
	_, jwtContext, err := checkLoggedIn(c, jwtWrapper, sessionMgr, true)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// new cookie without username for the logged out user
	signedToken, _, _ := jwtWrapper.GenerateToken("ok", "You have successfully logged out!", "false", "", "", jwtContext.Uuid)
	session.NewCookie(c, 15, signedToken)
	sessionMgr.DeleteSession(jwtContext.Username)

	return c.Redirect(http.StatusSeeOther, "/")
}
