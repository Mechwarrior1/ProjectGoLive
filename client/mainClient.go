package main

import (
	"client/controller"
	"client/encrypt"
	"client/jwtsession"
	"client/session"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// var (
// 	logger1 logger //logs activities
// 	s       http.Server

// 	// a struct to handle all the server session and user information.

// )

// Init initiates the handler functions, server and logger.
type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func StartServer() (http.Server, *echo.Echo, error) {
	e := echo.New()
	t := &Template{
		templates: template.Must(template.ParseGlob("controller/templates/*.gohtml")),
	}
	e.Renderer = t

	client := &http.Client{}
	sessionMgr := &session.Session{
		MapSession: &map[string]session.SessionStruct{},
		ApiKey:     string(encrypt.DecryptFromFile("secure/apikey")),
		Client:     client,
	}

	searchSession := make(map[string]controller.SearchSession)

	// c1, c2 := loggerGo()
	// logger1 = logger{c1, c2}

	// logger1.logTrace("TRACE", "Server started")
	// key1 = anonFunc() //decrypt api key from file

	jwtWrapper := &jwtsession.JwtWrapper{
		string(encrypt.DecryptFromFile("secure/secretkey.txt")),
		"GoRecycle",
		10,
	}

	e.GET("/logout", func(c echo.Context) error {
		return controller.Logout(c, jwtWrapper, sessionMgr)
	})
	e.GET("/seepost", func(c echo.Context) error {
		return controller.SeePostAll_GET(c, jwtWrapper, sessionMgr, searchSession)
	})

	e.POST("/seepost", func(c echo.Context) error {
		return controller.SeePostAll_POST(c)
	})

	e.POST("/login", func(c echo.Context) error {
		return controller.Login_POST(c, jwtWrapper, sessionMgr)
	})

	e.GET("/login", func(c echo.Context) error {
		return controller.Login_GET(c, jwtWrapper, sessionMgr)
	})

	e.POST("/signup", func(c echo.Context) error {
		return controller.Signup_POST(c, jwtWrapper, sessionMgr)
	})

	e.GET("/signup", func(c echo.Context) error {
		return controller.Signup_GET(c, jwtWrapper, sessionMgr)
	})

	e.POST("/getpost/:id", func(c echo.Context) error {
		return controller.GetPostDetail_POST(c, jwtWrapper, sessionMgr)
	})

	e.GET("/getpost/:id", func(c echo.Context) error {
		return controller.GetPostDetail_GET(c, jwtWrapper, sessionMgr)
	})

	e.GET("/complete/:id", func(c echo.Context) error {
		return controller.PostComplete(c, jwtWrapper, sessionMgr)
	})

	e.GET("/createpost", func(c echo.Context) error {
		return controller.CreatePost_GET(c, jwtWrapper, sessionMgr)
	})

	e.POST("/createpost", func(c echo.Context) error {
		return controller.CreatePost_POST(c, jwtWrapper, sessionMgr)
	})

	e.GET("/editpost/:id", func(c echo.Context) error {
		return controller.EditPost_GET(c, jwtWrapper, sessionMgr)
	})

	e.POST("/editpost/:id", func(c echo.Context) error {
		return controller.EditPost_POST(c, jwtWrapper, sessionMgr)
	})

	// e.GET("/seepostuser/:id", func(c echo.Context) error {
	// 	return controller.SeePostUser(c, jwtWrapper, sessionMgr)
	// })

	e.GET("/user", func(c echo.Context) error {
		return controller.GetUser_GET(c, jwtWrapper, sessionMgr)
	})

	e.POST("/user", func(c echo.Context) error {
		return controller.GetUser_POST(c, jwtWrapper, sessionMgr)
	})

	e.POST("/", func(c echo.Context) error {
		return controller.Index_POST(c, jwtWrapper, sessionMgr)
	})

	e.GET("/", func(c echo.Context) error {
		return controller.Index_GET(c, jwtWrapper, sessionMgr)
	})

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, err=${error}, path=${path}, time=${time_unix}\n",
	}))

	port := "5221"
	fmt.Println("listening at port " + port)
	s := http.Server{Addr: ":" + port, Handler: e}

	go sessionMgr.PruneOldSessions()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	return s, e, nil
}

func main() {
	s, e, _ := StartServer()
	if err := s.ListenAndServeTLS("secure//cert.pem", "secure//key.pem"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
