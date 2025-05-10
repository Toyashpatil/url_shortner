package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"path"

	"github.com/gorilla/mux"
)

type UrlController struct {
	DB *sql.DB
}

func (u *UrlController) ShortTheUrl(w http.ResponseWriter, r *http.Request) {

	var urlIncoming struct {
		LongURL string `json:"long"`
	}

	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	randomLetter := string(letters[rand.Intn(len(letters))])

	randomNumber := rand.Intn(900) + 100

	result := fmt.Sprintf("%s%d", randomLetter, randomNumber)
	// shortURL := "http://localhost:3000/" + result

	scheme := "http"
	if r.TLS != nil { // if you’re on HTTPS
		scheme = "https"
	}

	newu := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   path.Dir(r.URL.Path) + "/", // always leaves trailing “/”
	}

	fmt.Print(w, newu.String())

	var id string

	if err := json.NewDecoder(r.Body).Decode(&urlIncoming); err != nil {
		http.Error(w, "invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	err := u.DB.QueryRow(
		`INSERT INTO urls(long_url,code) VALUES($1,$2) RETURNING code`,
		urlIncoming.LongURL, result,
	).Scan(&id)
	if err != nil {
		fmt.Println("error in adding url" + err.Error())
	}
	defer r.Body.Close()
	shortUrl := newu.String() + id

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shortUrl)
}

func (u *UrlController) GetTheUrl(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	code := vars["code"] // extract the slug from URL

	var longURL string
	err := u.DB.
		QueryRow("SELECT long_url FROM urls WHERE code = $1", code).
		Scan(&longURL)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, longURL, http.StatusFound)
}
