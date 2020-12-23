package types

type ActiveAddress struct {
	Address    string `json:"address" bson:"address"`
	IsContract bool   `json:"isContract" bson:"isContract"`
}
