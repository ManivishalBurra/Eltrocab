package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ManivishalBurra/Eltrocab/models"
	U "github.com/ManivishalBurra/Eltrocab/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtKey = []byte("secret_key")

type Credentials struct {
	Mail     string `json:"mail"`
	Password string `json:"password"`
}

type Drivermail struct {
	Mail string `json:"mail"`
}

type Driverconf struct {
	Mail     string `json:"mail"`
	UserMail string `json:"usermail"`
}

type Bookings struct {
	Name     string  `json:"name"`
	Mail     string  `json:"mail"`
	Distance float64 `json:"distance in km"`
	Fare     float64 `json:"fare in Rs"`
}

type Token struct {
	Token string `json:"token"`
}

type Claims struct {
	Mail string `json:"mail"`
	jwt.StandardClaims
}

func CreateDriver(w http.ResponseWriter, r *http.Request) {
	driverDetails := models.Driver{}
	client, err := U.Session()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)
	err = json.NewDecoder(r.Body).Decode(&driverDetails)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	driverDetails.Id = primitive.NewObjectID()
	client.Database("eltrocab").Collection("driver").InsertOne(ctx, driverDetails)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	uj, err := json.Marshal(driverDetails)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}

func LoginDriver(w http.ResponseWriter, r *http.Request) {
	var credentials Credentials
	json.NewDecoder(r.Body).Decode(&credentials)
	client, err := U.Session()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)
	fmt.Println(credentials.Mail)
	cursor, err := client.Database("eltrocab").Collection("driver").Find(ctx, bson.M{"mail": credentials.Mail})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)
	var data bson.M
	for cursor.Next(ctx) {
		if err = cursor.Decode(&data); err != nil {
			log.Fatal(err)
		}
	}

	if credentials.Password != data["password"] {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	claims := &Claims{
		Mail:           credentials.Mail,
		StandardClaims: jwt.StandardClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Header.Add("Authorization", tokenString)
	result, err := client.Database("eltrocab").Collection("driver").UpdateOne(
		ctx,
		bson.M{"mail": credentials.Mail},
		bson.D{
			{"$set", bson.D{{"token", tokenString}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("updated %v doc\n", result.ModifiedCount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var tkn Token
	tkn.Token = tokenString
	uj, err := json.Marshal(tkn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}

func FetchRequest(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	tokenStr := strings.ReplaceAll(auth, "Bearer ", "")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	mail := U.Decode(tokenStr)
	var credentials Drivermail
	json.NewDecoder(r.Body).Decode(&credentials)
	if mail != credentials.Mail {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	client, err := U.Session()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)
	result, err := client.Database("eltrocab").Collection("request").Find(ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var element []bson.M
	if err = result.All(ctx, &element); err != nil {
		log.Fatal(err)
	}
	defer result.Close(ctx)
	data := models.Request{}
	var rides Bookings
	var bookings []Bookings
	for _, ele := range element {
		mapstructure.Decode(ele, &data)
		if data.CustomerConfirmation == "cancel" || data.DriverConfirmation == "accepted" {
			continue
		}
		fare := U.Fare(data.Lat, data.Long, data.DstLat, data.DstLng)
		mapstructure.Decode(ele, &rides)
		rides.Fare = fare[1]
		rides.Distance = fare[0]
		bookings = append(bookings, rides)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	uj, err := json.Marshal(bookings)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}

func DriverConfirm(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	tokenStr := strings.ReplaceAll(auth, "Bearer ", "")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	mail := U.Decode(tokenStr)
	var driverconf Driverconf
	err = json.NewDecoder(r.Body).Decode(&driverconf)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if driverconf.Mail != mail {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	client, err := U.Session()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)
	fmt.Println(driverconf)
	result, err := client.Database("eltrocab").Collection("request").UpdateOne(
		ctx,
		bson.M{"mail": driverconf.UserMail},
		bson.D{
			{"$set", bson.D{{"driverconfirmation", "accepted"}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var message Message
	message.Message = "customer is wating!! ride booked"
	uj, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}

func DriverLogout(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	tokenStr := strings.ReplaceAll(auth, "Bearer ", "")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	mail := U.Decode(tokenStr)
	var usermail UserMail
	json.NewDecoder(r.Body).Decode(&usermail)
	if usermail.Mail != mail {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	fmt.Println(usermail)
	fmt.Println(mail)
	client, err := U.Session()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)
	result, err := client.Database("eltrocab").Collection("driver").UpdateOne(
		ctx,
		bson.M{"mail": usermail.Mail},
		bson.D{
			{"$set", bson.D{{"token", ""}}},
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var message Message
	message.Message = "You Logged out"
	uj, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}

func DriverCancelRide(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	tokenStr := strings.ReplaceAll(auth, "Bearer ", "")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	mail := U.Decode(tokenStr)
	var usermail Driverconf
	json.NewDecoder(r.Body).Decode(&usermail)
	fmt.Println(usermail.Mail)
	fmt.Println(mail)
	if usermail.Mail != mail {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	client, err := U.Session()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)
	client.Database("eltrocab").Collection("request").UpdateOne(
		ctx,
		bson.M{"mail": usermail.UserMail},
		bson.D{
			{"$set", bson.D{{"driverconfirmation", "cancel"}}},
		},
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var message Message
	message.Message = "Your ride is cancelled"
	uj, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}
