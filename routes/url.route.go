package routes

import (
	"github.com/gorilla/mux"
	"github.com/urlshortner/config"
	"github.com/urlshortner/controllers"
)

func UrlRoutes(router *mux.Router) {

	controller := controllers.UrlController{
		DB: config.DB,
	}

	router.HandleFunc("/urlshort", controller.ShortTheUrl).Methods("POST")
	router.HandleFunc("/{code}", controller.GetTheUrl).Methods("GET")

}
