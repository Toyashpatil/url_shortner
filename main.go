package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/urlshortner/config"
	"github.com/urlshortner/routes"
)

func init() {
	// Load .env file; if missing, keep going (we might be in prod)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; falling back to environment variables")
	}
}

func checkHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Health ok")
}

func main() {

	config.Connect()
	defer config.DB.Close()
	r := mux.NewRouter()

	r.HandleFunc("/", checkHealth)

	api := r.PathPrefix("/api").Subrouter()
	v1 := api.PathPrefix("/v1").Subrouter()
	routes.UrlRoutes(v1)
	fmt.Println("server started")
	r.Use(mux.CORSMethodMiddleware(r))
	err := http.ListenAndServe("127.0.0.1:3000", r)
	if err != nil {
		fmt.Println("error in starting the server")
	}

}
