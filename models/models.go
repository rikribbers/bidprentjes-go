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

type SearchParams struct {
	Query    string `form:"query"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=25"`
}

type PaginatedResponse struct {
	Items      []Bidprentje `json:"items"`
	TotalCount int          `json:"total_count"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
}
