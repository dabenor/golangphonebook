package contacts_test

import (
	"bytes"
	"errors"
	"golangphonebook/pkg/contacts"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// MockContactRepository is a mock implementation of the ContactRepository interface
type MockContactRepository struct {
	addContactFn    func(contact contacts.Contact) error
	updateContactFn func(id int, contact contacts.Contact) error
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
	if m.updateContactFn != nil {
		return m.updateContactFn(id, contact)
	}
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
			expectedResponse:   "Contact added to DB successfully",
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
			expectedResponse:   "Invalid request body, first name and phone must be correctly defined",
		},
		{
			name: "Missing First Name",
			requestBody: `{
				"id": 1,
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"
			}`,
			mockAddContactFn:   nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid request body, first name and phone must be correctly defined",
		},
		{
			name: "Missing Phone Number",
			requestBody: `{
				"id": 1,
				"first_name": "John",
				"last_name": "Doe",
				"address": "123 Main St"
			}`,
			mockAddContactFn:   nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid request body, first name and phone must be correctly defined",
		},
		{
			name: "Empty First Name",
			requestBody: `{
				"id": 1,
				"first_name": "",
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"
			}`,
			mockAddContactFn:   nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid request body, first name and phone must be correctly defined",
		},
		{
			name: "Duplicate Contact",
			requestBody: `{
				"id": 1,
				"first_name": "John",
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"
			}`,
			mockAddContactFn: func(contact contacts.Contact) error {
				return errors.New("contact with the same full name and phone number already exists")
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "contact with the same full name and phone number already exists",
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

func TestUpdateContact(t *testing.T) {
	tests := []struct {
		name                string
		url                 string
		requestBody         string
		mockUpdateContactFn func(id int, contact contacts.Contact) error
		expectedStatusCode  int
		expectedResponse    string
	}{
		{
			name: "Valid Update",
			url:  "/updateContact/1",
			requestBody: `{
				"first_name": "John",
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"
			}`,
			mockUpdateContactFn: func(id int, contact contacts.Contact) error {
				return nil
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "Contact updated successfully",
		},
		{
			name: "Invalid ID",
			url:  "/updateContact/invalid",
			requestBody: `{
				"first_name": "John",
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"
			}`,
			mockUpdateContactFn: nil,
			expectedStatusCode:  http.StatusBadRequest,
			expectedResponse:    "Invalid ID, IDs can only be integers",
		},
		{
			name: "Invalid JSON Request",
			url:  "/updateContact/1",
			requestBody: `{
				"first_name": "John",
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"`, // Missing closing brace
			mockUpdateContactFn: nil,
			expectedStatusCode:  http.StatusBadRequest,
			expectedResponse:    "Invalid request body",
		},
		{
			name: "Missing First Name",
			url:  "/updateContact/1",
			requestBody: `{
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"
			}`,
			mockUpdateContactFn: nil,
			expectedStatusCode:  http.StatusBadRequest,
			expectedResponse:    "Invalid request body",
		},
		{
			name: "Duplicate Contact",
			url:  "/updateContact/1",
			requestBody: `{
				"first_name": "John",
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"
			}`,
			mockUpdateContactFn: func(id int, contact contacts.Contact) error {
				return errors.New("another contact with the same first name, last name, and phone number already exists")
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Duplicate contact with the same first name, last name, and phone number already exists",
		},
		{
			name: "Internal Server Error",
			url:  "/updateContact/1",
			requestBody: `{
				"first_name": "John",
				"last_name": "Doe",
				"phone": "+1234567890",
				"address": "123 Main St"
			}`,
			mockUpdateContactFn: func(id int, contact contacts.Contact) error {
				return errors.New("database error")
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "Failed to update contact due to an internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with the body
			req, err := http.NewRequest("PUT", tt.url, bytes.NewBuffer([]byte(tt.requestBody)))
			assert.NoError(t, err)

			// Create response recorder to capture the response
			rr := httptest.NewRecorder()

			// Create a router and register the handler to simulate correct routing
			router := mux.NewRouter()
			router.HandleFunc("/updateContact/{id}", func(w http.ResponseWriter, r *http.Request) {
				contacts.UpdateContact(w, r, &MockContactRepository{
					updateContactFn: tt.mockUpdateContactFn,
				})
			}).Methods("PUT")

			// Serve the request
			router.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			// Check response body if expected
			if tt.expectedResponse != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedResponse)
			}
		})
	}
}
