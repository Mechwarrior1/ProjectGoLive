package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	_ "github.com/go-sql-driver/mysql" // go mod init api_server.go
)

type (
	userSecret struct {
		ID          string `json:"ID"`
		Username    string `json:"Username"`
		Password    string `json:"Password"`
		IsAdmin     string `json:"IsAdmin"`
		CommentItem string `json:"CommentItem"`
	}

	userInfo struct {
		ID          string `json:"ID"`
		Username    string `json:"Username"`
		LastLogin   string `json:"LastLogin"`
		DateJoin    string `json:"DateJoin"`
		CommentItem string `json:"CommentItem"`
	}

	itemListing struct {
		ID              string `json:"ID"`
		Username        string `json:"Username"`
		Name            string `json:"Name"`
		ImageLink       string `json:"ImageLink"`
		DatePosted      string `json:"DatePosted"`
		CommentItem     string `json:"CommentItem"`
		ConditionItem   string `json:"ConditionItem"`
		Cat             string `json:"Cat"`
		ContactMeetInfo string `json:"ContactMeetInfo"`
		Similarity      string `json:"Similarity"`
		Completion      string `json:"Completion"`
	}

	commentUser struct {
		ID          string `json:"ID"`
		Username    string `json:"Username"`
		ForUsername string `json:"ForUsername"`
		Date        string `json:"Date"`
		CommentItem string `json:"CommentItem"`
	}
	commentItem struct {
		ID          string `json:"ID"`
		Username    string `json:"Username"`
		ForItem     string `json:"ForItem"`
		Date        string `json:"Date"`
		CommentItem string `json:"CommentItem"`
	}

	dataPacket struct {
		// key to access rest api
		Key         string        `json:"Key"`
		ErrorMsg    string        `json:"ErrorMsg"`
		InfoType    string        `json:"InfoType"` // 5 types: userSecret, userInfo, itemListing, commentUser, commentItem
		ResBool     string        `json:"ResBool"`
		RequestUser string        `json:"RequestUser"`
		DataInfo    []interface{} `json:"DataInfo"`
	}

	//
	dbHandler struct {
		DB *sql.DB
	}
)

// Opens db and returns a struct to access it
func openDB() dbHandler {
	pass := string(decryptFromFile("secure/mysql"))
	db, err := sql.Open("mysql", "myuser:"+pass+"@tcp(127.0.0.1:60575)/my_db")
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("no issue")
	}

	dbHandler1 := dbHandler{db}
	return dbHandler1
}

/*
UserSecret  (ID , Username, Password, IsAdmin, CommentItem);
UserInfo    (ID , Username, LastLogin, DateJoin, CommentItem);
ItemListing (ID , Username, Name, ImageLink, DatePosted INT, CommentItem, ConditionItem, Cat, ContactMeetInfo);
CommentUser (ID , Username, ForUsername, Date, CommentItem);
CommentItem (ID , Username, ForItem, Date, CommentItem);
*/

// func (s userSecret) returnParameter() (*string, *string, *string, *string, *string) {
// 	return &(s.ID), &(s.Username), &(s.Password), &(s.IsAdmin), &(s.CommentItem)
// }

// func (s userInfo) returnParameter() (*string, *string, *string, *string, *string) {
// 	return &s.ID, &s.Username, &s.LastLogin, &s.DateJoin, &s.CommentItem
// }

// func (s itemListing) returnParameter() (*string, *string, *string, *string, *int, *string, *string, *string, *string) {
// 	return &s.ID, &s.Username, &s.Name, &s.ImageLink, &s.DatePosted, &s.CommentItem, &s.ConditionItem, &s.Cat, &s.ContactMeetInfo
// }

// func (s commentUser) returnParameter() (*string, *string, *string, *string, *string) {
// 	return &s.ID, &s.Username, &s.ForUsername, &s.Date, &s.CommentItem
// }

// func (s commentItem) returnParameter() (*string, *string, *string, *string, *string) {
// 	return &s.ID, &s.Username, &s.ForItem, &s.Date, &s.CommentItem
// }

// type genData interface {
// 	returnParameter() (*string, *string, *string, *string, *string)
// }

// func getStruct(dbTable string) (genData, interface{}) {
// 	var data1 genData
// 	switch dbTable {
// 	case "UserSecret":
// 		item := userSecret{}
// 		data1 = item
// 		return data1, item
// 	case "UserInfo":
// 		item := userInfo{}
// 		data1 = item
// 		return data1, item
// 	// case "itemListing":
// 	// 	data1 = itemListing{}  //needs a seperate call due to different output
// 	case "CommentUser":
// 		item := commentUser{}
// 		data1 = item
// 		return data1, item
// 	case "CommentItem":
// 		item := commentItem{}
// 		data1 = item
// 		return data1, item
// 	}
// 	return data1, nil
// }

func appendNoError(allData []interface{}, data1 interface{}, err error) []interface{} {
	if err != nil {
		fmt.Println("log error: s" + err.Error())
		return allData
	}
	allData = append(allData, data1)
	return allData
}

// access the DB and get all records
func (dbHandler dbHandler) getRecord(dbTable string) ([]interface{}, error) {
	// allData := []genData{}
	allData := make([]interface{}, 0)
	results, err := dbHandler.DB.Query("Select * FROM my_db." + dbTable)
	if err != nil {
		return allData, err
	}
	for results.Next() {
		switch dbTable {
		// case "UserSecret":
		// 	data1 := userSecret{}
		// 	err = results.Scan(&data1.ID, &data1.Username, &data1.Password, &data1.IsAdmin, &data1.CommentItem)
		// 	allData = appendNoError(allData, data1, err)
		case "UserInfo":
			data1 := userInfo{}
			err = results.Scan(&data1.ID, &data1.Username, &data1.LastLogin, &data1.DateJoin, &data1.CommentItem)
			allData = appendNoError(allData, data1, err)
		case "ItemListing":
			data1 := itemListing{} //needs a seperate call due to different output
			err = results.Scan(&data1.ID, &data1.Username, &data1.Name, &data1.ImageLink, &data1.DatePosted, &data1.CommentItem, &data1.ConditionItem, &data1.Cat, &data1.ContactMeetInfo, &data1.Completion)
			allData = appendNoError(allData, data1, err)
		case "CommentUser":
			data1 := commentUser{}
			err = results.Scan(&data1.ID, &data1.Username, &data1.ForUsername, &data1.Date, &data1.CommentItem)
			allData = appendNoError(allData, data1, err)
		case "CommentItem":
			data1 := commentItem{}
			err = results.Scan(&data1.ID, &data1.Username, &data1.ForItem, &data1.Date, &data1.CommentItem)
			allData = appendNoError(allData, data1, err)
		default:
			return allData, errors.New(dbTable + " not found")
		}
	}
	return allData, nil
}

// access the DB and get all records
func (dbHandler dbHandler) getRecordlisting(dbTable string, requestWords string) ([]interface{}, error) {
	// allData := []genData{}
	allData := []interface{}{}
	requestWords2 := strings.Fields(requestWords) //split the words for embeding
	requestWordsEmbed := embed.getWordEmbeddingCombine(requestWords2, []string{})
	results, err := dbHandler.DB.Query("Select * FROM my_db." + dbTable)

	if err != nil {
		return allData, err
	}
	for results.Next() {
		data1 := itemListing{} //needs a seperate call due to different output
		err = results.Scan(&data1.ID, &data1.Username, &data1.Name, &data1.ImageLink, &data1.DatePosted, &data1.CommentItem, &data1.ConditionItem, &data1.Cat, &data1.ContactMeetInfo, &data1.Completion)
		if err != nil {
			fmt.Println("logger: error at getRecordlisting:" + err.Error())
		}

		if data1.Completion != "true" {
			requestWordsEmbed2 := embed.getWordEmbeddingCombine(strings.Fields(data1.Name), []string{})
			addVal := float32(0)
			addVal2 := float32(0)

			for _, word := range requestWords2 {
				if strings.Contains(data1.Name, word) {
					fmt.Println("0.05, "+data1.Name, word)
					addVal += 0.05
				}
				if strings.Contains(data1.CommentItem, word) {
					addVal2 += 0.005
				}
			}

			addVal3 := math.Min(0.15, math.Max(float64(addVal2), 0))
			addVal4 := math.Min(0.2, math.Max(float64(addVal), 0))
			cosSim := cosineSimilarity(requestWordsEmbed, requestWordsEmbed2)
			data1.Similarity = fmt.Sprintf("%f", cosSim+float32(addVal3+addVal4))
			fmt.Println(requestWords, data1.Name, cosSim+float32(addVal3+addVal4))
			allData = append(allData, data1)
		}
	}
	return allData, nil
}

// access the DB and get a single record, search using courseName
func (dbHandler dbHandler) getSingleRecord(dbTable string, queryString string) ([]interface{}, error) {
	//queryString examples, " WHERE ID = 1" or "WHERE Username = alvin"
	// allData := []genData{}
	allData := make([]interface{}, 0)
	// fmt.Println("Select * FROM my_db." + dbTable + " " + queryString)
	results, err := dbHandler.DB.Query("Select * FROM my_db." + dbTable + " " + queryString)
	if err != nil {
		return allData, err
	}
	results.Next()
	switch dbTable {
	case "UserSecret":
		data1 := userSecret{}
		err = results.Scan(&data1.ID, &data1.Username, &data1.Password, &data1.IsAdmin, &data1.CommentItem)
		allData = appendNoError(allData, data1, err)
	case "UserInfo":
		data1 := userInfo{}
		err = results.Scan(&data1.ID, &data1.Username, &data1.LastLogin, &data1.DateJoin, &data1.CommentItem)
		allData = appendNoError(allData, data1, err)
	case "ItemListing":
		data1 := itemListing{} //needs a seperate call due to different output
		err = results.Scan(&data1.ID, &data1.Username, &data1.Name, &data1.ImageLink, &data1.DatePosted, &data1.CommentItem, &data1.ConditionItem, &data1.Cat, &data1.ContactMeetInfo, &data1.Completion)
		allData = appendNoError(allData, data1, err)
	case "CommentUser":
		data1 := commentUser{}
		err = results.Scan(&data1.ID, &data1.Username, &data1.ForUsername, &data1.Date, &data1.CommentItem)
		allData = appendNoError(allData, data1, err)
	case "CommentItem":
		data1 := commentItem{}
		err = results.Scan(&data1.ID, &data1.Username, &data1.ForItem, &data1.Date, &data1.CommentItem)
		allData = appendNoError(allData, data1, err)
	default:
		return allData, errors.New(dbTable + " not found in switch")
	}
	return allData, err
}

// func (dbHandler dbHandler) insertRecord(dbTable string, ID string, courseName string, courseTitle string, courseDesc string) error {
// 	_, err := dbHandler.DB.Exec("INSERT INTO my_db."+dbTable+" VALUES (?,?,?,?,?)",
// 		ID, courseName, courseTitle, courseDesc)
// 	return err
// }

// // post a record into the DB
func (dbHandler dbHandler) insertRecord(dbTable string, values ...string) error {
	var err error
	switch dbTable {
	case "UserSecret":
		_, err = dbHandler.DB.Exec("INSERT INTO my_db."+dbTable+" VALUES (?,?,?,?,?)", values[0], values[1], values[2], values[3], values[4])
	case "UserInfo":
		_, err = dbHandler.DB.Exec("INSERT INTO my_db."+dbTable+" VALUES (?,?,?,?,?)", values[0], values[1], values[2], values[3], values[4])
	case "ItemListing":
		_, err = dbHandler.DB.Exec("INSERT INTO my_db."+dbTable+" VALUES (?,?,?,?,?,?,?,?,?,?)", values[0], values[1], values[2], values[3], values[4], values[5], values[6], values[7], values[8], values[9])
		// err := results.Scan(&data1.ID, &data1.Username, &data1.Name, &data1.ImageLink, &data1.DatePosted, &data1.CommentItem, &data1.ConditionItem, &data1.Cat, &data1.ContactMeetInfo)
	case "CommentUser":
		_, err = dbHandler.DB.Exec("INSERT INTO my_db."+dbTable+" VALUES (?,?,?,?,?)", values[0], values[1], values[2], values[3], values[4])
	case "CommentItem":
		_, err = dbHandler.DB.Exec("INSERT INTO my_db."+dbTable+" VALUES (?,?,?,?,?)", values[0], values[1], values[2], values[3], values[4])
	default:
		return errors.New(dbTable + " not found in switch")
	}
	return err
}

// get the current max ID in the server
func (dbHandler dbHandler) getMaxID(dbTable string) (int, error) {
	results, err := dbHandler.DB.Query("SELECT MAX(ID) FROM my_db." + dbTable)
	results.Next()
	var maxID int
	results.Scan(&maxID)
	return maxID, err
}

/*
UserSecret  (ID , Username, Password, IsAdmin, CommentItem);
UserInfo    (ID , Username, LastLogin, DateJoin, CommentItem);
ItemListing (ID , Username, Name, ImageLink, DatePosted INT, CommentItem, ConditionItem, Cat, ContactMeetInfo);
CommentUser (ID , Username, ForUsername, Date, CommentItem);
CommentItem (ID , Username, ForItem, Date, CommentItem);
*/

// edit a single record on DB, chosen record based on ID
func (dbHandler dbHandler) editRecord(dbTable string, values ...interface{}) error {
	var err error
	switch dbTable {
	case "UserSecret":
		aa, err := dbHandler.DB.Exec("UPDATE "+dbTable+" SET Username=?, Password=?, IsAdmin=?, CommentItem=? WHERE ID=?", values[0], values[1], values[2], values[3], values[4])
		fmt.Println(aa, values, err)
	case "UserInfo":
		_, err = dbHandler.DB.Exec("UPDATE "+dbTable+" SET Username=?, LastLogin=?, DateJoin=?, CommentItem=? WHERE ID=?", values[0], values[1], values[2], values[3], values[4])
	case "ItemListing":
		_, err = dbHandler.DB.Exec("UPDATE "+dbTable+" SET Username=?, Name=?, ImageLink=?, DatePosted=?, CommentItem=?, ConditionItem=?, Cat=?, ContactMeetInfo=?, Completion=? WHERE ID=?", values[0], values[1], values[2], values[3], values[4], values[5], values[6], values[7], values[8], values[9])
		// err := results.Scan( &data1.Username, &data1.Name, &data1.ImageLink, &data1.DatePosted, &data1.CommentItem, &data1.ConditionItem, &data1.Cat, &data1.ContactMeetInfo, &data1.ID,)
	case "CommentUser":
		_, err = dbHandler.DB.Exec("UPDATE "+dbTable+" SET Username=?, ForUsername=?, Date=?, CommentItem=? WHERE ID=?", values[0], values[1], values[2], values[3], values[4])
	case "CommentItem":
		_, err = dbHandler.DB.Exec("UPDATE "+dbTable+" SET Username=?, ForItem=?, Date=?, CommentItem=? WHERE ID=?", values[0], values[1], values[2], values[3], values[4])
	default:
		return errors.New(dbTable + " not found in switch")
	}
	return err
}

// delete a single record, chosen based on provided ID
func (dbHandler dbHandler) deleteRecord(dbTable string, id string) error {
	_, err := dbHandler.DB.Exec("DELETE FROM "+dbTable+" WHERE ID=?", id)
	return err
}
