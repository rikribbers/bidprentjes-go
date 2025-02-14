package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBidprentjeJSON(t *testing.T) {
	now := time.Now()
	b := Bidprentje{
		ID:                "test-id",
		Voornaam:          "Jan",
		Tussenvoegsel:     "van",
		Achternaam:        "Test",
		Geboortedatum:     now,
		Geboorteplaats:    "Amsterdam",
		Overlijdensdatum:  now,
		Overlijdensplaats: "Rotterdam",
		Scan:              true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// Test marshaling
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("Failed to marshal Bidprentje: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Bidprentje
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Bidprentje: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != b.ID {
		t.Errorf("Expected ID %s, got %s", b.ID, unmarshaled.ID)
	}
	if unmarshaled.Voornaam != b.Voornaam {
		t.Errorf("Expected Voornaam %s, got %s", b.Voornaam, unmarshaled.Voornaam)
	}
	// Add more field comparisons as needed
}
