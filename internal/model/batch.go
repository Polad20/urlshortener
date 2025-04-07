package model

type Incoming struct {
	Correlation_id string `json:"correlation_id"`
	Original_url   string `json:"original_url"`
}

type DbSave struct {
	UserID         string `db:"userID"`
	Correlation_id string `db:"correlation_id"`
	Original_url   string `db:"original_url"`
	Short_url      string `db:"short_url"`
}

type ClientResponse struct {
	Correlation_id string `json:"correlation_id"`
	Short_url      string `json:"short_url"`
}
