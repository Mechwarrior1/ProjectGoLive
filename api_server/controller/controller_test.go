package controller

import (
	"apiserver/encrypt"
	"apiserver/mysql"
	"database/sql"
	"encoding/json"
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
	db, mock := NewMock()
	dbHandler1 := mysql.DBHandler{
		db,
		string(encrypt.DecryptFromFile("secure/mysql", "secure/keys.xml"))}
	defer func() {
		dbHandler1.DB.Close()
	}()

	e := echo.New()
	e.GET("/api/v0/check", func(c echo.Context) error {
		return PwCheck(c, &dbHandler1)
	})

	// enter mock data into DB
	// mock UserSecret and UserInfo data
	bPassword, _ := bcrypt.GenerateFromPassword([]byte("john"), bcrypt.MinCost)

	rows := sqlmock.NewRows([]string{"ID", "Username", "Password", "IsAdmin", "CommentItem"}).
		AddRow("000001", "john", bPassword, "false", "nil")
	query := "Select \\* FROM my_db." + "UserSecret" + " WHERE Username = \\?"
	mock.ExpectQuery(query).WillReturnRows(rows)
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

		//assert.Equal(t, userJSON, rec.Body.String()) //to add test for reponse body

	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// func TestGetAllInfo(t *testing.T) {
// 	e := echo.New()
// 	e.GET("/api/v0/allinfo", func(c echo.Context) error {
// 		return controller.GetAllInfo(c, &dbHandler1, embed)
// 	})

// 	rec := httptest.NewRecorder()
// 	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
// 	c := e.NewContext(req, rec)
// 	c.SetParamNames("limit", "offset")
// 	c.SetParamValues("2", "0")

// 	if assert.NoError(t, apihandler.GetAllUser(c)) {
// 		assert.Equal(t, http.StatusOK, rec.Code)

// 		//assert.Equal(t, userJSON, rec.Body.String()) //to add test for reponse body

// 	}
// }

// func main() {
// 	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

// 	embed := word2vec.GetWord2Vec()
// 	dbHandler1 := mysql.OpenDB()
// 	defer dbHandler1.DB.Close()

// 	e := echo.New()
// 	e.GET("/api/v0/check", func(c echo.Context) error {
// 		return controller.PwCheck(c, &dbHandler1)
// 	})
// 	e.GET("/api/v0/allinfo", func(c echo.Context) error {
// 		return controller.GetAllInfo(c, &dbHandler1, embed)
// 	})
// 	e.GET("/api/v0/username", func(c echo.Context) error {
// 		return controller.UsernameCheck(c, &dbHandler1)
// 	})
// 	e.POST("/api/v0/db/info", func(c echo.Context) error {
// 		return controller.GenInfoPost(c, &dbHandler1)
// 	})
// 	e.GET("/api/v0/db/info", func(c echo.Context) error {
// 		return controller.GenInfoGet(c, &dbHandler1)
// 	})
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
