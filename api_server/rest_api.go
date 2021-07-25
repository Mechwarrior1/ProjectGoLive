package main

import (
	"apiserver/controller"
	"apiserver/mysql"
	"apiserver/word2vec"
	"crypto/tls"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // go mod init api_server.go
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	embed := word2vec.GetWord2Vec()
	dbHandler1 := mysql.OpenDB()
	defer dbHandler1.DB.Close()

	e := echo.New()
	e.GET("/api/v0/check", func(c echo.Context) error {
		return controller.PwCheck(c, &dbHandler1)
	})
	e.GET("/api/v0/comment/:id", func(c echo.Context) error {
		return controller.GetAllComment(c, &dbHandler1, embed)
	})
	e.GET("/api/v0/listing/", func(c echo.Context) error {
		return controller.GetAllListing(c, &dbHandler1, embed)
	})
	e.GET("/api/v0/username/:username", func(c echo.Context) error {
		return controller.CheckUsername(c, &dbHandler1)
	})
	e.POST("/api/v0/db/info", func(c echo.Context) error {
		return controller.GenInfoPost(c, &dbHandler1)
	})
	e.POST("/api/v0/db/signup", func(c echo.Context) error {
		return controller.Signup(c, &dbHandler1)
	})

	e.PUT("/api/v0/db/completed/:id", func(c echo.Context) error {
		return controller.Completed(c, &dbHandler1)
	})

	e.GET("/api/v0/db/info/", func(c echo.Context) error { //attach query parameters
		return controller.GenInfoGet(c, &dbHandler1)
	})
	// e.DELETE("/api/v0/db/info", func(c echo.Context) error {
	// 	return controller.GenInfoDelete(c, &dbHandler1)
	// })
	e.PUT("/api/v0/db/info", func(c echo.Context) error {
		return controller.GenInfoPut(c, &dbHandler1)
	})

	port := "5555"
	fmt.Println("listening at port " + port)
	s := http.Server{Addr: ":" + port, Handler: e}

	e.Use(middleware.Recover())
	// e.Use(middleware.Logger())

	if err := s.ListenAndServeTLS("secure//cert.pem", "secure//key.pem"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}

}
