package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID        primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Username  string               `json:"username" bson:"username"`
	Followers []primitive.ObjectID `json:"followers,omitempty" bson:"followers,omitempty"`
	Following []primitive.ObjectID `json:"following,omitempty" bson:"following,omitempty"`
	Tweets    []primitive.ObjectID `json:"tweets,omitempty" bson:"tweets,omitempty"`
}
