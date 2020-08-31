package main

import (
	"fmt"
	"log"
)

type userBlog struct { // Blog belonging to the single User logged In
	ID         int    `json:"PostId"`
	UserName   string `json:"UserName"`
	UserID     int    `json:"UserID"`
	ProfilePic string `json:"ProfilePic"`
	Image      string `json:"PostImage"`
	Content    string `json:"Content"`
	DatePosted string `json:"DatePosted"`
}

type userBlogs []userBlog // slice of all blogs belonging to the single user.

type blogFeed struct { // all blogs belonging to all users in Webpage
	ID         int
	UserName   string
	UserID     int
	ProfilePic string
	Image      string
	Content    string
	DatePosted string
}
type blogFeeds []blogFeed // slice of all blogs in Feed

type blogComment struct { // the struct of all comments under ONE blogPost
	ID         int
	UserName   string
	UserID     int
	ProfilePic string
	Content    string
	DatePosted string
	BlogID     int // the id of the blog the comment is under.
}
type blogComments []blogComment // slice of all comments on a post

func getUser(data int) user {
	var (
		ID         int
		name       string
		userName   string
		email      string
		profilePic string
	)

	stmt, err := db.Prepare("SELECT ID, fullname, userName, email, profilePic FROM users WHERE ID = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(data).Scan(&ID, &name, &userName, &email, &profilePic)
	if err != nil {
		fmt.Println("you must be logged in")
	}
	userLoggedIn := user{
		ID:         ID,
		Name:       name,
		UserName:   userName,
		Email:      email,
		ProfilePic: profilePic,
	}
	return userLoggedIn
}

func verifyUser(email string) bool {
	var name string
	var check bool

	stmt, err := db.Prepare("SELECT email FROM users WHERE email = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(email).Scan(&name)
	if err != nil {
		check = false
	} else {
		check = true
	}
	return check
}

//save image to db
func saveImage(ID int, content, time, imageID string) {
	var (
		userName   string
		userEmail  string
		profilePic string
	)
	stmt, err := db.Prepare(`INSERT INTO userimages(userEmail, userName, userID, profilePic, content, image, datePosted) VALUES (?,?,?,?,?,?,?);`)
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	stmtUser, err := db.Prepare("SELECT email, userName, profilePic FROM users WHERE ID = ?") // userID userName and ProfilePic
	if err != nil {
		log.Fatal(err)
	}
	defer stmtUser.Close()
	err = stmtUser.QueryRow(ID).Scan(&userEmail, &userName, &profilePic)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = stmt.Exec(userEmail, userName, ID, profilePic, content, imageID, time)
	if err != nil {
		log.Fatalln(err)
	}
}

func getPost(data int) userBlogs {
	var (
		idDb         int
		userNameDb   string
		userID       int
		profilePicDb string
		imageDb      string
		contentDb    string
		datePostedDb string
	)
	var dataSlice userBlogs
	// grabing data from dataBase
	stmt, err := db.Prepare("SELECT id, userName, userID, profilePic, content, image, datePosted FROM userimages WHERE userID = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(data)
	if err != nil {
		log.Fatalln(err)
	}

	for rows.Next() { // find the data
		err = rows.Scan(&idDb, &userNameDb, &userID, &profilePicDb, &contentDb, &imageDb, &datePostedDb)
		if err != nil {
			log.Fatalln(err)
		}
		data := userBlog{ // insert Data into a struct for every loop
			ID:         idDb,
			UserName:   userNameDb,
			UserID:     userID,
			ProfilePic: profilePicDb,
			Image:      imageDb,
			Content:    contentDb,
			DatePosted: datePostedDb,
		}
		dataSlice = append(dataSlice, data) // insert into slice on each loop
	}
	return dataSlice
}

func getBlogFeed() blogFeeds {
	var (
		idDb         int
		userNameDb   string
		userID       int
		profilePicDb string
		imageDb      string
		contentDb    string
		datePostedDb string
	)
	var dataSlice blogFeeds
	// grabing data from dataBase
	stmt, err := db.Prepare("SELECT id, userName, userID, profilePic, content, image, datePosted FROM userimages")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Fatalln(err)
	}

	for rows.Next() { // find the data
		err = rows.Scan(&idDb, &userNameDb, &userID, &profilePicDb, &contentDb, &imageDb, &datePostedDb)
		if err != nil {
			log.Fatalln(err)
		}
		data := blogFeed{ // insert Data into a struct for every loop
			ID:         idDb,
			UserName:   userNameDb,
			UserID:     userID,
			ProfilePic: profilePicDb,
			Image:      imageDb,
			Content:    contentDb,
			DatePosted: datePostedDb,
		}
		dataSlice = append(dataSlice, data) // insert into slice on each loop
	}
	return dataSlice
}

func getComments(id int) blogComments {
	var (
		idDb         int
		userNameDb   string
		userID       int
		profilePicDb string
		contentDb    string
		datePostedDb string
		blogID       int
		dataSlice    blogComments
	)
	stmt, err := db.Prepare("SELECT * FROM comments WHERE blogID = ? ")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		log.Fatalln(err)
	}

	for rows.Next() { // find the data
		err = rows.Scan(&idDb, &userNameDb, &userID, &profilePicDb, &contentDb, &datePostedDb, &blogID)
		if err != nil {
			log.Fatalln(err)
		}
		data := blogComment{ // insert Data into a struct for every loop
			ID:         idDb,
			UserName:   userNameDb,
			UserID:     userID,
			ProfilePic: profilePicDb,
			Content:    contentDb,
			DatePosted: datePostedDb,
			BlogID:     blogID,
		}
		dataSlice = append(dataSlice, data) // insert into slice on each loop
	}
	return dataSlice
}

func saveComment(ID int, content string, time string, blogID int) {
	var (
		userName   string
		profilePic string
		userID     int
	)
	stmt, err := db.Prepare(`INSERT INTO comments VALUES (?,?,?,?,?,?,?);`)
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	stmtUser, err := db.Prepare("SELECT ID, userName, profilePic FROM users WHERE ID = ?") // obtain nuserName and ProfilePic
	if err != nil {
		log.Fatal(err)
	}
	defer stmtUser.Close()
	err = stmtUser.QueryRow(ID).Scan(&userID, &userName, &profilePic)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = stmt.Exec(nil, userName, userID, profilePic, content, time, blogID) // insert comment into database
	if err != nil {
		log.Fatalln(err)
	}
}

func getBlogIDPost(data int) userBlog {
	var blog userBlog

	var (
		idDb         int
		userNameDb   string
		userID       int
		profilePicDb string
		imageDb      string
		contentDb    string
		datePostedDb string
	)
	// grabing data from dataBase
	stmt, err := db.Prepare("SELECT id, userName, userID, profilePic, content, image, datePosted FROM userimages WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(data)
	if err != nil {
		log.Fatalln(err)
	}

	for rows.Next() { // find the data
		err = rows.Scan(&idDb, &userNameDb, &userID, &profilePicDb, &contentDb, &imageDb, &datePostedDb)
		if err != nil {
			log.Fatalln(err)
		}
		blog = userBlog{
			ID:         idDb,
			UserName:   userNameDb,
			UserID:     userID,
			ProfilePic: profilePicDb,
			Image:      imageDb,
			Content:    contentDb,
			DatePosted: datePostedDb,
		}
	}
	return blog
}
