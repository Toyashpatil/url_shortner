package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/urlshortner/models"
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
	}

	fmt.Print(w, newu.String())
	uidVal := r.Context().Value(models.ContextKeyUserID)
	if uidVal == nil {
		http.Error(w, "no user in context", http.StatusUnauthorized)
		return
	}
	userID, ok := uidVal.(int)
	if !ok {
		http.Error(w, "bad user id", http.StatusInternalServerError)
		return
	}

	var id string
	shortUrl := newu.String()

	if err := json.NewDecoder(r.Body).Decode(&urlIncoming); err != nil {
		http.Error(w, "invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	err := u.DB.QueryRow(
		`INSERT INTO urls(long_url,code,user_id,base_url) VALUES($1,$2,$3,$4) RETURNING code`,
		urlIncoming.LongURL, result, userID, shortUrl,
	).Scan(&id)
	if err != nil {
		fmt.Println("error in adding url" + err.Error())
	}
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Base string
		Code string
	}{
		Base: shortUrl,
		Code: id,
	})
}

func (u *UrlController) GetTheUrl(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	code := vars["code"]

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

func (u *UrlController) GetUsersUrl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uidVal := r.Context().Value(models.ContextKeyUserID)
	if uidVal == nil {
		http.Error(w, "no user in context", http.StatusUnauthorized)
		return
	}
	userID, ok := uidVal.(int)
	if !ok {
		http.Error(w, "bad user id", http.StatusInternalServerError)
		return
	}

	const query = `
        SELECT id, base_url, code, long_url, user_id, created_at
        FROM urls
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := u.DB.QueryContext(r.Context(), query, userID)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var allUrls []models.URL
	for rows.Next() {
		var rec models.URL
		if err := rows.Scan(
			&rec.ID,
			&rec.BaseURL,
			&rec.Code,
			&rec.LongURL,
			&rec.UserID,
			&rec.CreatedAt,
		); err != nil {
			http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		allUrls = append(allUrls, rec)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "row iteration error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(allUrls)
}

func (u *UrlController) DeleteUrl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Auth: get user ID
	uidVal := r.Context().Value(models.ContextKeyUserID)
	if uidVal == nil {
		http.Error(w, "no user in context", http.StatusUnauthorized)
		return
	}
	userID, ok := uidVal.(int)
	if !ok {
		http.Error(w, "invalid user id", http.StatusInternalServerError)
		return
	}

	// 2. Get the URL ID from the path
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		http.Error(w, "missing url id", http.StatusBadRequest)
		return
	}
	urlID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid url id", http.StatusBadRequest)
		return
	}

	// 3. Perform the delete, scoped to this user
	res, err := u.DB.ExecContext(
		r.Context(),
		`DELETE FROM urls
         WHERE id = $1
           AND user_id = $2`,
		urlID, userID,
	)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Check that we actually deleted something
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "no such url for this user", http.StatusNotFound)
		return
	}

	// 5. Return success JSON
	json.NewEncoder(w).Encode(struct {
		Success bool `json:"success"`
	}{Success: true})
}
