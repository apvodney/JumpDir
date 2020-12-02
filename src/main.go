package main

import (
	"github.com/apvodney/JumpDir/api"
	"github.com/apvodney/JumpDir/debug"

	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	Api *api.Api
)

func handleAPI(w http.ResponseWriter, r *http.Request) {
	postHandler := map[string]http.HandlerFunc{
		"startReg": func(w http.ResponseWriter, r *http.Request) {
			username := r.PostForm.Get("username")
			email := r.PostForm.Get("email")
			password := r.PostForm.Get("password")
			secret, err := Api.Copy().StartReg(r.Context(), username, email, password)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			fmt.Fprint(w, base64.URLEncoding.EncodeToString(secret))
		},
		"finishReg": func(w http.ResponseWriter, r *http.Request) {
			secret, err := base64.URLEncoding.DecodeString(r.PostForm.Get("secret"))
			if err != nil {
				http.Error(w, "Error parsing secret base64", http.StatusUnprocessableEntity)
				return
			}
			email := r.PostForm.Get("email")
			userID, err := Api.Copy().FinishReg(r.Context(), secret, email)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			fmt.Fprint(w, base64.URLEncoding.EncodeToString(userID))
		},
		"getToken": func(w http.ResponseWriter, r *http.Request) {
			username := r.PostForm.Get("username")
			password := r.PostForm.Get("password")
			token, err := Api.Copy().GetToken(r.Context(), username, password)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			fmt.Fprint(w, base64.URLEncoding.EncodeToString(token))
		},
	}

	var hf http.HandlerFunc
	switch r.Method {
	case "POST":
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Malformed form data", http.StatusBadRequest)
			return
		}

		var ok bool
		hf, ok = postHandler[r.URL.Path]
		if !ok {
			http.Error(w, "API endpoint does not exist", http.StatusNotFound)
			fmt.Print("invalid endpoint:", r.URL.Path)
			return
		}
	}
	hf(w, r)
}

func frontpage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "Frontpage\n")
}

func main() {
	var err error
	err, Api = api.Initialize()
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", frontpage)
	fmt.Println(debug.True)
	if debug.True {
		http.Handle("/api/", http.StripPrefix("/api/", http.HandlerFunc(handleAPI)))
		fmt.Println("api registered.")
	}
	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      http.DefaultServeMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
