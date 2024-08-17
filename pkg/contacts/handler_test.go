package contacts_test

import (
	"bytes"
	"errors"
	"golangphonebook/pkg/contacts"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockContactRepository is a mock implementation of the ContactRepository interface
type MockContactRepository struct {
	addContactFn func(contact contacts.Contact) error
}

func (m *MockContactRepository) AddContact(contact contacts.Contact) error {
	if m.addContactFn != nil {
		return m.addContactFn(contact)
	}
	return nil
}

func (m *MockContactRepository) GetContacts(page int) error {
	return nil
}

func (m *MockContactRepository) GetAllContacts() {
	return
}

func (m *MockContactRepository) UpdateContact(id int, contact contacts.Contact) error {
	return nil
}

func (m *MockContactRepository) DeleteContact(id int) error {
	return nil
}

func (m *MockContactRepository) GetContactCount() (int64, error) {
	return 0, nil
}

func TestPutContact(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        string
		mockAddContactFn   func(contact contacts.Contact) error
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Valid Request",
			requestBody: `{
				"id": 1,
				"first_name": "John",
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"
			}`,
			mockAddContactFn: func(contact contacts.Contact) error {
				return nil
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Invalid JSON Request",
			requestBody: `{
				"id": 1,
				"first_name": "John",
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"`, // Missing closing brace
			mockAddContactFn:   nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid request body",
		},
		{
			name: "Failed to Insert Contact",
			requestBody: `{
				"id": 1,
				"first_name": "John",
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"
			}`,
			mockAddContactFn: func(contact contacts.Contact) error {
				return errors.New("database error")
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "Failed to insert contact to db with error database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock repository with the provided mock function
			mockRepo := &MockContactRepository{
				addContactFn: tt.mockAddContactFn,
			}

			// Create a new request with the provided body
			req, err := http.NewRequest("PUT", "/contact", bytes.NewBuffer([]byte(tt.requestBody)))
			assert.NoError(t, err)

			// Create a response recorder to capture the response
			rr := httptest.NewRecorder()

			// Call the PutContact handler with the mock repository
			contacts.PutContact(rr, req, mockRepo)

			// Assert the status code
			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			// Assert the response body if expected
			if tt.expectedResponse != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedResponse)
			}
		})
	}
}
