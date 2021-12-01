package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Request struct {
	Id                   primitive.ObjectID `json:"id" bson:"_id"`
	Name                 string             `json:"username" validate:"required,min=2,max=100"`
	Mail                 string             `json:"mail"`
	Lat                  float64            `json:"lat"`
	Long                 float64            `json:"long"`
	DstLat               float64            `json:"dstlat"`
	DstLng               float64            `json:"dstlng"`
	DriverConfirmation   string             `json:"driverconfirmation"`
	CustomerConfirmation string             `json:"customerconfirmation"`
}
