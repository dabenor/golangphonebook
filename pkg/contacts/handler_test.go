package contacts

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPutContact(t *testing.T) {
	// Mock the addContact function to simulate different behaviors
	addContact = func(contact Contact) error {
		if contact.FirstName == "Error" {
			return errors.New("mock database error")
		}
		return nil
	}

	tests := []struct {
		name               string
		requestBody        string
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:               "Valid Input",
			requestBody:        `{"id":1,"first_name":"John","last_name":"Doe","phone":"+1234567890","address":"123 Main St"}`,
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "",
		},
		{
			name:               "Invalid JSON Body",
			requestBody:        `{"id":1,"first_name":"John","last_name":"Doe","phone":"+1234567890",`,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid request body",
		},
		{
			name:               "Error Adding Contact to Database",
			requestBody:        `{"id":2,"first_name":"Error","last_name":"Doe","phone":"+1234567890","address":"123 Main St"}`,
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "Failed to insert contact to db with error mock database error",
		},
		{
			name:               "Missing Required Fields",
			requestBody:        `{"id":3,"last_name":"Doe","phone":"+1234567890","address":"123 Main St"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with the provided body
			req, err := http.NewRequest("PUT", "/contact", bytes.NewBuffer([]byte(tt.requestBody)))
			if err != nil {
				t.Fatal(err)
			}

			// Create a response recorder to capture the response
			rr := httptest.NewRecorder()

			// Call the handler function
			PutContact(rr, req)

			// Assert the status code
			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			// Assert the response body (if expected)
			if tt.expectedResponse != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedResponse)
			}
		})
	}
}
