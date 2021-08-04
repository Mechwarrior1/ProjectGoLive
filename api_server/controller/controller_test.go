package controller

import (
	"apiserver/encrypt"
	"apiserver/mysql"
	"apiserver/word2vec"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql" // go mod init api_server.go
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func NewMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

func TestGetPwCheck(t *testing.T) {
	// load variables
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		""}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.GET("/api/v0/check", func(c echo.Context) error {
		return PwCheck(c, &dbHandler1)
	})

	// mock DB
	// mock for querying (UserSecret data)
	bPassword, _ := bcrypt.GenerateFromPassword([]byte("john"), bcrypt.MinCost)
	rows := sqlmock.NewRows([]string{"ID", "Username", "Password", "IsAdmin", "CommentItem"}).
		AddRow("000001", "john", bPassword, "false", "nil")
	query := "Select \\* FROM my_db." + "UserSecret" + " WHERE Username = \\?"
	mock.ExpectQuery(query).WillReturnRows(rows)

	rows2 := sqlmock.NewRows([]string{"ID", "Username", "LastLogin", "DateJoin", "CommentItem"}).
		AddRow("000001", "john", "11-7-2021", "2", "3")
	query2 := "Select \\* FROM my_db." + "UserInfo" + " WHERE Username = \\?"
	mock.ExpectQuery(query2).WillReturnRows(rows2)

	//mock for editing entry (UserInfo data)
	queryEdit := "UPDATE " + "UserInfo" + " SET LastLogin=\\?, CommentItem=\\? WHERE Username=\\?"
	prep := mock.ExpectPrepare(queryEdit)
	prep.ExpectExec().WithArgs("20-7-2021", "3", "john").WillReturnResult(sqlmock.NewResult(0, 1))

	// json payload to api
	newMap := make(map[string]interface{})
	newMap["Username"] = "john"
	newMap["Password"] = "john"
	newMap["LastLogin"] = "20-7-2021"

	dataPacket1 := mysql.DataPacket{
		dbHandler1.ApiKey,
		"",
		"UserSecret",
		"",
		"",
		[]interface{}{newMap},
	}

	payloadJson, _ := json.Marshal(dataPacket1)

	// making the call to api
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v0/check", strings.NewReader(string(payloadJson)))
	c := e.NewContext(req, rec)

	if assert.NoError(t, PwCheck(c, &dbHandler1)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		//check response
		json_map := make(map[string]interface{})
		json.NewDecoder(rec.Body).Decode(&json_map)
		assert.Equal(t, json_map["ResBool"], "true")

	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

/// this test requires a word2vec binary file
/// to be updated along with the yml file for a w2v file
// func TestGetAllListing(t *testing.T) {
// 	// load variables
// 	embed := word2vec.GetWord2Vec()
// 	db, mock := NewMock()
// 	dbHandler1 := mysql.DBHandler{
// 		db,
// 		""}
// 	defer dbHandler1.DB.Close()

// 	e := echo.New()
// 	e.GET("/api/v0/listing", func(c echo.Context) error {
// 		return GetAllListing(c, &dbHandler1, embed)
// 	})

// 	// mock for querying
// 	rows := sqlmock.NewRows([]string{"ID", "Username", "Name", "ImageLink", "DatePosted", "CommentItem", "ConditionItem", "Cat", "ContactMeetInfo", "Completion"}).
// 		AddRow("000001", "john", "plastic", "1", "2", "3", "4", "5", "6", "7").
// 		AddRow("000002", "darren", "PET", "1", "2", "3", "4", "5", "6", "7")
// 	query := "Select \\* FROM my_db.ItemListing"
// 	mock.ExpectQuery(query).WillReturnRows(rows)

// 	q := make(url.Values)
// 	q.Set("name", "Plastic")
// 	q.Set("filter", "")

// 	rec := httptest.NewRecorder()
// 	req := httptest.NewRequest(http.MethodGet, "/api/v0/listing/?"+q.Encode(), nil)
// 	fmt.Println("/api/v0/listing/?" + q.Encode())
// 	c := e.NewContext(req, rec)

// 	if assert.NoError(t, GetAllListing(c, &dbHandler1, embed)) {
// 		assert.Equal(t, http.StatusOK, rec.Code)

// 		//check response
// 		json_map := make(map[string]interface{})
// 		json.NewDecoder(rec.Body).Decode(&json_map)
// 		json_map2 := json_map["DataInfo"].([]interface{})

// 		// fmt.Println(json_map2)
// 		assert.Equal(t, json_map2[0].(map[string]interface{})["ID"], "000001")

// 		//test word2vec as well
// 		assert.Equal(t, json_map2[0].(map[string]interface{})["Similarity"], float64(0.65796596))
// 		assert.Equal(t, json_map2[1].(map[string]interface{})["Similarity"], float64(0.34202674))
// 	}
// }

func TestGetAllComment(t *testing.T) {
	// load variables
	var embed word2vec.Embeddings
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		string(encrypt.DecryptFromFile("secure/mysql", "secure/keys.xml"))}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.GET("/api/v0/comment", func(c echo.Context) error {
		return GetAllComment(c, &dbHandler1, &embed)
	})

	// mock for querying
	rows := sqlmock.NewRows([]string{"ID", "Username", "ForItem", "Date", "CommentItem"}).
		AddRow("000001", "john", "000001", "1", "2").
		AddRow("000002", "darren", "000002", "1", "2").
		AddRow("000003", "darren", "000001", "1", "2")
	query := "Select \\* FROM my_db.CommentItem"
	mock.ExpectQuery(query).WillReturnRows(rows)

	//query parameters
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v0/comment", nil)
	c := e.NewContext(req, rec)

	c.SetPath("/api/v0/comment/:id")
	c.SetParamNames("id")
	c.SetParamValues("000001")

	if assert.NoError(t, GetAllComment(c, &dbHandler1, &embed)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		//check response
		json_map := make(map[string]interface{})
		json.NewDecoder(rec.Body).Decode(&json_map)
		json_map2 := json_map["DataInfo"].([]interface{})

		// check if it returns only 1st and 3rd
		assert.Equal(t, json_map2[0].(map[string]interface{})["ID"], "000001")
		assert.Equal(t, json_map2[1].(map[string]interface{})["ID"], "000003")
	}
}

func TestUsernameCheckTrue(t *testing.T) {
	// load variables
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		""}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.GET("/api/v0/username/:username", func(c echo.Context) error {
		return CheckUsername(c, &dbHandler1)
	})

	// mock for querying
	rows := sqlmock.NewRows([]string{"ID", "Username", "LastLogin", "DateJoin", "CommentItem"}).
		AddRow("000001", "john", "21-7-2021", "1", "2")
	query := "Select \\* FROM my_db.UserInfo WHERE Username = \\?"
	mock.ExpectQuery(query).WillReturnRows(rows)

	// test api
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	c := e.NewContext(req, rec)
	c.SetPath("/api/v0/username/:username")
	c.SetParamNames("username")
	c.SetParamValues("john")

	// fmt.Println(rec.Body)

	if assert.NoError(t, CheckUsername(c, &dbHandler1)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		//check response
		json_map := make(map[string]interface{})
		json.NewDecoder(rec.Body).Decode(&json_map)

		assert.Equal(t, json_map["ResBool"], "true")

	}
}

func TestUsernameCheckFalse(t *testing.T) {
	// load variables
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		""}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.GET("/api/v0/username", func(c echo.Context) error {
		return CheckUsername(c, &dbHandler1)
	})

	// mock for querying
	query := "Select \\* FROM my_db.UserInfo WHERE Username = \\?"
	mock.ExpectQuery(query)

	// test api
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v0/username", nil)

	c := e.NewContext(req, rec)
	c.SetPath("/api/v0/username/:username")
	c.SetParamNames("username")
	c.SetParamValues("")

	// fmt.Println(rec.Body)

	if assert.NoError(t, CheckUsername(c, &dbHandler1)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code) //bad request
	}
}

func TestGenInfoPost(t *testing.T) {
	// load variables
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		""}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.POST("/api/v0/db/info", func(c echo.Context) error {
		return GenInfoPost(c, &dbHandler1)
	})

	// mock for querying
	query := "SELECT MAX\\(ID\\) FROM my_db.ItemListing" //for MaxID query
	rows := sqlmock.NewRows([]string{"ID"}).
		AddRow(1)
	mock.ExpectQuery(query).WillReturnRows(rows)

	query2 := "INSERT INTO my_db.ItemListing VALUES \\(\\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?\\)"

	prep := mock.ExpectPrepare(query2)
	prep.ExpectExec().WithArgs("000002", "john", "johnee", "nil", "nil", "nil", "nil", "nil", "nil", "nil").WillReturnResult(sqlmock.NewResult(0, 1))

	// json payload to api
	newMap := make(map[string]interface{})
	newMap["Username"] = "john"
	newMap["Name"] = "johnee"
	newMap["ImageLink"] = "nil"
	newMap["DatePosted"] = "nil"
	newMap["CommentItem"] = "nil"
	newMap["ConditionItem"] = "nil"
	newMap["Cat"] = "nil"
	newMap["ContactMeetInfo"] = "nil"
	newMap["Completion"] = "nil"

	dataPacket1 := mysql.DataPacket{
		dbHandler1.ApiKey,
		"",
		"ItemListing",
		"",
		"",
		[]interface{}{newMap},
	}

	payloadJson, _ := json.Marshal(dataPacket1)

	// test api
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v0/db/info", strings.NewReader(string(payloadJson)))
	c := e.NewContext(req, rec)

	if assert.NoError(t, GenInfoPost(c, &dbHandler1)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		//check response
		json_map := make(map[string]interface{})
		json.NewDecoder(rec.Body).Decode(&json_map)

		assert.Equal(t, json_map["ResBool"], "true")

	}
	// add more for other database tables
}

func TestGenInfoGet(t *testing.T) {
	// load variables
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		""}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.GET("/api/v0/db/info", func(c echo.Context) error {
		return GenInfoGet(c, &dbHandler1)
	})

	// mock for querying
	rows := sqlmock.NewRows([]string{"ID", "Username", "Name", "ImageLink", "DatePosted", "CommentItem", "ConditionItem", "Cat", "ContactMeetInfo", "Completion"}).
		AddRow("000001", "john", "plastic", "www.plasticsimage.com", "20-7-2021", "plastics for all", "Worn out", "Plastic", "see profile", "false")

	query := "Select \\* FROM my_db.ItemListing WHERE ID = \\?"
	mock.ExpectQuery(query).WithArgs("000001").WillReturnRows(rows)

	// test api
	q := make(url.Values)
	q.Set("id", "000001")
	q.Set("db", "ItemListing")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v0/db/info?"+q.Encode(), nil)
	c := e.NewContext(req, rec)

	if assert.NoError(t, GenInfoGet(c, &dbHandler1)) {
		assert.Equal(t, http.StatusCreated, rec.Code)

		//check response
		json_map := make(map[string]interface{})
		json.NewDecoder(rec.Body).Decode(&json_map)
		json_map2 := json_map["DataInfo"].([]interface{})

		//check some of the returned inputs
		assert.Equal(t, json_map["ResBool"], "true")
		assert.Equal(t, json_map2[0].(map[string]interface{})["ID"], "000001")
		assert.Equal(t, json_map2[0].(map[string]interface{})["Name"], "plastic")

	}
}

// func TestGenInfoDelete(t *testing.T) {
// 	// load variables
// 	db, mock := NewMock()
// 	dbHandler1 := mysql.DBHandler{
// 		db,
// 		""}
// 	defer dbHandler1.DB.Close()

// 	e := echo.New()
// 	e.DELETE("/api/v0/db/info", func(c echo.Context) error {
// 		return GenInfoDelete(c, &dbHandler1)
// 	})

// 	// mock for querying
// 	query := "DELETE FROM ItemListing WHERE id = \\?"
// 	prep := mock.ExpectPrepare(query)
// 	prep.ExpectExec().WithArgs("00001").WillReturnResult(sqlmock.NewResult(0, 1))

// 	// json payload
// 	newMap := make(map[string]interface{})
// 	newMap["ID"] = "000001"

// 	dataPacket1 := mysql.DataPacket{
// 		dbHandler1.ApiKey,
// 		"",
// 		"ItemListing",
// 		"",
// 		"",
// 		[]interface{}{newMap},
// 	}
// 	payloadJson, _ := json.Marshal(dataPacket1)

// 	// test api
// 	rec := httptest.NewRecorder()
// 	req := httptest.NewRequest(http.MethodDelete, "/api/v0/db/info", strings.NewReader(string(payloadJson)))
// 	c := e.NewContext(req, rec)
// 	fmt.Println(rec.Body)

// 	if assert.NoError(t, GenInfoDelete(c, &dbHandler1)) {
// 		assert.Equal(t, http.StatusOK, rec.Code)

// 		//check response
// 		json_map := make(map[string]interface{})
// 		json.NewDecoder(rec.Body).Decode(&json_map)

// 		//check some of the returned inputs
// 		assert.Equal(t, json_map["ResBool"], "true")

// 	}
// }

func TestGenInfoEdit(t *testing.T) {
	// load variables
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		""}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.PUT("/api/v0/db/info", func(c echo.Context) error {
		return GenInfoPut(c, &dbHandler1)
	})

	// mock for querying
	query := "UPDATE ItemListing SET ImageLink=\\?, CommentItem=\\?, ConditionItem=\\?, Cat=\\?, ContactMeetInfo=\\?, Completion=\\? WHERE ID=\\?"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("image", "comment", "condition", "cat", "contact", "false", "000001").WillReturnResult(sqlmock.NewResult(0, 1))

	// json payload
	newMap := make(map[string]interface{})
	newMap["ImageLink"] = "image"
	newMap["CommentItem"] = "comment"
	newMap["ConditionItem"] = "condition"
	newMap["Cat"] = "cat"
	newMap["ContactMeetInfo"] = "contact"
	newMap["Completion"] = "false"
	newMap["ID"] = "000001"

	dataPacket1 := mysql.DataPacket{
		dbHandler1.ApiKey,
		"",
		"ItemListing",
		"",
		"",
		[]interface{}{newMap},
	}
	payloadJson, _ := json.Marshal(dataPacket1)

	// test api
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v0/db/info", strings.NewReader(string(payloadJson)))
	c := e.NewContext(req, rec)

	// fmt.Println(rec.Body)

	if assert.NoError(t, GenInfoPut(c, &dbHandler1)) {
		assert.Equal(t, http.StatusCreated, rec.Code)

		//check response
		json_map := make(map[string]interface{})
		json.NewDecoder(rec.Body).Decode(&json_map)

		//check some of the returned inputs
		assert.Equal(t, json_map["ResBool"], "true")

	}
}

// func main() {
// 	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

// 	embed := word2vec.GetWord2Vec()
// 	dbHandler1 := mysql.OpenDB()
// 	defer dbHandler1.DB.Close()

// 	e := echo.New()

// 	port := "5555"
// 	fmt.Println("listening at port " + port)
// 	s := http.Server{Addr: ":" + port, Handler: e}

// 	e.Use(middleware.Recover())
// 	e.Use(middleware.Logger())

// 	if err := s.ListenAndServeTLS("secure//cert.pem", "secure//key.pem"); err != nil && err != http.ErrServerClosed {
// 		e.Logger.Fatal(err)
// 	}

// }

func TestSignup(t *testing.T) {
	// load variables
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		""}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.GET("/api/v0/signup", func(c echo.Context) error {
		return Signup(c, &dbHandler1)
	})

	// mock for querying
	query := "SELECT MAX\\(ID\\) FROM my_db.UserSecret" //for MaxID query
	rows := mock.NewRows([]string{"ID"}).
		AddRow("000001")
	mock.ExpectQuery(query).WillReturnRows(rows)

	mock.ExpectBegin()

	//mock for usersecret
	query2 := "INSERT INTO my_db." + "UserSecret" + " VALUES \\(\\?, \\?, \\?, \\?, \\?\\)"
	prep2 := mock.ExpectPrepare(query2)
	prep2.ExpectExec().WithArgs("000002", "john", "john", "", "nil").WillReturnResult(sqlmock.NewResult(0, 1))

	//mock for userinfo
	query3 := "INSERT INTO my_db." + "UserInfo" + " VALUES \\(\\?, \\?, \\?, \\?, \\?\\)"
	prep3 := mock.ExpectPrepare(query3)
	prep3.ExpectExec().WithArgs("000002", "john", "20-7-2021", "20-7-2021", "nil").WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// test api

	userSecret1 := make(map[string]interface{})
	userSecret1["Username"] = "john"
	userSecret1["Password"] = "john"
	userSecret1["CommentItem"] = "nil"
	userSecret1["DateJoin"] = "20-7-2021"
	userSecret1["LastLogin"] = "20-7-2021"

	jsonData1 := mysql.DataPacket{
		// key to access rest api
		Key:         dbHandler1.ApiKey,
		ErrorMsg:    "",
		InfoType:    "UserSecret",
		ResBool:     "",
		RequestUser: "",
		DataInfo:    []interface{}{userSecret1},
	}

	payloadJson, _ := json.Marshal(jsonData1)

	// making the call to api
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v0/signup", strings.NewReader(string(payloadJson)))

	c := e.NewContext(req, rec)

	// fmt.Println(rec.Body)

	if assert.NoError(t, Signup(c, &dbHandler1)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		//check response
		json_map := make(map[string]interface{})
		json.NewDecoder(rec.Body).Decode(&json_map)

		assert.Equal(t, json_map["ResBool"], "true")

	}
}

func TestCompleted(t *testing.T) {
	// load variables
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		""}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.GET("/api/v0/completed/:id", func(c echo.Context) error {
		return Completed(c, &dbHandler1)
	})

	// mock for querying
	query := "SELECT Username FROM my_db.ItemListing WHERE ID=000001"
	rows := mock.NewRows([]string{"Username"}).
		AddRow("john") //apparently there is no logic and does not check for largest, willreturnrows directly just returns
	mock.ExpectQuery(query).WillReturnRows(rows)

	//mock for update completed
	mock.ExpectBegin()
	query2 := "UPDATE my_db.ItemListing SET Completion =\\? WHERE ID=\\?"
	prep2 := mock.ExpectPrepare(query2)
	prep2.ExpectExec().WithArgs("true", "000001").WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// test api

	jsonData1 := mysql.DataPacket{
		// key to access rest api
		Key:         dbHandler1.ApiKey,
		ErrorMsg:    "",
		InfoType:    "ItemListing",
		ResBool:     "",
		RequestUser: "john",
		DataInfo:    []interface{}{},
	}

	payloadJson, _ := json.Marshal(jsonData1)

	// making the call to api
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v0/completed/", strings.NewReader(string(payloadJson)))

	c := e.NewContext(req, rec)
	c.SetPath("/api/v0/completed/:id")
	c.SetParamNames("id")
	c.SetParamValues("000001")
	// fmt.Println(rec.Body)

	if assert.NoError(t, Completed(c, &dbHandler1)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		//check response
		json_map := make(map[string]interface{})
		json.NewDecoder(rec.Body).Decode(&json_map)

		assert.Equal(t, json_map["ResBool"], "true")

	}
}
