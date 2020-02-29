package data

type OilPrice struct {
	Id          uint   `json:"index"`
	Location    string `json:"location"`
	Price92     string `json:"price92"`
	Price95     string `json:"price95"`
	Price98     string `json:"price98"`
	PriceDiesel string `json:"pricediesel"`
}
