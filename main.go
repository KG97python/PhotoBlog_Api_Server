package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var db *sql.DB
var err error
var r = mux.NewRouter()

type imagesID []string //slice to append imageIDs to

var images1 imagesID // var for slice type

var dbUsers = make(map[int]user)      // userID , user
var dbSessions = make(map[string]int) // sessionID return userID

type user struct {
	ID         int    `json:"Id"`
	Name       string `json:"Name"`
	UserName   string `json:"UserName"`
	Email      string `json:"Email"`
	ProfilePic string `json:"ProfilePic"`
}

var tpl *template.Template

func init() {
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

	connection = fmt.Sprintf("%s:%s@unix(/cloudsql/%s)/%s", dbUser, dbPwd, instanceConnectionName, dbName)
	db, _ = sql.Open("mysql", connection)

	headers := handlers.AllowedHeaders([]string{"X-Requested-with", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
	origins := handlers.AllowedOrigins([]string{"*"})

	urls()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":"+port, handlers.CORS(headers, methods, origins)(r))
}

func home(w http.ResponseWriter, req *http.Request) {
	tpl.Execute(w, nil)
}
