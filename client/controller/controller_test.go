package controller

import (
	"bytes"
	"client/jwtsession"
	"client/session"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

var (
	// GetDoFunc fetches the mock client's `Do` func
	Client HTTPClient
)

type (
	// Custom type that allows setting the func that our Mock Do func will run instead
	MockDoFunc func(req *http.Request) (*http.Response, error)

	// MockClient is the mock client
	MockClient struct {
		MockDo MockDoFunc
	}
)

// Overriding what the Do function should "do" in our MockClient
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

// provides the wrapper required by handlers
func getDependency(address string, encode io.Reader) (*jwtsession.JwtWrapper, *session.Session, *httptest.ResponseRecorder, *http.Request, string, *echo.Echo, echo.Context) {
	jwtWrapper := &jwtsession.JwtWrapper{
		"key",
		"GoRecycle",
		10,
	}

	sessionMgr := &session.Session{
		MapSession: &map[string]session.SessionStruct{"username": session.SessionStruct{"uuid", 123}},
		ApiKey:     "key",
		Client:     &http.Client{},
	}

	generatedToken, _, _ := jwtWrapper.GenerateToken("success", "msg", "false", "lastlogin", "username", "uuid")

	//mock form values to function post
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, address, encode)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	req.AddCookie(&http.Cookie{
		Name:   "goRecycleCookie",
		Value:  generatedToken,
		MaxAge: 300,
	})
	e := echo.New()
	c := e.NewContext(req, rec)
	tem := &Template{
		templates: template.Must(template.ParseGlob("templates/*.gohtml")),
	}
	e.Renderer = tem

	return jwtWrapper, sessionMgr, rec, req, generatedToken, e, c
}

// the func that reads request from the perspective of the api server
func readJSONBody(r *http.Request) (map[string]interface{}, error) {
	// decode JSON body
	json_map := make(map[string]interface{})

	err1 := json.NewDecoder(r.Body).Decode(&json_map)
	if err1 == nil {

		if json_map["Key"] != "key" {
			return json_map, errors.New("incorrect api key supplied")
		}

		return json_map, nil
	}

	return json_map, errors.New("error while attempting to read body of request")
}

func TestCreatePost_POST(t *testing.T) {
	// create dependencies

	// mock client Do for handler
	// build our response JSON

	// create a new reader with that JSON

	client := &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {

			dataPacket1, _ := readJSONBody(req)
			dataPacket1 = dataPacket1["DataInfo"].([]interface{})[0].(map[string]interface{})

			mapResponse := make(map[string]string)
			assert.Equal(t, dataPacket1["Name"], "john")
			assert.Equal(t, dataPacket1["ImageLink"], "ImageLink.com")

			var statusCode int
			if assert.Equal(t, dataPacket1["ConditionItem"], "ConditionItem") {
				statusCode = 201
				mapResponse["ErrorMsg"] = "nil"
			} else {
				statusCode = 401
				mapResponse["ErrorMsg"] = "information not equal"
			}

			jsonResponse, _ := json.Marshal(mapResponse)
			r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

			return &http.Response{
				StatusCode: statusCode,
				Body:       r,
			}, nil
		},
	}

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "ImageLink.com")

	jwtWrapper, sessionMgr, rec, _, _, _, c := getDependency("/createpost", strings.NewReader(f.Encode()))
	sessionMgr.Client = client

	// fmt.Println(rec.Body)

	if assert.NoError(t, CreatePost_POST(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusFound, rec.Code)
	}
}

func TestCreatePost_POSTError(t *testing.T) {
	// mock client Do for handler
	// build our response JSON

	client := &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {

			dataPacket1, _ := readJSONBody(req)
			dataPacket1 = dataPacket1["DataInfo"].([]interface{})[0].(map[string]interface{})

			mapResponse := make(map[string]string)
			assert.Equal(t, dataPacket1["Name"], "john")
			assert.Equal(t, dataPacket1["ImageLink"], "ImageLink.com")

			var statusCode int
			if assert.Equal(t, dataPacket1["ConditionItem"], "ConditionItem") {
				statusCode = 201
				mapResponse["ErrorMsg"] = "nil"
			} else {
				statusCode = 401
				mapResponse["ErrorMsg"] = "information not equal"
			}

			jsonResponse, _ := json.Marshal(mapResponse)
			r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

			return &http.Response{
				StatusCode: statusCode,
				Body:       r,
			}, nil
		},
	}

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "<script>ImageLink.com")

	jwtWrapper, sessionMgr, rec, req, _, e, c := getDependency("/createpost", strings.NewReader(f.Encode()))
	sessionMgr.Client = client

	if assert.NoError(t, CreatePost_POST(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusSeeOther, rec.Code)

		//create new req to extract cookie
		req = &http.Request{Header: http.Header{"Cookie": rec.HeaderMap["Set-Cookie"]}}
		c = e.NewContext(req, rec)
		claims, _ := sessionMgr.GetCookieJwt(c, jwtWrapper)

		//checker jwtclaims
		assert.Equal(t, "error", claims.Context.Success)
		assert.Equal(t, "Please fill in the form correctly, without special characters for the name and title", claims.Context.Msg)
	}
}

func TestCreatePost_POSTError2(t *testing.T) {
	// mock client Do for handler
	// build our response JSON
	client := &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			mapResponse := make(map[string]string)
			mapResponse["ErrorMsg"] = "error"

			jsonResponse, _ := json.Marshal(mapResponse)
			r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

			return &http.Response{
				StatusCode: 404,
				Body:       r,
			}, nil
		},
	}

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "ImageLink.com")

	jwtWrapper, sessionMgr, rec, _, _, e, c := getDependency("/createpost", strings.NewReader(f.Encode()))
	sessionMgr.Client = client

	// fmt.Println(rec.Body)

	if assert.NoError(t, CreatePost_POST(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusSeeOther, rec.Code)

		//create new req to extract cookie
		req := &http.Request{Header: http.Header{"Cookie": rec.HeaderMap["Set-Cookie"]}}
		c = e.NewContext(req, rec)
		claims, _ := sessionMgr.GetCookieJwt(c, jwtWrapper)

		//checker jwtclaims
		assert.Equal(t, "error", claims.Context.Success)
		assert.Equal(t, "An error has occurred, please try again later", claims.Context.Msg)
	}
}

func TestCreatePost_GET(t *testing.T) {
	// create dependencies

	jwtWrapper, sessionMgr, rec, _, _, _, c := getDependency("/createpost", nil)

	if assert.NoError(t, CreatePost_GET(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestCreatePost_GETError(t *testing.T) {
	// create dependencies
	jwtWrapper, sessionMgr, rec, _, _, _, c := getDependency("/createpost", nil)

	//change
	sessionMgr.MapSession = &map[string]session.SessionStruct{"user1": session.SessionStruct{"uuid", 123}}

	if assert.NoError(t, CreatePost_GET(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusSeeOther, rec.Code)
	}
}

func TestEditPost_POST(t *testing.T) {
	// mock client Do for handler
	// build our response JSON
	client := &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {

			dataPacket1, _ := readJSONBody(req)
			dataPacket1 = dataPacket1["DataInfo"].([]interface{})[0].(map[string]interface{})

			mapResponse := make(map[string]string)
			assert.Equal(t, dataPacket1["Name"], "john")
			assert.Equal(t, dataPacket1["ImageLink"], "ImageLink.com")

			var statusCode int
			if assert.Equal(t, dataPacket1["ConditionItem"], "ConditionItem") {
				statusCode = 201
				mapResponse["ErrorMsg"] = "nil"
			} else {
				statusCode = 401
				mapResponse["ErrorMsg"] = "information not equal"
			}

			jsonResponse, _ := json.Marshal(mapResponse)
			r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

			return &http.Response{
				StatusCode: statusCode,
				Body:       r,
			}, nil
		},
	}

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "ImageLink.com")

	jwtWrapper, sessionMgr, rec, _, _, _, c := getDependency("/editpost", strings.NewReader(f.Encode()))
	sessionMgr.Client = client

	if assert.NoError(t, EditPost_POST(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusFound, rec.Code)
	}
}

func TestEditPost_POSTError(t *testing.T) {
	// mock client Do for handler
	// build our response JSON
	client := &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {

			dataPacket1, _ := readJSONBody(req)
			dataPacket1 = dataPacket1["DataInfo"].([]interface{})[0].(map[string]interface{})

			mapResponse := make(map[string]string)
			assert.Equal(t, dataPacket1["Name"], "john")
			assert.Equal(t, dataPacket1["ImageLink"], "ImageLink.com")

			var statusCode int
			if assert.Equal(t, dataPacket1["ConditionItem"], "ConditionItem") {
				statusCode = 201
				mapResponse["ErrorMsg"] = "nil"
			} else {
				statusCode = 401
				mapResponse["ErrorMsg"] = "information not equal"
			}

			jsonResponse, _ := json.Marshal(mapResponse)
			r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

			return &http.Response{
				StatusCode: statusCode,
				Body:       r,
			}, nil
		},
	}

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "<script>ImageLink.com")

	jwtWrapper, sessionMgr, rec, req, _, e, c := getDependency("/editpost", strings.NewReader(f.Encode()))
	sessionMgr.Client = client

	if assert.NoError(t, EditPost_POST(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusSeeOther, rec.Code)

		//create new req to extract cookie
		req = &http.Request{Header: http.Header{"Cookie": rec.HeaderMap["Set-Cookie"]}}
		c = e.NewContext(req, rec)
		claims, _ := sessionMgr.GetCookieJwt(c, jwtWrapper)

		//checker jwtclaims
		assert.Equal(t, "error", claims.Context.Success)
		assert.Equal(t, "Please fill in the form correctly, without special characters for the name and title", claims.Context.Msg)
	}
}

func TestEditPost_POSTError2(t *testing.T) {
	// mock client Do for handler
	// build our response JSON
	client := &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			mapResponse := make(map[string]string)
			mapResponse["ErrorMsg"] = "error"

			jsonResponse, _ := json.Marshal(mapResponse)
			r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

			return &http.Response{
				StatusCode: 404,
				Body:       r,
			}, nil
		},
	}

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "ImageLink.com")

	jwtWrapper, sessionMgr, rec, req, _, e, c := getDependency("/editpost", strings.NewReader(f.Encode()))
	sessionMgr.Client = client

	// fmt.Println(rec.Body)

	if assert.NoError(t, EditPost_POST(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusSeeOther, rec.Code)

		//create new req to extract cookie
		req = &http.Request{Header: http.Header{"Cookie": rec.HeaderMap["Set-Cookie"]}}
		c = e.NewContext(req, rec)
		claims, _ := sessionMgr.GetCookieJwt(c, jwtWrapper)

		//checker jwtclaims
		assert.Equal(t, "error", claims.Context.Success)
		assert.Equal(t, "An error has occurred, please try again later", claims.Context.Msg)
	}
}

// provide dummy data for mock api reply
func getDummy(db string) (newMap map[string]interface{}) {
	newMap = make(map[string]interface{})
	switch db {
	case "ItemListing":
		newMap["ID"] = "000001"
		newMap["Username"] = "username"
		newMap["Name"] = "username"
		newMap["ImageLink"] = "nil"
		newMap["DatePosted"] = "nil"
		newMap["CommentItem"] = "nil"
		newMap["ConditionItem"] = "nil"
		newMap["Cat"] = "nil"
		newMap["ContactMeetInfo"] = "nil"
		newMap["Completion"] = "nil"
	case "UserSecret":
		newMap["Username"] = "username"
		newMap["Password"] = "username"
		newMap["IsAdmin"] = "false"
		newMap["ID"] = "000001"
		newMap["CommentItem"] = "nil"
	case "UserInfo":
		newMap["ID"] = "000001"
		newMap["Username"] = "username"
		newMap["LastLogin"] = "20-7-2021"
		newMap["CommentItem"] = "nil"
		newMap["DateJoin"] = "20-7-2021"
	case "CommentItem":
		newMap["ID"] = "000001"
		newMap["Username"] = "username"
		newMap["ForItem"] = "000001"
		newMap["CommentItem"] = "nil"
		newMap["Date"] = "20-7-2021"
	}
	return
}

func TestEditPost_GET(t *testing.T) {
	// mock client Do for handler
	// build our response JSON

	client := &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			id := req.URL.Query().Get("id")
			db := req.URL.Query().Get("db")

			assert.Equal(t, "ItemListing", db)

			mapResponse := make(map[string]interface{})
			mapResponse["DataInfo"] = []interface{}{getDummy("ItemListing")}
			var statusCode int

			if assert.Equal(t, "000001", id) {
				statusCode = 201
				mapResponse["ErrorMsg"] = "nil"
			} else {
				statusCode = 401
				mapResponse["ErrorMsg"] = "assert not equal"
			}

			jsonResponse, _ := json.Marshal(mapResponse)
			r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

			return &http.Response{
				StatusCode: statusCode,
				Body:       r,
			}, nil
		},
	}

	jwtWrapper, sessionMgr, rec, _, _, _, c := getDependency("/editpost/000001", nil)
	sessionMgr.Client = client

	c.SetPath("/editpost/000001")
	c.SetParamNames("id")
	c.SetParamValues("000001")

	if assert.NoError(t, EditPost_GET(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

// test for error, different user
func TestEditPost_GETError(t *testing.T) {
	// mock client Do for handler
	// build our response JSON
	client := &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			id := req.URL.Query().Get("id")
			db := req.URL.Query().Get("db")

			assert.Equal(t, "ItemListing", db)

			mapResponse := make(map[string]interface{})
			dummyMap := getDummy("ItemListing")
			dummyMap["Username"] = "user1"
			mapResponse["DataInfo"] = []interface{}{dummyMap}
			var statusCode int

			if assert.Equal(t, "000001", id) {
				statusCode = 201
				mapResponse["ErrorMsg"] = "nil"
			} else {
				statusCode = 401
				mapResponse["ErrorMsg"] = "assert not equal"
			}

			jsonResponse, _ := json.Marshal(mapResponse)
			r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

			return &http.Response{
				StatusCode: statusCode,
				Body:       r,
			}, nil
		},
	}

	jwtWrapper, sessionMgr, rec, _, _, _, c := getDependency("/editpost/000001", nil)
	sessionMgr.Client = client

	c.SetPath("/editpost/000001")
	c.SetParamNames("id")
	c.SetParamValues("000001")

	if assert.NoError(t, EditPost_GET(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusSeeOther, rec.Code)
	}
}

// test sorting functions like ascending, 7days and similarity
func TestSortPostAsc(t *testing.T) {
	dummyInfoArray := []interface{}{}
	for i := 1; i < 30; i++ {
		dummyInfo := getDummy("ItemListing")
		dummyInfo["ID"] = i
		dummyInfo["DatePosted"] = strconv.FormatInt(time.Now().Unix()-int64(i*24*60*60/3), 10)
		dummyInfo["Similarity"] = float64(1 / float64(i)) // or (7-i)/7
		dummyInfoArray = append(dummyInfoArray, dummyInfo)
	}
	dummyInfoArraySorted, _ := SortPost(dummyInfoArray, "7days", "All", "asc")
	if assert.Equal(t, 20, len(dummyInfoArraySorted)) {
		assert.Equal(t, dummyInfoArraySorted[0].(map[string]interface{})["ID"], 20)
	}
}

// test other sorting functions like descending, 30days, similarity and category
func TestSortPostDesc(t *testing.T) {
	dummyInfoArray := []interface{}{}
	for i := 1; i < 100; i++ {
		dummyInfo := getDummy("ItemListing")
		if i%2 == 0 {
			dummyInfo["Cat"] = "Cat1"
		}
		dummyInfo["ID"] = i
		dummyInfo["DatePosted"] = strconv.FormatInt(time.Now().Unix()-int64(i*24*60*60), 10)
		dummyInfo["Similarity"] = float64(1 / float64(i)) // or (7-i)/7
		dummyInfoArray = append(dummyInfoArray, dummyInfo)
	}
	dummyInfoArraySorted, _ := SortPost(dummyInfoArray, "30days", "Cat1", "desc")
	if assert.Equal(t, 14, len(dummyInfoArraySorted)) {
		assert.Equal(t, dummyInfoArraySorted[0].(map[string]interface{})["ID"], 2)
		assert.Equal(t, dummyInfoArraySorted[0].(map[string]interface{})["Cat"], "Cat1")
	}
}

func TestIndex_GET(t *testing.T) {
	jwtWrapper, sessionMgr, rec, _, _, _, c := getDependency("/", nil)

	c.SetPath("/")

	if assert.NoError(t, Index_GET(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestSeePostAll_GETIndex(t *testing.T) {
	// mock client Do for handler
	// build our response JSON

	client := &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {

			//get query inputs
			itemName := req.URL.Query().Get("name")
			filterUsername := req.URL.Query().Get("filter")
			filterDate := req.URL.Query().Get("date")
			filterCat := req.URL.Query().Get("cat")

			//check inputs from client
			assert.Equal(t, "PET", itemName)
			assert.Equal(t, "", filterUsername)
			assert.Equal(t, "7days", filterDate)

			mapResponse := make(map[string]interface{})
			var statusCode int
			if assert.Equal(t, filterCat, "cat") {
				statusCode = 201
				mapResponse["ErrorMsg"] = "nil"
			} else {
				statusCode = 401
				mapResponse["ErrorMsg"] = "information not equal"
			}

			//response back to client

			mapResponse["DataInfo"] = []string{"000001", "000002", "000003"}
			mapResponse["ErrorMsg"] = "nil"

			jsonResponse, _ := json.Marshal(mapResponse)
			r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

			return &http.Response{
				StatusCode: statusCode,
				Body:       r,
			}, nil

			// } // as there are 2 api calls in handler
			// dataPacket1, _ := readJSONBody(req)
			// indArr := dataPacket1["DataInfo"].([]interface{})[0].([]interface{})

			// //check inputs for error
			// assert.Equal(t, "000001", indArr[0])
			// assert.Equal(t, "000002", indArr[1])
			// assert.Equal(t, "000003", indArr[2])

			// //response back to client
			// mapResponse := make(map[string]interface{})
			// newDataInfo := []interface{}{}
			// for _, ind := range indArr {
			// 	newData := getDummy("ItemListing")
			// 	newData["ID"] = ind
			// 	newDataInfo = append(newDataInfo, newData)
			// }

			// mapResponse["DataInfo"] = newDataInfo
			// mapResponse["ErrorMsg"] = "nil"

			// jsonResponse, _ := json.Marshal(mapResponse)
			// r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

			// return &http.Response{
			// 	StatusCode: 200,
			// 	Body:       r,
			// }, nil

		},
	}

	//dependencies
	q := make(url.Values)
	q.Set("search", "PET")
	q.Set("cat", "cat")
	q.Set("date", "7days")
	fmt.Println("/editpost?" + q.Encode())
	searchSession := make(map[string]SearchSession)
	jwtWrapper, sessionMgr, rec, _, _, _, c := getDependency("/editpost?"+q.Encode(), nil)
	sessionMgr.Client = client

	if assert.NoError(t, SeePostAll_GET(c, jwtWrapper, sessionMgr, searchSession)) {
		assert.Equal(t, http.StatusFound, rec.Code)
	}
}

// mock a search session, corresponding to the search requirement
// should return statusOK
func TestSeePostAll_GETRender(t *testing.T) {
	// mock client Do for handler
	// build our response JSON
	clientSwitch := false

	client := &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, clientSwitch, false) // assert client is not called twice

			if !clientSwitch {
				//get query inputs
				dataPacket1, _ := readJSONBody(req)
				indArr := dataPacket1["DataInfo"].([]interface{})[0].([]interface{})

				//check inputs for error
				assert.Equal(t, "000001", indArr[0])
				assert.Equal(t, "000002", indArr[1])
				assert.Equal(t, "000003", indArr[2])

				mapResponse := make(map[string]interface{})
				statusCode := 201
				mapResponse["ErrorMsg"] = "nil"

				//response back to client

				mapResponse["DataInfo"] = []map[string]interface{}{getDummy("ItemListing"), getDummy("ItemListing")}
				mapResponse["ErrorMsg"] = "nil"

				jsonResponse, _ := json.Marshal(mapResponse)
				r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

				clientSwitch = true
				return &http.Response{
					StatusCode: statusCode,
					Body:       r,
				}, nil
			}
			return &http.Response{
				StatusCode: 401,
				Body:       nil,
			}, errors.New("error, mockClient should not be called twice")
		},
	}

	//dependencies
	q := make(url.Values)
	q.Set("sesid", "123")
	q.Set("pg", "1")
	// fmt.Println("/editpost?" + q.Encode())
	searchSession := make(map[string]SearchSession)
	searchSession["123"] = SearchSession{123, []interface{}{"000001", "000002", "000003"}}
	jwtWrapper, sessionMgr, rec, _, _, _, c := getDependency("/editpost?"+q.Encode(), nil)
	sessionMgr.Client = client

	if assert.NoError(t, SeePostAll_GET(c, jwtWrapper, sessionMgr, searchSession)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		// assert.Equal(t, json_map["ResBool"], "true")

	}
}

func TestPagination(t *testing.T) {
	pagination1 := pagination(3, 5, "test")
	testResult := "<li class='page-item'><a class='page-link no-border' href='seepost?sesid=test&pg=1'>1</a></li><li class='page-item'><a class='page-link no-border' href='seepost?sesid=test&pg=2'>2</a></li><li class='page-item active'><a class='page-link no-border' href=#>3</a></li><li class='page-item'><a class='page-link no-border' href='seepost?sesid=test&pg=4'>4</a></li><li class='page-item'><a class='page-link no-border' href='seepost?sesid=test&pg=5'>5</a></li>"
	assert.Equal(t, pagination1, testResult)
}
