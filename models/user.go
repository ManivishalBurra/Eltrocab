package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id       primitive.ObjectID `json:"id" bson:"_id"`
	Name     string             `json:"username" validate:"required,min=2,max=100"`
	Mail     string             `json:"mail"`
	Password string             `json:"password"`
	Token    string             `json:"token"`
}
