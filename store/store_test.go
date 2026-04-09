package store

import (
	"context"
	"os"
	"strings"
	"testing"

	"bidprentjes-api/models"
)

func TestStoreWithScans(t *testing.T) {
	// Setup
	csvData := `1,Jan,,Jansen,1900-01-01,Amsterdam,1980-01-01,Amsterdam,true
2,Piet,,Pietersen,1910-01-01,Rotterdam,1990-01-01,Rotterdam,false
`
	scansData := `1,scan1.jpg
1,scan2.jpg
`

	os.MkdirAll("data", 0755)
	err := os.WriteFile("bidprentjes_test.csv", []byte(csvData), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("bidprentjes_test.csv")

	err = os.WriteFile("data/scans.csv", []byte(scansData), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("data/scans.csv")

	s := NewStore(context.Background(), "")
	defer s.Close()

	// Manually trigger processing with our test file
	scanFile, err := os.Open("data/scans.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer scanFile.Close()

	scanMap, err := s.parseScans(scanFile)
	if err != nil {
		t.Fatal(err)
	}

	n, err := s.ProcessCSVUpload(strings.NewReader(csvData), scanMap)
	if err != nil {
		t.Fatal(err)
	}

	if n != 2 {
		t.Errorf("Expected 2 records, got %d", n)
	}

	// Verify record 1 has Photo=true and 2 scans
	b1, exists := s.Get("1")
	if !exists {
		t.Fatal("Record 1 not found")
	}
	if !b1.Photo {
		t.Error("Expected Photo=true for record 1")
	}
	if len(b1.Scans) != 2 {
		t.Errorf("Expected 2 scans for record 1, got %d", len(b1.Scans))
	}
	if b1.Scans[0] != "scan1.jpg" || b1.Scans[1] != "scan2.jpg" {
		t.Errorf("Unexpected scan IDs: %v", b1.Scans)
	}

	// Verify record 2 has Photo=false and 0 scans
	b2, exists := s.Get("2")
	if !exists {
		t.Fatal("Record 2 not found")
	}
	if b2.Photo {
		t.Error("Expected Photo=false for record 2")
	}
	if len(b2.Scans) != 0 {
		t.Errorf("Expected 0 scans for record 2, got %d", len(b2.Scans))
	}
}

func TestSearchWithPhotos(t *testing.T) {
	s := NewStore(context.Background(), "")
	defer s.Close()

	s.BatchCreate([]*models.Bidprentje{
		{ID: "1", Achternaam: "Jansen", Photo: true, Scans: []string{"scan1.jpg"}},
		{ID: "2", Achternaam: "Pietersen", Photo: false, Scans: []string{}},
	})

	// Search for Jansen
	res := s.Search(models.SearchParams{Query: "Jansen", Page: 1, PageSize: 10})
	if res.TotalCount != 1 {
		t.Errorf("Expected 1 result, got %d", res.TotalCount)
	}
	if !res.Items[0].Photo {
		t.Error("Expected Photo=true for Jansen")
	}
	if len(res.Items[0].Scans) != 1 {
		t.Errorf("Expected 1 scan, got %d", len(res.Items[0].Scans))
	}
}
