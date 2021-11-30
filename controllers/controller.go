package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	U "github.com/manivishalburra/eltrocab/Utils"
	"github.com/manivishalburra/eltrocab/models"
	"github.com/mmcloughlin/spherand"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtKey = []byte("secret_key")

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
	lat, lng := spherand.Geographical()
	driverDetails.Lat = lat
	driverDetails.Long = lng
	client.Database("eltrocab").Collection("driver").InsertOne(ctx, driverDetails)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	uj, err := json.Marshal(driverDetails)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}
func GetSuitableRide(w http.ResponseWriter, r *http.Request) {
	//drivers := models.Driver{}
	client, err := U.Session()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)
	cursor, err := client.Database("eltrocab").Collection("driver").Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	var dr []bson.M
	if err = cursor.All(ctx, &dr); err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	uj, err := json.Marshal(dr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", uj)
}

type Credentials struct {
	Mail     string `json:"mail"`
	Password string `json:"password"`
}

type Claims struct {
	Mail string `json:"mail"`
	jwt.StandardClaims
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
	expirationTime := time.Now().Add(time.Minute * 5)
	claims := &Claims{
		Mail: credentials.Mail,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w,
		&http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
		})
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
}
