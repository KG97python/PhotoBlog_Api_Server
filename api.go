package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/option"
)

// LoginData is the login info submitted by user in front-end
type LoginData struct {
	Email    string
	Password string
	UUID     string
}

// RegisterData is the register info submitted by user in front-end
type RegisterData struct {
	Email    string
	UserName string
	FullName string
	Password string
}

func userAPI(w http.ResponseWriter, req *http.Request) {
	// obtain user UUID cookie Value
	vars := mux.Vars(req)
	uuid := vars["uuid"]
	// use UUID to obtain user ID
	ID := dbSessions[uuid]

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u := getUser(ID)
	json.NewEncoder(w).Encode(u)
}

func postAPI(w http.ResponseWriter, req *http.Request) {
	// obtain user UUID cookie Value
	vars := mux.Vars(req)
	uuid := vars["uuid"]

	// use UUID to obtain user ID
	ID := dbSessions[uuid]

	p := getPost(ID)
	json.NewEncoder(w).Encode(p)
}

func feedAPI(w http.ResponseWriter, req *http.Request) {
	var blogs blogFeeds
	blogs = getBlogFeed()

	json.NewEncoder(w).Encode(blogs)
}

func blogIDAPI(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	temp := vars["blogID"]
	id, _ := strconv.Atoi(temp)
	comments := getComments(id)

	json.NewEncoder(w).Encode(comments)
}

func singleBlogAPI(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	temp := vars["blogID"]
	blogID, _ := strconv.Atoi(temp)
	blog := getBlogIDPost(blogID)

	json.NewEncoder(w).Encode(blog)
}

func requestLoginData(w http.ResponseWriter, req *http.Request) {
	var (
		ID         int
		passwordDb string
		login      LoginData
	)

	err := json.NewDecoder(req.Body).Decode(&login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	stmt, err := db.Prepare("SELECT ID, password FROM users WHERE email = ?")
	if err != nil {
		fmt.Println(err, "error with email")
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(login.Email).Scan(&ID, &passwordDb)
	if err != nil {
		fmt.Println(err, "error with password")
		http.Error(w, "Email and/or password do not match", http.StatusForbidden)
		return
	}
	// match password with stored password
	err = bcrypt.CompareHashAndPassword([]byte(passwordDb), []byte(login.Password))
	if err != nil {
		http.Error(w, "Email and/or password do not match", http.StatusForbidden)
		return
	}
	// login purposes
	u := getUser(ID)
	dbSessions[login.UUID] = ID
	dbUsers[ID] = u
}

func requestRegisterData(w http.ResponseWriter, req *http.Request) {

	// capture form
	name := req.FormValue("name")
	username := req.FormValue("username")
	email := req.FormValue("email")
	password := req.FormValue("password")

	// check to see if user already has a account. If so, redirect.
	if verifyUser(email) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	}

	// hash and secure the user password before storing in DB
	p, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// capture ProfilePic
	f, _, err := req.FormFile("image")
	// if image is emptly assign default profilePic.
	if err != nil {
		// insert into the database
		stmt, err := db.Prepare(`INSERT INTO users(fullName, userName, email, password, profilePic) VALUES(?,?,?,?,?);`)
		if err != nil {
			println(err)
		}
		defer stmt.Close()
		pic := "default.png"
		_, err = stmt.Exec(name, username, email, p, pic)
		if err != nil {
			log.Fatal(err)
		}
		return
	} else {
		defer f.Close()
		sID := uuid.NewV4()

		// store in MYSQL bucket
		ctx := context.Background()
		client, err := storage.NewClient(ctx, option.WithCredentialsFile("./socialmedia-287916-d8a7458dc360.json"))
		if err != nil {
			fmt.Println(err)
		}
		bh := client.Bucket("socialmedia-287916.appspot.com")
		obj := bh.Object(sID.String())

		wc := obj.NewWriter(ctx)
		io.Copy(wc, f)
		wc.Close()
		if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
			fmt.Println(err)
		}
		defer f.Close()

		// insert into the database
		stmt, err := db.Prepare(`INSERT INTO users(fullName, userName, email, password, profilePic) VALUES(?,?,?,?,?);`)
		if err != nil {
			println(err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(name, username, email, p, sID.String())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func requestLogout(w http.ResponseWriter, req *http.Request) {
	// get uuid value
	type loggedOut struct {
		UUID string
	}
	var logout loggedOut
	err := json.NewDecoder(req.Body).Decode(&logout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// use uuid to get iser ID
	userID := dbSessions[logout.UUID]
	// delete sessions
	delete(dbSessions, logout.UUID)
	delete(dbUsers, userID)
}

func uploadBlogPost(w http.ResponseWriter, req *http.Request) {
	f, _, _ := req.FormFile("image")
	sID := uuid.NewV4()

	//save image on gloud bucket
	if f != nil {
		ctx := context.Background()
		client, err := storage.NewClient(ctx, option.WithCredentialsFile("./socialmedia-287916-d8a7458dc360.json"))
		if err != nil {
			fmt.Println(err)
		}
		bh := client.Bucket("socialmedia-287916.appspot.com")
		obj := bh.Object(sID.String())

		wc := obj.NewWriter(ctx)
		io.Copy(wc, f)
		wc.Close()
		if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
			fmt.Println(err)
		}
	}
	defer f.Close()
	content := req.FormValue("content")
	dt := time.Now()
	// use the uuid to obtain user ID
	ID := req.FormValue("uuid")
	userID := dbSessions[ID]
	saveImage(userID, content, dt.Format("2006-01-02 15:04:05"), sID.String())
}

func postUserComment(w http.ResponseWriter, req *http.Request) {

	comment := req.FormValue("userComment")
	iD := req.FormValue("ID")
	// use the uuid to obtain user ID
	ID := req.FormValue("uuid")
	userID := dbSessions[ID]

	blogID, _ := strconv.Atoi(iD)
	dt := time.Now()
	saveComment(userID, comment, dt.Format("2006-01-02 15:04:05"), blogID)
}

func updateUserName(w http.ResponseWriter, req *http.Request) {

	newUsername := req.FormValue("newUsername")
	// use the uuid to obtain user ID
	ID := req.FormValue("uuid")
	userID := dbSessions[ID]
	u := getUser(userID) // current info

	stmtUsers, _ := db.Prepare(`UPDATE users SET userName=? WHERE userName=?;`)
	defer stmtUsers.Close()

	stmtimages, _ := db.Prepare(`UPDATE userimages SET userName=? WHERE userName=?;`) // find a way to make this more efficiant, instead of 3 queries.
	defer stmtimages.Close()

	stmtcomments, _ := db.Prepare(`UPDATE comments SET userName=? WHERE userName=?;`)
	defer stmtcomments.Close()

	_, _ = stmtUsers.Exec(newUsername, u.UserName)
	_, _ = stmtimages.Exec(newUsername, u.UserName)
	_, _ = stmtcomments.Exec(newUsername, u.UserName)
}

func updateProfilePic(w http.ResponseWriter, req *http.Request) {
	// use the uuid to obtain user ID
	iD := req.FormValue("uuid")
	userID := dbSessions[iD]
	u := getUser(userID) // current info
	sID := uuid.NewV4()  // new pic name

	// get newprofilePic
	f, _, err := req.FormFile("ProfilePic")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//save image on gloud bucket
	if f != nil {
		ctx := context.Background()
		client, err := storage.NewClient(ctx, option.WithCredentialsFile("./socialmedia-287916-d8a7458dc360.json"))
		if err != nil {
			fmt.Println(err)
		}
		bh := client.Bucket("socialmedia-287916.appspot.com")
		obj := bh.Object(sID.String())

		wc := obj.NewWriter(ctx)
		io.Copy(wc, f)
		wc.Close()
		if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
			fmt.Println(err)
		}
	}
	defer f.Close()

	// update the database.
	stmtUsers, _ := db.Prepare(`UPDATE users SET profilePic=? WHERE profilePic=?;`)
	defer stmtUsers.Close()

	stmtimages, _ := db.Prepare(`UPDATE userimages SET profilePic=? WHERE profilePic=?;`) // find a way to make this more efficiant, instead of 3 queries.
	defer stmtimages.Close()

	stmtcomments, _ := db.Prepare(`UPDATE comments SET profilePic=? WHERE profilePic=?;`)
	defer stmtcomments.Close()

	_, _ = stmtUsers.Exec(sID.String(), u.ProfilePic)
	_, _ = stmtimages.Exec(sID.String(), u.ProfilePic)
	_, _ = stmtcomments.Exec(sID.String(), u.ProfilePic)
}
