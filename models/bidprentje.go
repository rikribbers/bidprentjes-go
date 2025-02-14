package models

import "time"

type Bidprentje struct {
	ID                string    `json:"id"`
	Voornaam          string    `json:"voornaam"`
	Tussenvoegsel     string    `json:"tussenvoegsel"`
	Achternaam        string    `json:"achternaam"`
	Geboortedatum     time.Time `json:"geboortedatum"`
	Geboorteplaats    string    `json:"geboorteplaats"`
	Overlijdensdatum  time.Time `json:"overlijdensdatum"`
	Overlijdensplaats string    `json:"overlijdensplaats"`
	Scan              bool      `json:"scan"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type BidprentjeSearch struct {
	Query string `json:"query"`
}

type PaginatedResponse struct {
	Items      []*Bidprentje `json:"items"`
	TotalCount int           `json:"total_count"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
}

type ListParams struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Query    string `form:"query"`
}
