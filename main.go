package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
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

	routes.UserRoutes(r)
	routes.UrlRoutes(r)
	fmt.Println("server started")
	corsOpts := handlers.CORS(
		handlers.AllowedOrigins([]string{
			"*", // add any other allowed origins here
		}),
		handlers.AllowedMethods([]string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		}),
		handlers.AllowedHeaders([]string{
			"Content-Type",
			"Authorization",
		}),
		// If you need to expose custom headers in responses:
		// handlers.ExposedHeaders([]string{"X-My-Custom-Header"}),
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	err := http.ListenAndServe(addr, corsOpts(r))
	if err != nil {
		fmt.Println("error in starting the server")
	}

}
