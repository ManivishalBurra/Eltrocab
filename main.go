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
	app.HandleFunc("/logindriver", C.LoginDriver)
	app.HandleFunc("/cabrequests", C.FetchRequest)
	app.HandleFunc("/setuser", C.CreateUser)
	app.HandleFunc("/loginuser", C.LoginUser)
	app.HandleFunc("/bookride", C.BookRide)
	app.HandleFunc("/driverconfirm", C.DriverConfirm)
	app.HandleFunc("/ridestatus", C.RideStatus)
	app.HandleFunc("/usercancelride", C.LoginUser)
	app.HandleFunc("/userlogout", C.UserLogout)
	app.HandleFunc("/driverlogout", C.DriverLogout)
	log.Fatal(http.ListenAndServe(":8080", app))
}
