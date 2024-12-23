package main

import (
	"fmt"
	"net/http"
	"net/url"
)

var usedCodes = make([]string, 0)

const callbackURL = "http://localhost:3000/callback"

func handleAuth(w http.ResponseWriter, r *http.Request) {

	u, err := url.Parse(r.URL.String())
	if err != nil {
		panic(err)
	}

	queryParams := u.Query()
	client_id := queryParams.Get("client_id")
	redirect_uri := queryParams.Get("redirect_uri")
	response_type := queryParams.Get("response_type")
	// scope := queryParams.Get("scope")
	// state := queryParams.Get("state")

	if client_id == "" || redirect_uri == "" || response_type == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
		return
	}

	if redirect_uri != callbackURL {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad redirect URI"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Let's authenticate"))

}

func main() {

	usedCodes = append(usedCodes, "hello-world")
	fmt.Println(usedCodes)
	mux := http.NewServeMux()

	mux.HandleFunc("/auth", handleAuth)

	http.ListenAndServe(":8080", mux)

}
