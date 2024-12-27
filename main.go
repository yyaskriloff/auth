package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

type AuthorizationCodeFlow struct {
	ClientID     string
	RedirectURI  string
	ResponseType string
	Scope        string
	State        string
	IP           string
}

var session *cache.Cache

// var usedCodes = make([]string, 0)

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
	scope := queryParams.Get("scope")
	state := queryParams.Get("state")

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

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	authInfo := AuthorizationCodeFlow{
		ClientID:     client_id,
		RedirectURI:  redirect_uri,
		ResponseType: response_type,
		Scope:        scope,
		State:        state,
		IP:           ip,
	}

	secret := uuid.New().String()
	if secret == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}
	session.Set(secret, authInfo, cache.DefaultExpiration)

	redirect, err := url.Parse("/login")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}
	q := redirect.Query()
	q.Set("sid", secret)

	redirect.RawQuery = q.Encode()
	http.Redirect(w, r, redirect.String(), http.StatusFound)

}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		file, err := os.Open("./views/index.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
			return
		}
		defer file.Close()

		fi, err := file.Stat()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
			return
		}

		http.ServeContent(w, r, file.Name(), fi.ModTime(), file)
		return
	}

	u, err := url.Parse(r.URL.String())
	if err != nil {
		panic(err)
	}

	queryParams := u.Query()
	sid := queryParams.Get("sid")

	fmt.Println("sid", sid)

	authInfo, found := session.Get(sid)
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
		return
	}

	verified, ok := authInfo.(AuthorizationCodeFlow)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	fmt.Println("verified", verified)

	urlWithAccessToken, err := url.Parse(verified.RedirectURI)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	q := urlWithAccessToken.Query()
	q.Set("code", "123456")

	urlWithAccessToken.RawQuery = q.Encode()

	http.Redirect(w, r, urlWithAccessToken.String(), http.StatusFound)

}

func main() {

	session = cache.New(5*time.Minute, cache.NoExpiration)
	mux := http.NewServeMux()

	mux.HandleFunc("/auth", handleAuth)
	mux.HandleFunc("/login", handleLogin)

	http.ListenAndServe(":8080", mux)

}
