package models

import (
	"encoding/json"
	"time"
)

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
}

// MarshalJSON implements custom JSON marshaling for Bidprentje
func (b Bidprentje) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID                string `json:"id"`
		Voornaam          string `json:"voornaam"`
		Tussenvoegsel     string `json:"tussenvoegsel"`
		Achternaam        string `json:"achternaam"`
		Geboortedatum     string `json:"geboortedatum"`
		Geboorteplaats    string `json:"geboorteplaats"`
		Overlijdensdatum  string `json:"overlijdensdatum"`
		Overlijdensplaats string `json:"overlijdensplaats"`
		Scan              bool   `json:"scan"`
	}{
		ID:                b.ID,
		Voornaam:          b.Voornaam,
		Tussenvoegsel:     b.Tussenvoegsel,
		Achternaam:        b.Achternaam,
		Geboortedatum:     b.Geboortedatum.Format("2006-01-02"),
		Geboorteplaats:    b.Geboorteplaats,
		Overlijdensdatum:  b.Overlijdensdatum.Format("2006-01-02"),
		Overlijdensplaats: b.Overlijdensplaats,
		Scan:              b.Scan,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Bidprentje
func (b *Bidprentje) UnmarshalJSON(data []byte) error {
	aux := &struct {
		ID                string `json:"id"`
		Voornaam          string `json:"voornaam"`
		Tussenvoegsel     string `json:"tussenvoegsel"`
		Achternaam        string `json:"achternaam"`
		Geboortedatum     string `json:"geboortedatum"`
		Geboorteplaats    string `json:"geboorteplaats"`
		Overlijdensdatum  string `json:"overlijdensdatum"`
		Overlijdensplaats string `json:"overlijdensplaats"`
		Scan              bool   `json:"scan"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	b.ID = aux.ID
	b.Voornaam = aux.Voornaam
	b.Tussenvoegsel = aux.Tussenvoegsel
	b.Achternaam = aux.Achternaam
	b.Geboorteplaats = aux.Geboorteplaats
	b.Overlijdensplaats = aux.Overlijdensplaats
	b.Scan = aux.Scan

	var err error
	if aux.Geboortedatum != "" {
		b.Geboortedatum, err = time.Parse("2006-01-02", aux.Geboortedatum)
		if err != nil {
			return err
		}
	}
	if aux.Overlijdensdatum != "" {
		b.Overlijdensdatum, err = time.Parse("2006-01-02", aux.Overlijdensdatum)
		if err != nil {
			return err
		}
	}
	return nil
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
