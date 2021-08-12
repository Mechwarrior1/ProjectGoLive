package mysql

import (
	"database/sql"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func NewMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

//tested getting item with ID
func TestGetSingleRecord(t *testing.T) {
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	dbTableAll := []string{"UserInfo", "UserSecret", "ItemListing", "CommentUser", "CommentItem"}

	rows := sqlmock.NewRows([]string{"ID"})

	for _, dbTable := range dbTableAll {
		query := "Select \\* FROM my_db." + dbTable + " WHERE id = \\?"

		switch dbTable {

		case "UserInfo":
			rows = sqlmock.NewRows([]string{"ID", "Username", "LastLogin", "DateJoin", "CommentItem"}).
				AddRow("000000", "john", "20-7-2021", "20-7-2021", "nil")

		case "UserSecret":
			rows = sqlmock.NewRows([]string{"ID", "Username", "Password", "IsAdmin", "CommentItem"}).
				AddRow("000000", "john", "123", "true", "nil")

		case "ItemListing":
			rows = sqlmock.NewRows([]string{"ID", "Username", "Name", "ImageLink", "DatePosted", "CommentItem", "ConditionItem", "Cat", "ContactMeetInfo", "Completion"}).
				AddRow("000000", "john", "plastic", "www.plasticsimage.com", "20-7-2021", "plastics for all", "Worn out", "Plastic", "see profile", "false")

		case "CommentUser":
			rows = sqlmock.NewRows([]string{"ID", "Username", "ForUsername", "Date", "CommentItem"}).
				AddRow("000000", "john", "darren", "20-7-2021", "nil")

		case "CommentItem":
			rows = sqlmock.NewRows([]string{"ID", "Username", "ForItem", "Date", "CommentItem"}).
				AddRow("000000", "john", "Plastics", "20-7-2021", "nil")
		}

		mock.ExpectQuery(query).WithArgs("000000").WillReturnRows(rows)

		user, err := dbHandler1.GetSingleRecord(dbTable, "WHERE id = ?", "000000")
		assert.NotNil(t, user)
		assert.NoError(t, err)
	}
}

func TestInsertRecord1(t *testing.T) {
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	dbTable := "UserInfo"

	query := "INSERT INTO my_db." + dbTable + " VALUES \\(\\?, \\?, \\?, \\?, \\?\\)"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("000001", "john", "20-7-2021", "20-7-2021", "nil").WillReturnResult(sqlmock.NewResult(0, 1))

	newMap := make(map[string]string)

	newMap["Username"] = "john"
	newMap["LastLogin"] = "20-7-2021"
	newMap["DateJoin"] = "20-7-2021"
	newMap["CommentItem"] = "nil"

	err := dbHandler1.InsertRecord(dbTable, newMap, "000001")
	assert.NoError(t, err)
}

func TestInsertRecord2(t *testing.T) {
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	dbTable := "UserSecret"

	query := "INSERT INTO my_db." + dbTable + " VALUES \\(\\?, \\?, \\?, \\?, \\?\\)"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("000001", "john", "20-7-2021", "20-7-2021", "nil").WillReturnResult(sqlmock.NewResult(0, 1))

	newMap := make(map[string]string)

	newMap["Username"] = "john"
	newMap["Password"] = "20-7-2021"
	newMap["IsAdmin"] = "20-7-2021"
	newMap["CommentItem"] = "nil"

	err := dbHandler1.InsertRecord(dbTable, newMap, "000001")
	assert.NoError(t, err)
}

func TestInsertRecord3(t *testing.T) {
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	dbTable := "ItemListing"

	query := "INSERT INTO my_db." + dbTable + " VALUES \\(\\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?\\)"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("000001", "john", "johnee", "nil", "nil", "nil", "nil", "nil", "nil", "nil").WillReturnResult(sqlmock.NewResult(0, 1))

	newMap := make(map[string]string)

	newMap["Username"] = "john"
	newMap["Name"] = "johnee"
	newMap["ImageLink"] = "nil"
	newMap["DatePosted"] = "nil"
	newMap["CommentItem"] = "nil"
	newMap["ConditionItem"] = "nil"
	newMap["Cat"] = "nil"
	newMap["ContactMeetInfo"] = "nil"
	newMap["Completion"] = "nil"

	err := dbHandler1.InsertRecord(dbTable, newMap, "000001")
	assert.NoError(t, err)
}

func TestInsertRecord4(t *testing.T) {
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	dbTable := "CommentUser"

	query := "INSERT INTO my_db." + dbTable + " VALUES \\(\\?, \\?, \\?, \\?, \\?\\)"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("000001", "john", "johnee", "20-7-2021", "nil").WillReturnResult(sqlmock.NewResult(0, 1))

	newMap := make(map[string]string)

	newMap["Username"] = "john"
	newMap["ForUsername"] = "johnee"
	newMap["Date"] = "20-7-2021"
	newMap["CommentItem"] = "nil"

	err := dbHandler1.InsertRecord(dbTable, newMap, "000001")
	assert.NoError(t, err)
}

func TestInsertRecord5(t *testing.T) {
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	dbTable := "CommentItem"

	query := "INSERT INTO my_db." + dbTable + " VALUES \\(\\?, \\?, \\?, \\?, \\?\\)"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("000001", "john", "cartoon", "20-7-2021", "nil").WillReturnResult(sqlmock.NewResult(0, 1))

	newMap := make(map[string]string)

	newMap["Username"] = "john"
	newMap["ForItem"] = "cartoon"
	newMap["Date"] = "20-7-2021"
	newMap["CommentItem"] = "nil"

	err := dbHandler1.InsertRecord(dbTable, newMap, "000001")
	assert.NoError(t, err)
}

func TestEditRecord1(t *testing.T) {
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	dbTable := "UserInfo"

	query := "UPDATE " + dbTable + " SET LastLogin=\\?, CommentItem=\\? WHERE ID=\\?"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("20-7-2021", "nil", "000001").WillReturnResult(sqlmock.NewResult(0, 1))

	newMap := make(map[string]string)

	newMap["LastLogin"] = "20-7-2021"
	newMap["ID"] = "000001"
	newMap["CommentItem"] = "nil"

	err := dbHandler1.EditRecord(dbTable, newMap)
	assert.NoError(t, err)
}

func TestEditRecord2(t *testing.T) {
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	dbTable := "ItemListing"

	query := "UPDATE " + dbTable + " SET ImageLink=\\?, CommentItem=\\?, ConditionItem=\\?, Cat=\\?, ContactMeetInfo=\\?, Completion=\\? WHERE ID=\\?"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("image", "comment", "condition", "cat", "contact", "false", "000001").WillReturnResult(sqlmock.NewResult(0, 1))

	newMap := make(map[string]string)

	newMap["ImageLink"] = "image"
	newMap["CommentItem"] = "comment"
	newMap["ConditionItem"] = "condition"
	newMap["Cat"] = "cat"
	newMap["ContactMeetInfo"] = "contact"
	newMap["Completion"] = "false"
	newMap["ID"] = "000001"

	err := dbHandler1.EditRecord(dbTable, newMap)
	assert.NoError(t, err)
}

func TestEditRecord3(t *testing.T) {
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	dbTable := "CommentUser"

	query := "UPDATE " + dbTable + " SET CommentItem=\\? WHERE ID=\\?"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("comment", "000001").WillReturnResult(sqlmock.NewResult(0, 1))

	newMap := make(map[string]string)

	newMap["ID"] = "000001"
	newMap["CommentItem"] = "comment"

	err := dbHandler1.EditRecord(dbTable, newMap)
	assert.NoError(t, err)
}

func TestEditRecord4(t *testing.T) {
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	dbTable := "CommentItem"

	query := "UPDATE " + dbTable + " SET CommentItem=\\? WHERE ID=\\?"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("comment", "000001").WillReturnResult(sqlmock.NewResult(0, 1))

	newMap := make(map[string]string)

	newMap["ID"] = "000001"
	newMap["CommentItem"] = "comment"

	err := dbHandler1.EditRecord(dbTable, newMap)
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	dbTable := "ItemListing"
	query := "DELETE FROM " + dbTable + " WHERE id = \\?"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs("00001").WillReturnResult(sqlmock.NewResult(0, 1))

	err := dbHandler1.DeleteRecord(dbTable, "000001")
	assert.Error(t, err)
}

func TestGetMaxID(t *testing.T) {
	// load variables
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	// mock for querying
	query := "SELECT MAX\\(ID\\) FROM my_db.UserSecret" //for MaxID query
	rows := mock.NewRows([]string{"ID"}).
		AddRow("000001") //apparently there is no logic and does not check for largest, willreturnrows directly just returns
	mock.ExpectQuery(query).WillReturnRows(rows)

	num, err := dbHandler1.GetMaxID("UserSecret")
	assert.Equal(t, num, 1, "should be the same")
	assert.NoError(t, err)
}

func TestGetUsername(t *testing.T) {
	// load variables
	db, mock := NewMock()
	dbHandler1 := DBHandler{db, "", true}
	defer func() {
		dbHandler1.DB.Close()
	}()

	// mock for querying

	query := "SELECT Username FROM my_db.UserSecret WHERE ID=000001"
	rows := mock.NewRows([]string{"Username"}).
		AddRow("john") //apparently there is no logic and does not check for largest, willreturnrows directly just returns
	mock.ExpectQuery(query).WillReturnRows(rows)

	result, err := dbHandler1.GetUsername("UserSecret", "000001")
	assert.Equal(t, result, "john", "should be the same")
	assert.NoError(t, err)
}
