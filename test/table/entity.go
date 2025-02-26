package table

import "time"

type MyEntity struct {
	PartitionKey string    `json:"PartitionKey"`
	RowKey       string    `json:"RowKey"`
	Stock        int       `json:"Stock"`
	Price        float64   `json:"Price"`
	Comments     string    `json:"Comments"`
	OnSale       bool      `json:"OnSale"`
	ReducedPrice float64   `json:"ReducedPrice"`
	PurchaseDate time.Time `json:"PurchaseDate"`
	BinaryRep    []byte    `json:"BinaryRepresentation"`
}
