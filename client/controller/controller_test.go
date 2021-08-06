package controller

import (
	"bytes"
	"client/jwtsession"
	"client/session"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

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
func getWrapper() (*jwtsession.JwtWrapper, *session.Session) {
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
	return jwtWrapper, sessionMgr
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
	jwtWrapper, sessionMgr := getWrapper()

	generatedToken, _, _ := jwtWrapper.GenerateToken("success", "msg", "false", "lastlogin", "username", "uuid")

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

	sessionMgr.Client = client

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "ImageLink.com")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/createpost", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	req.AddCookie(&http.Cookie{
		Name:   "goRecycleCookie",
		Value:  generatedToken,
		MaxAge: 300,
		Path:   "/",
	})
	e := echo.New()
	c := e.NewContext(req, rec)

	// fmt.Println(rec.Body)

	if assert.NoError(t, CreatePost_POST(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusFound, rec.Code)
	}
}

func TestCreatePost_POSTError(t *testing.T) {
	// create dependencies
	jwtWrapper, sessionMgr := getWrapper()

	generatedToken, _, _ := jwtWrapper.GenerateToken("success", "msg", "false", "lastlogin", "username", "uuid")

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

	sessionMgr.Client = client

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "<script>ImageLink.com")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/createpost", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	req.AddCookie(&http.Cookie{
		Name:   "goRecycleCookie",
		Value:  generatedToken,
		MaxAge: 300,
		Path:   "/",
	})
	e := echo.New()
	c := e.NewContext(req, rec)

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
	// create dependencies
	jwtWrapper, sessionMgr := getWrapper()

	generatedToken, _, _ := jwtWrapper.GenerateToken("success", "msg", "false", "lastlogin", "username", "uuid")

	// mock client Do for handler
	// build our response JSON

	// create a new reader with that JSON

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

	sessionMgr.Client = client

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "ImageLink.com")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/createpost", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	req.AddCookie(&http.Cookie{
		Name:   "goRecycleCookie",
		Value:  generatedToken,
		MaxAge: 300,
		Path:   "/",
	})
	e := echo.New()
	c := e.NewContext(req, rec)

	// fmt.Println(rec.Body)

	if assert.NoError(t, CreatePost_POST(c, jwtWrapper, sessionMgr)) {
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

func TestEditPost_POST(t *testing.T) {
	// create dependencies
	jwtWrapper, sessionMgr := getWrapper()

	generatedToken, _, _ := jwtWrapper.GenerateToken("success", "msg", "false", "lastlogin", "username", "uuid")

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

	sessionMgr.Client = client

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "ImageLink.com")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/createpost", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	req.AddCookie(&http.Cookie{
		Name:   "goRecycleCookie",
		Value:  generatedToken,
		MaxAge: 300,
		Path:   "/",
	})
	e := echo.New()
	c := e.NewContext(req, rec)

	if assert.NoError(t, EditPost_POST(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusFound, rec.Code)
	}
}

func TestEditPost_POSTError(t *testing.T) {
	// create dependencies
	jwtWrapper, sessionMgr := getWrapper()

	generatedToken, _, _ := jwtWrapper.GenerateToken("success", "msg", "false", "lastlogin", "username", "uuid")

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

	sessionMgr.Client = client

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "<script>ImageLink.com")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/createpost", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	req.AddCookie(&http.Cookie{
		Name:   "goRecycleCookie",
		Value:  generatedToken,
		MaxAge: 300,
		Path:   "/",
	})
	e := echo.New()
	c := e.NewContext(req, rec)

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
	// create dependencies
	jwtWrapper, sessionMgr := getWrapper()

	generatedToken, _, _ := jwtWrapper.GenerateToken("success", "msg", "false", "lastlogin", "username", "uuid")

	// mock client Do for handler
	// build our response JSON

	// create a new reader with that JSON

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

	sessionMgr.Client = client

	//mock form values to function post
	f := make(url.Values)
	f.Set("PostName", "john")
	f.Set("PostComment", "CommentItem")
	f.Set("PostCondition", "ConditionItem")
	f.Set("PostCat", "Cat")
	f.Set("PostContactMeetInfo", "ContactMeetInfo")
	f.Set("PostImg2", "ImageLink.com")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/createpost", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	req.AddCookie(&http.Cookie{
		Name:   "goRecycleCookie",
		Value:  generatedToken,
		MaxAge: 300,
		Path:   "/",
	})
	e := echo.New()
	c := e.NewContext(req, rec)

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
func getDummy(db string) (newMap map[string]string) {
	newMap = make(map[string]string)
	switch db {
	case "ItemListing":
		newMap["ID"] = "000001"
		newMap["Username"] = "user"
		newMap["Name"] = "user"
		newMap["ImageLink"] = "nil"
		newMap["DatePosted"] = "nil"
		newMap["CommentItem"] = "nil"
		newMap["ConditionItem"] = "nil"
		newMap["Cat"] = "nil"
		newMap["ContactMeetInfo"] = "nil"
		newMap["Completion"] = "nil"
	case "UserSecret":
		newMap["Username"] = "user"
		newMap["Password"] = "user"
		newMap["IsAdmin"] = "false"
		newMap["ID"] = "000001"
		newMap["CommentItem"] = "nil"
	case "UserInfo":
		newMap["ID"] = "000001"
		newMap["Username"] = "user"
		newMap["LastLogin"] = "20-7-2021"
		newMap["CommentItem"] = "nil"
		newMap["DateJoin"] = "20-7-2021"
	case "CommentItem":
		newMap["ID"] = "000001"
		newMap["Username"] = "user"
		newMap["ForItem"] = "000001"
		newMap["CommentItem"] = "nil"
		newMap["Date"] = "20-7-2021"
	}
	return
}

func TestEditPost_GET(t *testing.T) {
	// create dependencies
	jwtWrapper, sessionMgr := getWrapper()

	generatedToken, _, _ := jwtWrapper.GenerateToken("success", "msg", "false", "lastlogin", "username", "uuid")

	// mock client Do for handler
	// build our response JSON

	// create a new reader with that JSON

	client := &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			id := req.URL.Query().Get("id")
			db := req.URL.Query().Get("db")
			fmt.Println(req.URL, req.URL.Query(), id, db)

			assert.Equal(t, "ItemListing", db)

			mapResponse := make(map[string]interface{})
			mapResponse["DataInfo"] = []interface{}{getDummy("ItemListing")}
			var statusCode int

			if assert.Equal(t, "000001", id) {
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

	sessionMgr.Client = client

	//mock form values to function post
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/editpost/000001", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	req.AddCookie(&http.Cookie{
		Name:   "goRecycleCookie",
		Value:  generatedToken,
		MaxAge: 300,
	})
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/editpost/000001")
	c.SetParamNames("id")
	c.SetParamValues("000001")

	if assert.NoError(t, EditPost_GET(c, jwtWrapper, sessionMgr)) {
		assert.Equal(t, http.StatusSeeOther, rec.Code)
	}
}
