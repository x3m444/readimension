package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kyicy/readimension/model"
	"github.com/kyicy/readimension/utility/config"
	"github.com/labstack/echo-contrib/session"
	"github.com/michaeljs1990/sqlitestore"
	"golang.org/x/crypto/acme/autocert"
)

func parseFlag() string {
	var environment string
	flag.StringVar(&environment, "env", "production", "running environment")
	flag.Parse()
	return environment
}

func main() {
	// book folder
	os.MkdirAll("books", 0777)

	// cover folder
	os.MkdirAll("covers", 0777)

	// upload folder
	os.MkdirAll("uploads", 0777)

	env := parseFlag()

	file, err := os.Open("config.json")
	checkError(err)

	bytes, err := ioutil.ReadAll(file)
	checkError(err)

	json.Unmarshal(bytes, &config.Configuratiosn)
	envConfig := config.Configuratiosn[env]
	config.SetENV(env)

	// Redis Session Store
	sessionStore, err := sqlitestore.NewSqliteStore("readimension.db", "sessions", "/", 3600*24*365, []byte(envConfig.SessionSecret))
	checkError(err)
	defer sessionStore.Close()

	// Mysql and Model

	db, err := gorm.Open("sqlite3", "readimension.db")
	defer db.Close()
	if env != "production" {
		db.LogMode(true)
	}
	defer db.Close()
	model.LoadModel(db)

	// Create Echo Server Instance
	e := createInstance(env)
	e.Use(session.Middleware(sessionStore))

	// Start the Server
	addr := fmt.Sprintf("%s:%s", envConfig.Addr, envConfig.Port)

	if env == "production" {
		if envConfig.Addr != "localhost" && envConfig.Addr != "" && envConfig.Addr != "127.0.0.1" {
			e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(envConfig.Addr)
			e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
			e.Logger.Fatal(e.StartAutoTLS(addr))
		}
	} else {
		e.Logger.Fatal(e.Start(addr))
	}

}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
