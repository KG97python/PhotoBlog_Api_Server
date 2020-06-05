package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"html/template"
	"os"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB
var err error
var r = mux.NewRouter()

var loggedIn bool // check to users are logged in

type imagesID []string //slice to append imageIDs to

var images1 imagesID // var for slice type

var userID int // keep track of who is logged In

type user struct {
	ID         int    `json:"Id"`
	Name       string `json:"Name"`
	UserName   string `json:"UserName"`
	Email      string `json:"Email"`
	ProfilePic string `json:"ProfilePic"`
}

var dbUsers = map[string]user{} // Email, userData
var tpl *template.Template

func init() {
	loggedIn = false
	tpl = template.Must(template.ParseFiles("home.html"))
}

func main() {
	var (
        dbUser                 = os.Getenv("DB_USER")
        dbPwd                  = os.Getenv("DB_PASS")
        instanceConnectionName = os.Getenv("INSTANCE_CONNECTION_NAME")
        dbName                 = os.Getenv("DB_NAME")
	)
	var connection string

	connection = fmt.Sprintf("%s:%s@unix(/cloudsql/%s)/%s", dbUser, dbPwd , instanceConnectionName, dbName)
	db, _ = sql.Open("mysql", connection)

	urls()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":" + port, r)
}

func home(w http.ResponseWriter, req *http.Request) {
	tpl.Execute(w, nil)
}