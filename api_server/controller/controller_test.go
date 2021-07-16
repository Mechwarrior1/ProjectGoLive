package controller

import (
	"apiserver/encrypt"
	"apiserver/mysql"
	"apiserver/word2vec"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
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

	//mock for editing entry (UserInfo data)
	queryEdit := "UPDATE " + "UserInfo" + " SET LastLogin=\\?, CommentItem=\\? WHERE ID=\\?"
	prep := mock.ExpectPrepare(queryEdit)
	prep.ExpectExec().WithArgs("20-7-2021", "nil", "000001").WillReturnResult(sqlmock.NewResult(0, 1))

	// json payload to api
	newMap := make(map[string]interface{})
	newMap["ID"] = "000001"
	newMap["Username"] = "john"
	newMap["Password"] = "john"
	newMap["LastLogin"] = "20-7-2021"
	newMap["CommentItem"] = "nil"

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
	e.ServeHTTP(rec, req)
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

func TestGetAllInfoItemListing(t *testing.T) {
	// load variables
	embed := word2vec.GetWord2Vec()
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		""}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.GET("/api/v0/allinfo", func(c echo.Context) error {
		return GetAllInfo(c, &dbHandler1, embed)
	})

	// mock for querying
	rows := sqlmock.NewRows([]string{"ID", "Username", "Name", "ImageLink", "DatePosted", "CommentItem", "ConditionItem", "Cat", "ContactMeetInfo", "Completion"}).
		AddRow("000001", "john", "plastic", "1", "2", "3", "4", "5", "6", "7").
		AddRow("000002", "darren", "PET", "1", "2", "3", "4", "5", "6", "7")
	query := "Select \\* FROM my_db.ItemListing"
	mock.ExpectQuery(query).WillReturnRows(rows)

	// json payload to api
	newMap := make(map[string]interface{})
	newMap["Name"] = "plastics" //
	dataPacket1 := mysql.DataPacket{
		dbHandler1.ApiKey,
		"",
		"ItemListing",
		"",
		"",
		[]interface{}{newMap},
	}

	payloadJson, _ := json.Marshal(dataPacket1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v0/allinfo", strings.NewReader(string(payloadJson)))
	c := e.NewContext(req, rec)
	e.ServeHTTP(rec, req)
	if assert.NoError(t, GetAllInfo(c, &dbHandler1, embed)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		//check response
		json_map := make(map[string]interface{})
		json.NewDecoder(rec.Body).Decode(&json_map)
		json_map2 := json_map["DataInfo"].([]interface{})

		// fmt.Println(json_map2)
		assert.Equal(t, json_map2[0].(map[string]interface{})["ID"], "000001")

		//test word2vec as well
		assert.Equal(t, json_map2[0].(map[string]interface{})["Similarity"], float64(0.59042674))
		assert.Equal(t, json_map2[1].(map[string]interface{})["Similarity"], float64(0.38560766))
	}
}

func TestGetAllInfoCommentItem(t *testing.T) {
	// load variables
	var embed word2vec.Embeddings
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		string(encrypt.DecryptFromFile("secure/mysql", "secure/keys.xml"))}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.GET("/api/v0/allinfo", func(c echo.Context) error {
		return GetAllInfo(c, &dbHandler1, &embed)
	})

	// mock for querying
	rows := sqlmock.NewRows([]string{"ID", "Username", "ForItem", "Date", "CommentItem"}).
		AddRow("000001", "john", "000001", "1", "2").
		AddRow("000002", "darren", "000002", "1", "2").
		AddRow("000003", "darren", "000001", "1", "2")
	query := "Select \\* FROM my_db.CommentItem"
	mock.ExpectQuery(query).WillReturnRows(rows)

	// json payload to api
	newMap := make(map[string]interface{})
	newMap["ID"] = "000001" //
	dataPacket1 := mysql.DataPacket{
		dbHandler1.ApiKey,
		"",
		"CommentItem",
		"",
		"",
		[]interface{}{newMap},
	}

	payloadJson, _ := json.Marshal(dataPacket1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v0/allinfo", strings.NewReader(string(payloadJson)))
	c := e.NewContext(req, rec)
	e.ServeHTTP(rec, req)
	if assert.NoError(t, GetAllInfo(c, &dbHandler1, &embed)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		//check response
		json_map := make(map[string]interface{})
		json.NewDecoder(rec.Body).Decode(&json_map)
		json_map2 := json_map["DataInfo"].([]interface{})
		// fmt.Println(json_map2)

		//check if it returns only 1st and 3rd
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
	e.GET("/api/v0/username", func(c echo.Context) error {
		return UsernameCheck(c, &dbHandler1)
	})

	// mock for querying
	rows := sqlmock.NewRows([]string{"ID", "Username", "LastLogin", "DateJoin", "CommentItem"}).
		AddRow("000001", "john", "21-7-2021", "1", "2")
	query := "Select \\* FROM my_db.UserInfo WHERE Username = \\?"
	mock.ExpectQuery(query).WillReturnRows(rows)

	// json payload to api
	newMap := make(map[string]interface{})
	newMap["Username"] = "john" //
	dataPacket1 := mysql.DataPacket{
		dbHandler1.ApiKey,
		"",
		"UserInfo",
		"",
		"",
		[]interface{}{newMap},
	}

	payloadJson, _ := json.Marshal(dataPacket1)

	// test api
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v0/username", strings.NewReader(string(payloadJson)))
	c := e.NewContext(req, rec)
	e.ServeHTTP(rec, req)
	// fmt.Println(rec.Body)

	if assert.NoError(t, UsernameCheck(c, &dbHandler1)) {
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
		return UsernameCheck(c, &dbHandler1)
	})

	// mock for querying
	query := "Select \\* FROM my_db.UserInfo WHERE Username = \\?"
	mock.ExpectQuery(query)

	// json payload to api
	newMap := make(map[string]interface{})
	var dataPacket1 mysql.DataPacket
	dataPacket1.DataInfo = []interface{}{newMap}
	payloadJson, _ := json.Marshal(dataPacket1)

	// test api
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v0/username", strings.NewReader(string(payloadJson)))
	c := e.NewContext(req, rec)
	e.ServeHTTP(rec, req)
	// fmt.Println(rec.Body)

	if assert.NoError(t, UsernameCheck(c, &dbHandler1)) {
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
	e.ServeHTTP(rec, req)

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

	newMap := make(map[string]interface{})
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
	req := httptest.NewRequest(http.MethodGet, "/api/v0/db/info", strings.NewReader(string(payloadJson)))
	c := e.NewContext(req, rec)
	e.ServeHTTP(rec, req)
	// fmt.Println(rec.Body)

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

func TestGenInfoDelete(t *testing.T) {
	// load variables
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		""}
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.DELETE("/api/v0/db/info", func(c echo.Context) error {
		return GenInfoDelete(c, &dbHandler1)
	})

	// mock for querying
	query := "DELETE FROM ItemListing WHERE id = \\?"
	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("00001").WillReturnResult(sqlmock.NewResult(0, 1))

	// json payload
	newMap := make(map[string]interface{})
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
	req := httptest.NewRequest(http.MethodDelete, "/api/v0/db/info", strings.NewReader(string(payloadJson)))
	c := e.NewContext(req, rec)
	e.ServeHTTP(rec, req)
	fmt.Println(rec.Body)

	if assert.NoError(t, GenInfoDelete(c, &dbHandler1)) {
		assert.Equal(t, http.StatusOK, rec.Code)

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

// 	e.DELETE("/api/v0/db/info", func(c echo.Context) error {
// 		return controller.GenInfoDelete(c, &dbHandler1)
// 	})
// 	e.PUT("/api/v0/db/info", func(c echo.Context) error {
// 		return controller.GenInfoPut(c, &dbHandler1)
// 	})

// 	port := "5555"
// 	fmt.Println("listening at port " + port)
// 	s := http.Server{Addr: ":" + port, Handler: e}

// 	e.Use(middleware.Recover())
// 	e.Use(middleware.Logger())

// 	if err := s.ListenAndServeTLS("secure//cert.pem", "secure//key.pem"); err != nil && err != http.ErrServerClosed {
// 		e.Logger.Fatal(err)
// 	}

// }
