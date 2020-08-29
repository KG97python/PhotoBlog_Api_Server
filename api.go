package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
}

// RegisterData is the register info submitted by user in front-end
type RegisterData struct {
	Email    string
	UserName string
	FullName string
	Password string
}

func userAPI(w http.ResponseWriter, req *http.Request) {
	u := getUser(userID)

	json.NewEncoder(w).Encode(u)
}

func postAPI(w http.ResponseWriter, req *http.Request) {
	p := getPost(userID)

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
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(login.Email).Scan(&ID, &passwordDb)
	if err != nil {
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
	userID = ID
	loggedIn = true
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
		// read
		bs, err := ioutil.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sID := uuid.NewV4()

		//store in server
		dst, err := os.Create(filepath.Join("userImages/profileImage", sID.String()))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = dst.Write(bs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

	var loggedOut struct {
		Loggedout string
	}
	err := json.NewDecoder(req.Body).Decode(&loggedOut)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	loggedIn = false
	userID = 0
}

func requestLoginStatus(w http.ResponseWriter, req *http.Request) {

	if loggedIn == true {
		json.NewEncoder(w).Encode("true")
	} else {
		json.NewEncoder(w).Encode("false")
	}
}

func uploadBlogPost(w http.ResponseWriter, req *http.Request) {
	f, _, _ := req.FormFile("image")
	sID := uuid.NewV4()
	if f != nil {
		ctx := context.Background()
		client, err := storage.NewClient(ctx, option.WithCredentialsFile("./massive-team-279205-069c46519860.json"))
		if err != nil {
			fmt.Println(err)
		}
		bh := client.Bucket("massive-team-279205.appspot.com")
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
	saveImage(userID, content, dt.Format("2006-01-02 15:04:05"), sID.String())
}

func postUserComment(w http.ResponseWriter, req *http.Request) {

	comment := req.FormValue("userComment")
	iD := req.FormValue("ID")
	blogID, _ := strconv.Atoi(iD)
	dt := time.Now()
	saveComment(userID, comment, dt.Format("2006-01-02 15:04:05"), blogID)
}

func updateUserName(w http.ResponseWriter, req *http.Request) {

	newUsername := req.FormValue("newUsername")
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

	u := getUser(userID) // current info

	f, _, err := req.FormFile("ProfilePic")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	// read
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sID := uuid.NewV4() // new pic name
	dst, err := os.Create(filepath.Join("userImages/profileImage", sID.String()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = dst.Write(bs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
