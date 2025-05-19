package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/urlshortner/models"
	"github.com/urlshortner/utils"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	DB *sql.DB
}

func (g *UserController) LoginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if in.Email == "" || in.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	var (
		userID       int
		passwordHash string
	)
	query := `
        SELECT id, password_hash
        FROM users
        WHERE email = $1
    `
	err := g.DB.
		QueryRowContext(r.Context(), query, in.Email).
		Scan(&userID, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(in.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := utils.GetJwtToken(strconv.Itoa(userID))
	if err != nil {
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}

	type resOut struct {
		Success bool   `json:"success"`
		Token   string `json:"token"`
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resOut{
		Success: true,
		Token:   token,
	})
}

func (g *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var in struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if in.Email == "" || in.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "could not hash password", http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
	var newID int
	query := `
		INSERT INTO users (name,email, password_hash, created_at)
		VALUES ($1, $2, $3,$4)
		RETURNING id`
	if err := g.DB.
		QueryRowContext(r.Context(), query, in.Name, in.Email, string(hash), now).
		Scan(&newID); err != nil {

		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			http.Error(w, "email already registered", http.StatusConflict)
			return
		}
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userToken, err := utils.GetJwtToken(strconv.Itoa(newID))
	type resOut struct {
		Success bool   `json:"success"`
		Token   string `json:"token"`
	}

	resp := resOut{
		Success: true,
		Token:   userToken,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (uc *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Pull the id from context (set by Auth middleware)
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

	// 2. Query DB
	var user models.User
	err := uc.DB.QueryRowContext(
		r.Context(),
		`SELECT id, email, created_at FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Email, &user.CreatedAt)

	switch err {
	case nil: // ok
	case sql.ErrNoRows:
		http.Error(w, "user not found", http.StatusNotFound)
		return
	default:
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Respond
	json.NewEncoder(w).Encode(struct {
		Success bool        `json:"success"`
		User    models.User `json:"user"`
	}{
		Success: true,
		User:    user,
	})
}
