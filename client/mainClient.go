package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

var (
	logger1 logger //logs activities
	s       http.Server
	key1    func() string

	tplIndex = template.Must(template.New("").ParseFiles("templates/header.gohtml", "templates/index.gohtml"))
	// tplDeleteCourse = template.Must(template.New("").ParseFiles("templates/deletecourse.gohtml", "templates/header.gohtml", "templates/footer.gohtml"))
	tplSeePostAll    = template.Must(template.New("").ParseFiles("templates/seepostall.gohtml", "templates/header.gohtml", "templates/footer.gohtml"))
	tplLogin         = template.Must(template.New("").ParseFiles("templates/login.gohtml", "templates/header.gohtml", "templates/footer.gohtml"))
	tplSignup        = template.Must(template.New("").ParseFiles("templates/signup.gohtml", "templates/header.gohtml", "templates/footer.gohtml"))
	tplGetPostDetail = template.Must(template.New("").ParseFiles("templates/getpostdetail.gohtml", "templates/header.gohtml", "templates/footer.gohtml"))
	tplCreatePost    = template.Must(template.New("").ParseFiles("templates/createpost.gohtml", "templates/header.gohtml", "templates/footer.gohtml"))
	tplEditPost      = template.Must(template.New("").ParseFiles("templates/editpost.gohtml", "templates/header.gohtml", "templates/footer.gohtml"))
	tplUpdateUser    = template.Must(template.New("").ParseFiles("templates/updateuser.gohtml", "templates/header.gohtml", "templates/footer.gohtml"))
	tplSeePostUser   = template.Must(template.New("").ParseFiles("templates/seepostuser.gohtml", "templates/header.gohtml", "templates/footer.gohtml"))
	// tplShutdown         = template.Must(template.New("").ParseFiles("templates/shutdown.gohtml", "templates/header.gohtml", "templates/footer.gohtml"))

	// a struct to handle all the server session and user information.
	sessionMgr = &sessionManager{
		mapPersistInfo: &map[string]*userPersistInfo{},
		mapSession:     &map[string]*sessionStruct{},
	}
)

// Init initiates the handler functions, server and logger.
func init() {
	// logger function
	c1, c2 := loggerGo()
	logger1 = logger{c1, c2}
	logger1.logTrace("TRACE", "Server started")
	key1 = anonFunc() //decrypt api key from file

	router := mux.NewRouter()
	router.HandleFunc("/logout", logout)
	router.Handle("/favicon.ico", http.NotFoundHandler())
	router.HandleFunc("/seepost", seePostAll)
	router.HandleFunc("/login", login)
	router.HandleFunc("/signup", signup)
	router.HandleFunc("/getpost/{id}", getPostDetail)
	router.HandleFunc("/complete/{id}", postComplete)
	router.HandleFunc("/createpost", createPost)
	router.HandleFunc("/editpost/{id}", editPost)
	router.HandleFunc("/seepostuser/{id}", seePostUser)
	// router.HandleFunc("/shutdown", shutdown)
	router.HandleFunc("/user", getUser)
	router.HandleFunc("/", index)

	s = http.Server{Addr: ":5221", Handler: router}

	go sessionMgr.pruneOldSessions()
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if err := s.ListenAndServeTLS("secure//cert.pem", "secure//key.pem"); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func anonFunc() func() string {
	key1 := string(decryptFromFile("secure/apikey"))
	return func() string {
		return key1
	}
}
