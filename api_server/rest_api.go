package main

import (
	"apiserver/controller"
	"apiserver/mysql"
	"apiserver/word2vec"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql" // go mod init api_server.go
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func StartServer() (http.Server, *echo.Echo, *mysql.DBHandler, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	embed := word2vec.GetWord2Vec()
	dbHandler1 := mysql.OpenDB()

	e := echo.New()
	e.GET("/api/v0/check", func(c echo.Context) error {
		return controller.PwCheck(c, &dbHandler1)
	})

	e.GET("/api/v0/comment/:id", func(c echo.Context) error {
		return controller.GetAllComment(c, &dbHandler1, embed)
	})

	e.GET("/api/v0/index", func(c echo.Context) error {
		return controller.GetAllListingIndex(c, &dbHandler1, embed)
	})

	e.GET("/api/v0/listing", func(c echo.Context) error {
		return controller.GetAllListing(c, &dbHandler1)
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

	e.GET("/api/v0/db/info", func(c echo.Context) error {
		return controller.GenInfoGet(c, &dbHandler1)
	})

	e.PUT("/api/v0/db/info", func(c echo.Context) error {
		return controller.GenInfoPut(c, &dbHandler1)
	})

	e.GET("/api/v0/health", controller.HealthCheckLiveness)

	e.GET("/api/v0/ready", func(c echo.Context) error {
		return controller.HealthCheckReadiness(c, &dbHandler1)
	})

	// e.DELETE("/api/v0/db/info", func(c echo.Context) error {
	// 	return controller.GenInfoDelete(c, &dbHandler1)
	// })

	// go routine for checking mysql connection, will update readiness if connected
	go func(dbHandler *mysql.DBHandler) {
		for {
			time.Sleep(10 * time.Second)
			_, err1 := dbHandler1.GetSingleRecord("ItemListing", "WHERE ID = ?", "000001")
			if err1 != nil {
				dbHandler.ReadyForTraffic = false
				fmt.Println("unable to contact mysql server", err1.Error())
			} else {
				dbHandler.ReadyForTraffic = true
			}
		}
	}(&dbHandler1)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !dbHandler1.ReadyForTraffic {
				fmt.Println("API is accessed, but is unable to contact sql server")
				responseJson := mysql.DataPacketSimple{
					"not ready for traffic",
					"false",
				}
				return c.JSON(503, responseJson) // encode to json and send
			}
			return next(c)
		}
	})

	port := "5555"
	fmt.Println("listening at port " + port)
	s := http.Server{Addr: ":" + port, Handler: e}

	e.Use(middleware.Recover())
	// e.Use(middleware.Logger())
	return s, e, &dbHandler1, nil
}

func main() {
	s, e, dbHandler1, _ := StartServer()
	defer dbHandler1.DB.Close()
	if err := s.ListenAndServeTLS("secure//cert.pem", "secure//key.pem"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
