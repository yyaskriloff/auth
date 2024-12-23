package main

import (
	"net"
	"net/http"
	"net/url"
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

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Let's authenticate"))

}

func main() {

	usedCodes = append(usedCodes, "hello-world")
	session = cache.New(5*time.Minute, cache.NoExpiration)
	usedCodes = append(usedCodes, "hello-world")
	mux := http.NewServeMux()

	mux.HandleFunc("/auth", handleAuth)

	http.ListenAndServe(":8080", mux)

}
