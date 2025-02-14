package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bidprentjes-api/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	handler := NewHandler()

	router.POST("/bidprentjes", handler.CreateBidprentje)
	router.GET("/bidprentjes/:id", handler.GetBidprentje)
	router.PUT("/bidprentjes/:id", handler.UpdateBidprentje)
	router.DELETE("/bidprentjes/:id", handler.DeleteBidprentje)
	router.GET("/bidprentjes", handler.ListBidprentjes)
	router.POST("/bidprentjes/search", handler.SearchBidprentjes)

	return router, handler
}

func createTestBidprentjes(t *testing.T, router *gin.Engine) []models.Bidprentje {
	testData := []models.Bidprentje{
		{
			Voornaam:          "Johannes",
			Tussenvoegsel:     "van",
			Achternaam:        "Amsterdam",
			Geboortedatum:     time.Now(),
			Geboorteplaats:    "Utrecht",
			Overlijdensdatum:  time.Now(),
			Overlijdensplaats: "Amsterdam",
			Scan:              true,
		},
		{
			Voornaam:          "Johanna",
			Tussenvoegsel:     "van der",
			Achternaam:        "Berg",
			Geboortedatum:     time.Now(),
			Geboorteplaats:    "Amsterdam",
			Overlijdensdatum:  time.Now(),
			Overlijdensplaats: "Rotterdam",
			Scan:              false,
		},
	}

	createdBidprentjes := make([]models.Bidprentje, 0, len(testData))
	for _, b := range testData {
		body, err := json.Marshal(b)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/bidprentjes", bytes.NewBuffer(body))
		router.ServeHTTP(w, req)

		var created models.Bidprentje
		err = json.Unmarshal(w.Body.Bytes(), &created)
		assert.NoError(t, err)
		createdBidprentjes = append(createdBidprentjes, created)
	}

	return createdBidprentjes
}

func TestCreateBidprentje(t *testing.T) {
	router, _ := setupTestRouter()
	bidprentje := models.Bidprentje{
		Voornaam:          "Test",
		Tussenvoegsel:     "van",
		Achternaam:        "User",
		Geboortedatum:     time.Now(),
		Geboorteplaats:    "TestCity",
		Overlijdensdatum:  time.Now(),
		Overlijdensplaats: "TestCity2",
		Scan:              true,
	}

	body, _ := json.Marshal(bidprentje)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/bidprentjes", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Bidprentje
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, bidprentje.Voornaam, response.Voornaam)
}

func TestGetBidprentje(t *testing.T) {
	router, _ := setupTestRouter()
	bidprentje := models.Bidprentje{
		Voornaam:          "Johannes",
		Tussenvoegsel:     "van",
		Achternaam:        "Amsterdam",
		Geboortedatum:     time.Now(),
		Geboorteplaats:    "Utrecht",
		Overlijdensdatum:  time.Now(),
		Overlijdensplaats: "Amsterdam",
		Scan:              true,
	}

	// Create a test bidprentje first
	body, _ := json.Marshal(bidprentje)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/bidprentjes", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	var created models.Bidprentje
	_ = json.Unmarshal(w.Body.Bytes(), &created)

	// Test getting the created bidprentje
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/bidprentjes/"+created.ID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Bidprentje
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, response.ID)
}

func TestUpdateBidprentje(t *testing.T) {
	router, _ := setupTestRouter()
	createdBidprentjes := createTestBidprentjes(t, router)
	bidprentje := createdBidprentjes[0]

	// Update the bidprentje
	bidprentje.Voornaam = "UpdatedName"
	body, _ := json.Marshal(bidprentje)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/bidprentjes/"+bidprentje.ID, bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Bidprentje
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "UpdatedName", response.Voornaam)

	// Verify search is updated
	search := models.BidprentjeSearch{Query: "UpdatedName"}
	body, _ = json.Marshal(search)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/bidprentjes/search", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	var searchResponse []models.Bidprentje
	err = json.Unmarshal(w.Body.Bytes(), &searchResponse)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(searchResponse))
	assert.Equal(t, "UpdatedName", searchResponse[0].Voornaam)
}

func TestDeleteBidprentje(t *testing.T) {
	router, _ := setupTestRouter()
	createdBidprentjes := createTestBidprentjes(t, router)
	bidprentje := createdBidprentjes[0]

	// Delete the bidprentje
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/bidprentjes/"+bidprentje.ID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's deleted from search
	search := models.BidprentjeSearch{Query: bidprentje.Voornaam}
	body, _ := json.Marshal(search)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/bidprentjes/search", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	var searchResponse []models.Bidprentje
	err := json.Unmarshal(w.Body.Bytes(), &searchResponse)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(searchResponse))
}

func TestListBidprentjes(t *testing.T) {
	router, _ := setupTestRouter()

	// Create multiple test bidprentjes
	for i := 0; i < 3; i++ {
		bidprentje := models.Bidprentje{
			Voornaam:          "Johannes",
			Tussenvoegsel:     "van",
			Achternaam:        "Amsterdam",
			Geboortedatum:     time.Now(),
			Geboorteplaats:    "Utrecht",
			Overlijdensdatum:  time.Now(),
			Overlijdensplaats: "Amsterdam",
			Scan:              true,
		}
		body, _ := json.Marshal(bidprentje)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/bidprentjes", bytes.NewBuffer(body))
		router.ServeHTTP(w, req)
	}

	// Test listing all bidprentjes
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/bidprentjes", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Bidprentje
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(response))
}

func TestSearchBidprentjes(t *testing.T) {
	router, _ := setupTestRouter()
	_ = createTestBidprentjes(t, router) // Just create the test data, we don't need the return value

	testCases := []struct {
		name          string
		searchQuery   string
		expectedCount int
		expectedName  string
	}{
		{
			name:          "Exact match",
			searchQuery:   "Johannes",
			expectedCount: 1,
			expectedName:  "Johannes",
		},
		{
			name:          "Fuzzy match - one character off",
			searchQuery:   "Johanns",
			expectedCount: 1,
			expectedName:  "Johannes",
		},
		{
			name:          "Fuzzy match - similar names",
			searchQuery:   "Johan",
			expectedCount: 2,
			expectedName:  "", // Don't check name as we might get either
		},
		{
			name:          "Location search",
			searchQuery:   "Amsterdam",
			expectedCount: 2, // Should find both birth and death places
		},
		{
			name:          "No results",
			searchQuery:   "NonExistent",
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			search := models.BidprentjeSearch{Query: tc.searchQuery}
			body, _ := json.Marshal(search)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/bidprentjes/search", bytes.NewBuffer(body))
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response []models.Bidprentje
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCount, len(response))

			if tc.expectedName != "" && len(response) > 0 {
				assert.Equal(t, tc.expectedName, response[0].Voornaam)
			}
		})
	}
}
