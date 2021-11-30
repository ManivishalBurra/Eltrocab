package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	C "github.com/manivishalburra/eltrocab/controllers"
)

func main() {
	app := mux.NewRouter()
	app.HandleFunc("/setdriver", C.CreateDriver)
	app.HandleFunc("/getsuitableride", C.GetSuitableRide)
	app.HandleFunc("/logindriver", C.LoginDriver)
	log.Fatal(http.ListenAndServe(":8080", app))
}
