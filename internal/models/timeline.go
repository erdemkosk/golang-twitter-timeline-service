package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Timeline struct {
	ID     string               `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID   `json:"userId" bson:"userId"`
	Tweets []primitive.ObjectID `json:"tweets,omitempty" bson:"tweets,omitempty"`
}
