package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"regexp"
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

var (
	key1       = anonFunc() //decrypt api key from file
	splitText  = regexp.MustCompile(`\s*,\s*|\s,*\s*`)
	stopWords2 = regexp.MustCompile("^(i|me|my|myself|we|our|ours|ourselves|you|your|yours|yourself|yourselves|he|him|his|himself|she|her|hers|herself|it|its|itself|they|them|their|theirs|themselves|what|which|who|whom|this|that|these|those|am|is|are|was|were|be|been|being|have|has|had|having|do|does|did|doing|a|an|the|and|but|if|or|because|as|until|while|of|at|by|for|with|about|against|between|into|through|during|before|after|above|below|to|from|up|down|in|out|on|off|over|under|again|values|further|then|once|here|there|when|where|why|how|all|any|both|each|few|more|most|other|some|such|no|nor|not|only|own|same|so|than|too|very|s|t|can|will|just|don|should|now|0o|0s|3a|3b|3d|6b|6o|a|a1|a2|a3|a4|ab|able|about|above|abst|ac|accordance|according|accordingly|across|act|actually|ad|added|adj|ae|af|affected|affecting|affects|after|afterwards|ag|again|against|ah|ain|ain't|aj|al|all|allow|allows|almost|alone|along|already|also|although|always|am|among|amongst|amoungst|amount|an|and|announce|another|any|anybody|anyhow|anymore|anyone|anything|anyway|anyways|anywhere|ao|ap|apart|apparently|appear|appreciate|appropriate|approximately|ar|are|aren|arent|aren't|arise|around|as|a's|aside|ask|asking|associated|at|au|auth|av|available|aw|away|awfully|ax|ay|az|b|b1|b2|b3|ba|back|bc|bd|be|became|because|become|becomes|becoming|been|before|beforehand|begin|beginning|beginnings|begins|behind|being|believe|below|beside|besides|best|better|between|beyond|bi|bill|biol|bj|bk|bl|bn|both|bottom|bp|br|brief|briefly|bs|bt|bu|but|bx|by|c|c1|c2|c3|ca|call|came|can|cannot|cant|can't|cause|causes|cc|cd|ce|certain|certainly|cf|cg|ch|changes|ci|cit|cj|cl|clearly|cm|c'mon|cn|co|com|come|comes|con|concerning|consequently|consider|considering|contain|containing|contains|corresponding|could|couldn|couldnt|couldn't|course|cp|cq|cr|cry|cs|c's|ct|cu|currently|cv|cx|cy|cz|d|d2|da|date|dc|dd|de|definitely|describe|described|despite|detail|df|di|did|didn|didn't|different|dj|dk|dl|do|does|doesn|doesn't|doing|don|done|don't|down|downwards|dp|dr|ds|dt|du|due|during|dx|dy|e|e2|e3|ea|each|ec|ed|edu|ee|ef|effect|eg|ei|eight|eighty|either|ej|el|eleven|else|elsewhere|em|empty|en|end|ending|enough|entirely|eo|ep|eq|er|es|especially|est|et|et-al|etc|eu|ev|even|ever|every|everybody|everyone|everything|everywhere|ex|exactly|example|except|ey|f|f2|fa|far|fc|few|ff|fi|fifteen|fifth|fify|fill|find|fire|first|five|fix|fj|fl|fn|fo|followed|following|follows|for|former|formerly|forth|forty|found|four|fr|from|front|fs|ft|fu|full|further|furthermore|fy|g|ga|gave|ge|get|gets|getting|gi|give|given|gives|giving|gj|gl|go|goes|going|gone|got|gotten|gr|greetings|gs|gy|h|h2|h3|had|hadn|hadn't|happens|hardly|has|hasn|hasnt|hasn't|have|haven|haven't|having|he|hed|he'd|he'll|hello|help|hence|her|here|hereafter|hereby|herein|heres|here's|hereupon|hers|herself|hes|he's|hh|hi|hid|him|himself|his|hither|hj|ho|home|hopefully|how|howbeit|however|how's|hr|hs|http|hu|hundred|hy|i|i2|i3|i4|i6|i7|i8|ia|ib|ibid|ic|id|i'd|ie|if|ig|ignored|ih|ii|ij|il|i'll|im|i'm|immediate|immediately|importance|important|in|inasmuch|inc|indeed|index|indicate|indicated|indicates|information|inner|insofar|instead|interest|into|invention|inward|io|ip|iq|ir|is|isn|isn't|it|itd|it'd|it'll|its|it's|itself|iv|i've|ix|iy|iz|j|jj|jr|js|jt|ju|just|k|ke|keep|keeps|kept|kg|kj|km|know|known|knows|ko|l|l2|la|largely|last|lately|later|latter|latterly|lb|lc|le|least|les|less|lest|let|lets|let's|lf|like|liked|likely|line|little|lj|ll|ll|ln|lo|look|looking|looks|los|lr|ls|lt|ltd|m|m2|ma|made|mainly|make|makes|many|may|maybe|me|mean|means|meantime|meanwhile|merely|mg|might|mightn|mightn't|mill|million|mine|miss|ml|mn|mo|more|moreover|most|mostly|move|mr|mrs|ms|mt|mu|much|mug|must|mustn|mustn't|my|myself|n|n2|na|name|namely|nay|nc|nd|ne|near|nearly|necessarily|necessary|need|needn|needn't|needs|neither|never|nevertheless|new|next|ng|ni|nine|ninety|nj|nl|nn|no|nobody|non|none|nonetheless|noone|nor|normally|nos|not|noted|nothing|novel|nowhere|nr|ns|nt|ny|o|oa|ob|obtain|obtained|obviously|oc|od|of|off|often|og|oh|oi|oj|ok|okay|ol|old|om|omitted|on|once|one|ones|only|onto|oo|op|oq|or|ord|os|ot|other|others|otherwise|ou|ought|our|ours|ourselves|out|outside|over|overall|ow|owing|own|ox|oz|p|p1|p2|p3|page|pagecount|pages|par|part|particular|particularly|pas|past|pc|pd|pe|per|perhaps|pf|ph|pi|pj|pk|pl|placed|please|plus|pm|pn|po|poorly|possible|possibly|potentially|pp|pq|pr|predominantly|present|presumably|previously|primarily|probably|promptly|proud|provides|ps|pt|pu|put|py|q|qj|qu|que|quickly|quite|qv|r|r2|ra|ran|rather|rc|rd|re|readily|really|reasonably|recent|recently|ref|refs|regarding|regardless|regards|related|relatively|research|research-articl|respectively|resulted|resulting|results|rf|rh|ri|right|rj|rl|rm|rn|ro|rq|rr|rs|rt|ru|run|rv|ry|s|s2|sa|said|same|saw|say|saying|says|sc|sd|se|sec|second|secondly|section|see|seeing|seem|seemed|seeming|seems|seen|self|selves|sensible|sent|serious|seriously|seven|several|sf|shall|shan|shan't|she|shed|she'd|she'll|shes|she's|should|shouldn|shouldn't|should've|show|showed|shown|showns|shows|si|side|significant|significantly|similar|similarly|since|sincere|six|sixty|sj|sl|slightly|sm|sn|so|some|somebody|somehow|someone|somethan|something|sometime|sometimes|somewhat|somewhere|soon|sorry|sp|specifically|specified|specify|specifying|sq|sr|ss|st|still|stop|strongly|sub|substantially|successfully|such|sufficiently|suggest|sup|sure|sy|system|sz|t|t1|t2|t3|take|taken|taking|tb|tc|td|te|tell|ten|tends|tf|th|than|thank|thanks|thanx|that|that'll|thats|that's|that've|the|their|theirs|them|themselves|then|thence|there|thereafter|thereby|thered|therefore|therein|there'll|thereof|therere|theres|there's|thereto|thereupon|there've|these|they|theyd|they'd|they'll|theyre|they're|they've|thickv|thin|think|third|this|thorough|thoroughly|those|thou|though|thoughh|thousand|three|throug|through|throughout|thru|thus|ti|til|tip|tj|tl|tm|tn|to|together|too|took|top|toward|towards|tp|tq|tr|tried|tries|truly|try|trying|ts|t's|tt|tv|twelve|twenty|twice|two|tx|u|u201d|ue|ui|uj|uk|um|un|under|unfortunately|unless|unlike|unlikely|until|unto|uo|up|upon|ups|ur|us|use|used|useful|usefully|usefulness|uses|using|usually|ut|v|va|value|various|vd|ve|ve|very|via|viz|vj|vo|vol|vols|volumtype|vq|vs|vt|vu|w|wa|want|wants|was|wasn|wasnt|wasn't|way|we|wed|we'd|welcome|well|we'll|well-b|went|were|we're|weren|werent|weren't|we've|what|whatever|what'll|whats|what's|when|whence|whenever|when's|where|whereafter|whereas|whereby|wherein|wheres|where's|whereupon|wherever|whether|which|while|whim|whither|who|whod|whoever|whole|who'll|whom|whomever|whos|who's|whose|why|why's|wi|widely|will|willing|wish|with|within|without|wo|won|wonder|wont|won't|words|world|would|wouldn|wouldnt|wouldn't|www|x|x1|x2|x3|xf|xi|xj|xk|xl|xn|xo|xs|xt|xv|xx|y|y2|yes|yet|yj|yl|you|youd|you'd|you'll|your|youre|you're|yours|yourself|yourselves|you've|yr|ys|yt|z|zero|zi|zz)$")
)

func anonFunc() func() string {
	key1 := string(decryptFromFile("secure/apikey", "secure/apikey.xml"))
	return func() string {
		return key1
	}
}

// Opens db and returns a struct to access it
func openDB() dbHandler {
	pass := string(decryptFromFile("secure/mysql", "secure/keys.xml"))
	db, err := sql.Open("mysql", "myuser:"+pass+"@tcp(127.0.0.1:60575)/my_db")
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("no issue")
	}

	dbHandler1 := dbHandler{db}
	return dbHandler1
}

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

// removes stop words and split into array of words
func cleanWord(input1 string, splitText *regexp.Regexp, stopWords *regexp.Regexp) []string {
	newArr := []string{}

	for _, word1 := range splitText.Split(input1, -1) {
		aa := stopWords.Match([]byte(word1))
		fmt.Println(word1, aa)

		if !aa {
			newArr = append(newArr, word1)
		}
	}
	return newArr
}

// access the DB and get all records
func (dbHandler dbHandler) getRecordlisting(dbTable string, requestWords string, filterUsername string) ([]interface{}, error) {
	// allData := []genData{}
	allData := []interface{}{}
	requestWords2 := cleanWord(requestWords, splitText, stopWords2) //clean and split the words for embeding
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
		if filterUsername == "" {
			if data1.Completion != "true" {
				requestWordsEmbed2 := embed.getWordEmbeddingCombine(strings.Fields(data1.Name), []string{})
				addVal := float32(0)
				addVal2 := float32(0)

				for _, word := range requestWords2 {
					if strings.Contains(data1.Name, word) {
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
		} else {
			if data1.Username == filterUsername {
				allData = append(allData, data1)
			}
		}
	}
	return allData, nil
}

// access the DB and get a single record, search using courseName
func (dbHandler dbHandler) getSingleRecord(dbTable string, queryString string) ([]interface{}, error) {
	//queryString examples, " WHERE ID = 1" or "WHERE Username = alvin"
	allData := make([]interface{}, 0)
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

// post a record into the DB
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

// edit a single record on DB, chosen record based on ID
func (dbHandler dbHandler) editRecord(dbTable string, values ...interface{}) error {
	var err error
	switch dbTable {
	// case "UserSecret":
	// 	_, err := dbHandler.DB.Exec("UPDATE "+dbTable+" SET Username=?, Password=?, IsAdmin=?, CommentItem=? WHERE ID=?", values[0], values[1], values[2], values[3], values[4])
	case "UserInfo":
		_, err = dbHandler.DB.Exec("UPDATE "+dbTable+" SET Username=?, LastLogin=?, DateJoin=?, CommentItem=? WHERE ID=?", values[0], values[1], values[2], values[3], values[4])
	case "ItemListing":
		fmt.Println("editrecord: ItemListing")
		_, err = dbHandler.DB.Exec("UPDATE "+dbTable+" SET ImageLink=?, CommentItem=?, ConditionItem=?, Cat=?, ContactMeetInfo=?, Completion=? WHERE ID=?", values[0], values[1], values[2], values[3], values[4], values[5], values[6])
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
