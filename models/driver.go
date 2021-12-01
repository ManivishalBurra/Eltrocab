package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Driver struct {
	Id                   primitive.ObjectID `json:"id" bson:"_id"`
	Name                 string             `json:"username" validate:"required,min=2,max=100"`
	Mail                 string             `json:"mail"`
	Password             string             `json:"password"`
	Lat                  float64            `json:"lat"`
	Long                 float64            `json:"long"`
	Token                string             `json:"token"`
	Refresh_Token        string             `json:"referesh_token"`
	CustomerConfirmation string             `json:"customer"`
	Pricing              float64            `json:"pricing"`
}
