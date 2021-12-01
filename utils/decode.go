package utils

import (
	"fmt"
	"reflect"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("secret_key")

func Decode(tokenString string) string {
	claims := jwt.MapClaims{}
	jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	var str string
	for key, val := range claims {
		fmt.Printf("Key: %v, value: %v\n", key, val)
		fmt.Println(reflect.TypeOf(val))
		str = fmt.Sprintf("%v", val)
	}
	return str

}
