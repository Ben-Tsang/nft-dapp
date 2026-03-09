package handler

type MintEventData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TokenURI    string `json:"tokenURI"`
}

type ListedEventData struct {
	Price    string `json:"price"`
	ListedAt string `json:"listedAt"`
}

type UnlistedEventData struct {
	UnlistedAt string `json:"unlistedAt"`
}

type SetPriceEventData struct {
	Price string `json:"price"`
	SetAt string `json:"setAt"`
}

type BuyEventData struct {
	Price string `json:"price"`
	BuyAt string `json:"buyAt"`
}
