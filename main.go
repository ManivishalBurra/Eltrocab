package main

import (
	"log"
	"net/http"

	C "github.com/ManivishalBurra/Eltrocab/controllers"
	"github.com/gorilla/mux"
)

func main() {
	app := mux.NewRouter()
	app.HandleFunc("/driver", C.CreateDriver)
	app.HandleFunc("/login/driver", C.LoginDriver)
	app.HandleFunc("/cabrequests", C.FetchRequest)
	app.HandleFunc("/user", C.CreateUser)
	app.HandleFunc("/login/user", C.LoginUser)
	app.HandleFunc("/bookride", C.BookRide)
	app.HandleFunc("/driverconfirm", C.DriverConfirm)
	app.HandleFunc("/ridestatus", C.RideStatus)
	app.HandleFunc("/usercancelride", C.UserCancelRide)
	app.HandleFunc("/drivercancelride", C.DriverCancelRide)
	app.HandleFunc("/user/logout", C.UserLogout)
	app.HandleFunc("/driver/logout", C.DriverLogout)
	log.Fatal(http.ListenAndServe(":8080", app))
}
