package main

import (
	gorilla_mux "github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func routes() *gorilla_mux.Router {
	mux := gorilla_mux.NewRouter()

	mux.HandleFunc("/subscription", CreateSubscriptionHandler).Methods("POST")
	mux.HandleFunc("/subscription/total", GetSubscriptionsTotalHandler).Methods("GET")
	mux.HandleFunc("/subscription/{id}", GetSubscriptionByIdHandler).Methods("GET")
	mux.HandleFunc("/subscription/{id}", UpdateSubscriptionHandler).Methods("PATCH")
	mux.HandleFunc("/subscription/{id}", DeleteSubscriptionHandler).Methods("DELETE")
	mux.HandleFunc("/subscription", GetAllSubscriptionHandler).Methods("GET")
	mux.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	return mux
}
