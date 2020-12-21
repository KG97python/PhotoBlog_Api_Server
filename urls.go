package main

import (
	"net/http"
)

func urls() {
	r.HandleFunc("/", home)
	r.PathPrefix("/userImages/").Handler(http.StripPrefix("/userImages/", http.FileServer(http.Dir("/userImages"))))
	r.HandleFunc("/api/userAPI/{uuid}", userAPI).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/postAPI/{uuid}", postAPI).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/feedAPI", feedAPI).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/{blogID}", singleBlogAPI).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/{blogID}/comments", blogIDAPI).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/request/LoginData", requestLoginData).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/request/RegisterData", requestRegisterData).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/request/LogOut", requestLogout).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/request/UploadBlog", uploadBlogPost).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/request/PostComment", postUserComment).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/settings/UpdateUserName", updateUserName).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/settings/UpdateProfilePic", updateProfilePic).Methods("POST", "OPTIONS")
}
