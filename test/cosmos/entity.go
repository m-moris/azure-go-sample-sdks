package cosmos

type Item struct {
	Id           string  `json:"id"`
	PartitionKey string  `json:"pk"`
	Category     string  `json:"category"`
	Name         string  `json:"name"`
	Quantity     int     `json:"quantity"`
	Price        float32 `json:"price"`
	Clearance    bool    `json:"clearance"`
}
