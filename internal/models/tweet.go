package models

type Tweet struct {
	ID     string `json:"id" bson:"_id,omitempty"`
	UserId string `json:"userid" bson:"userid"`
	Tweet  string `json:"tweet" bson:"tweet"`
}
