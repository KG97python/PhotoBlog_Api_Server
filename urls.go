package main

import (
	"net/http"
)

func urls() {
	r.HandleFunc("/", home)
	r.PathPrefix("/userImages/").Handler(http.StripPrefix("/userImages/", http.FileServer(http.Dir("/userImages"))))
	r.HandleFunc("/api/userAPI", userAPI)
	r.HandleFunc("/api/postAPI", postAPI)
	r.HandleFunc("/api/feedAPI", feedAPI)

	r.HandleFunc("/api/{blogID}", singleBlogAPI)
	r.HandleFunc("/api/{blogID}/comments", blogIDAPI)

	r.HandleFunc("/api/request/LoginData", requestLoginData)
	r.HandleFunc("/api/request/RegisterData", requestRegisterData)
	r.HandleFunc("/api/request/LogOut", requestLogout)
	r.HandleFunc("/api/request/LoginStatus", requestLoginStatus)
	r.HandleFunc("/api/request/UploadBlog", uploadBlogPost)
	r.HandleFunc("/api/request/PostComment", postUserComment)

	r.HandleFunc("/api/settings/UpdateUserName", updateUserName)
	r.HandleFunc("/api/settings/UpdateProfilePic", updateProfilePic)
}
