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

type UserMail struct {
	Mail string `json:"mail"`
}
type Location struct {
	Mail   string  `json:"mail"`
	Lat    float64 `json:"lat"`
	Long   float64 `json:"long"`
	DstLat float64 `json:"dstlat"`
	DstLng float64 `json:"dstlng"`
}

type Message struct {
	Message string `json:"message"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	userDetails := models.User{}
	client, err := U.Session()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)
	err = json.NewDecoder(r.Body).Decode(&userDetails)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userDetails.Id = primitive.NewObjectID()
	client.Database("eltrocab").Collection("user").InsertOne(ctx, userDetails)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	uj, err := json.Marshal(userDetails)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
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
	cursor, err := client.Database("eltrocab").Collection("user").Find(ctx, bson.M{"mail": credentials.Mail})
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
	//expirationTime := time.Now().Add(time.Minute * 5)
	claims := &Claims{
		Mail: credentials.Mail,
		// StandardClaims: jwt.StandardClaims{
		// 	ExpiresAt: expirationTime.Unix(),
		// },
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tkn := tokenString
	r.Header.Add("Authorization", tkn)
	data["token"] = tkn
	result, err := client.Database("eltrocab").Collection("user").UpdateOne(
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
	uj, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}

func BookRide(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	tokenStr := strings.ReplaceAll(auth, "Bearer ", "")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
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
	//Decoding is done!!!!!
	var credentials Credentials
	credentials.Mail = mail
	var location Location
	request := models.Request{}
	json.NewDecoder(r.Body).Decode(&location)
	fmt.Println(location)
	fmt.Println(credentials)
	println("safe")
	if location.Mail != credentials.Mail {
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
	cursor, err := client.Database("eltrocab").Collection("user").Find(ctx, bson.M{"mail": credentials.Mail})
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
	fmt.Println(data)
	mapstructure.Decode(data, &request)
	request.Id = primitive.NewObjectID()
	request.CustomerConfirmation = "pending"
	request.DriverConfirmation = "pending"
	request.DstLat = location.DstLat
	request.DstLng = location.DstLng
	request.Lat = location.Lat
	request.Long = location.Long
	check, err := client.Database("eltrocab").Collection("request").Find(ctx, bson.M{"mail": credentials.Mail})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer check.Close(ctx)
	var da bson.M
	for check.Next(ctx) {
		if err = check.Decode(&da); err != nil {
			log.Fatal(err)
		}
	}
	if len(da) != 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "you already requested a cab")
		return
	}
	client.Database("eltrocab").Collection("request").InsertOne(ctx, request)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	uj, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}

func RideStatus(w http.ResponseWriter, r *http.Request) {
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
	client, err := U.Session()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)
	cursor, err := client.Database("eltrocab").Collection("request").Find(ctx, bson.M{"mail": usermail.Mail})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var element []bson.M
	if err = cursor.All(ctx, &element); err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	data := models.Request{}
	var rides Bookings
	mapstructure.Decode(element[0], &data)
	if data.DriverConfirmation == "accepted" {
		arr := U.Fare(data.Lat, data.Long, data.DstLat, data.DstLng)
		mapstructure.Decode(data, rides)
		rides.Mail = usermail.Mail
		rides.Name = "User"
		rides.Distance = arr[0]
		rides.Fare = arr[1]
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Println(rides)
		uj, err := json.Marshal(rides)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, "%s\n", uj)
	} else if data.DriverConfirmation == "cancel" {
		fmt.Fprintf(w, "Your booking is cancelled, please book again\n")
	} else {
		fmt.Fprintf(w, "Your booking is in pending, please wait\n")
	}

}

func UserLogout(w http.ResponseWriter, r *http.Request) {
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
	result, err := client.Database("eltrocab").Collection("user").UpdateOne(
		ctx,
		bson.M{"mail": usermail.Mail},
		bson.D{
			{"$set", bson.D{{"token", ""}}},
		},
	)
	fmt.Println(result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var message Message
	message.Message = "You logged out"
	uj, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}

func UserCancelRide(w http.ResponseWriter, r *http.Request) {
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
		bson.M{"mail": usermail.Mail},
		bson.D{
			{"$set", bson.D{{"customerconfirmation", "cancel"}}},
		},
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var message Message
	message.Message = "You Cancelled the ride"
	uj, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}
