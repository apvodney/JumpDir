package main

import (
	"github.com/apvodney/JumpDir/api"
	"github.com/apvodney/JumpDir/debug"

	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	Api *api.Api
)

func handleAPI(w http.ResponseWriter, r *http.Request) {
	postHandler := map[string]http.HandlerFunc{
		"reg": func(w http.ResponseWriter, r *http.Request) {
			email := r.PostForm.Get("email")
			password := r.PostForm.Get("password")
			err, secret := Api.StartReg(r.Context(), email, password)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			fmt.Fprint(w, base64.URLEncoding.EncodeToString(secret))
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
		hf, ok = postHandler[strings.TrimLeft(r.URL.Path, "/api/")]
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
		http.HandleFunc("/api/", handleAPI)
		fmt.Println("api registered.")
	}
	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      http.TimeoutHandler(http.DefaultServeMux, 10*time.Second, "timeout"),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
