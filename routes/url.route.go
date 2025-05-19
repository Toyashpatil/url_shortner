package routes

import (
	"github.com/gorilla/mux"
	"github.com/urlshortner/config"
	"github.com/urlshortner/controllers"
	"github.com/urlshortner/middleware"
)

func UrlRoutes(router *mux.Router) {
	api := router.PathPrefix("/api").Subrouter()
	v1 := api.PathPrefix("/v1").Subrouter()

	controller := controllers.UrlController{
		DB: config.DB,
	}
	router.HandleFunc("/{code}", controller.GetTheUrl).Methods("GET")
	protected := v1.PathPrefix("/p").Subrouter()
	protected.Use(middleware.Auth)
	protected.HandleFunc("/urlshort", controller.ShortTheUrl).Methods("POST")
	protected.HandleFunc("/getuserurls", controller.GetUsersUrl).Methods("POST")
	protected.HandleFunc("/delete/{id}", controller.DeleteUrl).Methods("POST")

}
