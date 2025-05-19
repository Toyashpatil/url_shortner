package routes

import (
	"github.com/gorilla/mux"
	"github.com/urlshortner/config"
	"github.com/urlshortner/controllers"
	"github.com/urlshortner/middleware"
)

func UserRoutes(router *mux.Router) {

	api := router.PathPrefix("/api").Subrouter()
	v1 := api.PathPrefix("/v1").Subrouter()

	user := v1.PathPrefix("/user").Subrouter()

	controller := controllers.UserController{
		DB: config.DB,
	}

	user.HandleFunc("/createuser", controller.CreateUser).Methods("POST")
	user.HandleFunc("/login", controller.LoginUser).Methods("POST")
	protected := user.PathPrefix("/p").Subrouter()
	protected.Use(middleware.Auth)
	protected.HandleFunc("/getuser", controller.GetUser).Methods("GET")

}
