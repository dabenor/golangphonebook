package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golangphonebook/db"
	"golangphonebook/internal"
	"golangphonebook/pkg/contacts"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var testServer *httptest.Server

func setupRouter() *mux.Router {
	db, err := db.DBInit()
	if err != nil {
		internal.Logger.Error("Failed to initialize test database")
		panic(err)
	}
	repo := contacts.NewSQLContactRepository(db)

	router := mux.NewRouter()
	// C
	router.HandleFunc("/addContact", func(w http.ResponseWriter, r *http.Request) { contacts.PutContact(w, r, repo) }).Methods("POST")
	// R
	router.HandleFunc("/getContacts", func(w http.ResponseWriter, r *http.Request) { contacts.GetContacts(w, r, repo) }).Methods("GET")
	router.HandleFunc("/getAllContacts", func(w http.ResponseWriter, r *http.Request) { contacts.GetAllContacts(w, r, repo) }).Methods("GET")
	// U
	router.HandleFunc("/updateContact/{id}", func(w http.ResponseWriter, r *http.Request) { contacts.UpdateContact(w, r, repo) }).Methods("POST")
	// D
	router.HandleFunc("/deleteContact/{id}", func(w http.ResponseWriter, r *http.Request) { contacts.DeleteContact(w, r, repo) }).Methods("DELETE")
	return router
}

func TestE2E(t *testing.T) {
	setupTestServer()
	t.Run("CreateContact", testCreateContact)

	resetDatabase()
	t.Run("GetContact", testGetContact)

	resetDatabase()
	t.Run("UpdateContact", testUpdateContact)

	resetDatabase()
	t.Run("DeleteContact", testDeleteContact)

	resetDatabase()
	t.Run("SearchContacts", testSearchContacts)
}

func testCreateContact(t *testing.T) {
	contact := contacts.Contact{
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+1234567890",
		Address:   "123 Main St",
	}

	contactJSON, err := json.Marshal(contact)
	assert.NoError(t, err)

	resp, err := http.Post(testServer.URL+"/addContact", "application/json", bytes.NewBuffer(contactJSON))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func testGetContact(t *testing.T) {
	// Create a contact first
	contact := contacts.Contact{
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+1234567890",
		Address:   "123 Main St",
	}

	contactJSON, err := json.Marshal(contact)
	assert.NoError(t, err)

	resp, err := http.Post(testServer.URL+"/addContact", "application/json", bytes.NewBuffer(contactJSON))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Retrieve the contact
	resp, err = http.Get(testServer.URL + "/getContacts?page=1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var paginatedContacts contacts.PaginatedContacts
	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(paginatedContacts.Contacts), 1)
}

func testUpdateContact(t *testing.T) {
	// Create a contact first
	contact := contacts.Contact{
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+1234567890",
		Address:   "123 Main St",
	}

	contactJSON, err := json.Marshal(contact)
	assert.NoError(t, err)

	resp, err := http.Post(testServer.URL+"/addContact", "application/json", bytes.NewBuffer(contactJSON))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Update the contact
	updatedContact := contacts.Contact{
		FirstName: "Jane",
		LastName:  "Doe",
		Phone:     "+0987654321",
		Address:   "456 Elm St",
	}

	updatedContactJSON, err := json.Marshal(updatedContact)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/updateContact/1", bytes.NewBuffer(updatedContactJSON))
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func testDeleteContact(t *testing.T) {
	// Create a contact first
	contact := contacts.Contact{
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+1234567890",
		Address:   "123 Main St",
	}

	contactJSON, err := json.Marshal(contact)
	assert.NoError(t, err)

	resp, err := http.Post(testServer.URL+"/addContact", "application/json", bytes.NewBuffer(contactJSON))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Delete the contact
	req, err := http.NewRequest(http.MethodDelete, testServer.URL+"/deleteContact/1", nil)
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func testSearchContacts(t *testing.T) {
	// Create 35 contacts
	contactsToCreate := []contacts.Contact{
		{FirstName: "John", LastName: "Doe", Phone: "+1234567891", Address: "123 Main St"},
		{FirstName: "Jane", LastName: "Smith", Phone: "+1234567892", Address: "456 Elm St"},
	}

	// Fill up the list with a few more unique contacts
	for i := 3; i <= 35; i++ {
		contact := contacts.Contact{
			FirstName: fmt.Sprintf("Person%d", i),
			LastName:  fmt.Sprintf("Last%d", i),
			Phone:     fmt.Sprintf("+12345678%02d", i),
			Address:   fmt.Sprintf("%d Street", i),
		}
		contactsToCreate = append(contactsToCreate, contact)
	}

	// Add contacts to the DB
	for _, contact := range contactsToCreate {
		contactJSON, err := json.Marshal(contact)
		if err != nil {
			internal.Logger.Error(fmt.Sprintf("Error in marshalling json: %v", err))
		}
		assert.NoError(t, err)

		resp, err := http.Post(testServer.URL+"/addContact", "application/json", bytes.NewBuffer(contactJSON))
		if err != nil {
			internal.Logger.Error(fmt.Sprintf("Error in adding contact to DB for test: %v", err))
		}
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// // Test pagination: Get the first page (10 contacts)
	// resp, err := http.Get(testServer.URL + "/getContacts?page=1")
	// assert.NoError(t, err)
	// assert.Equal(t, http.StatusOK, resp.StatusCode)

	// var paginatedContacts contacts.PaginatedContacts
	// err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	// assert.NoError(t, err)
	// assert.Equal(t, 10, len(paginatedContacts.Contacts))
	// assert.Equal(t, 1, paginatedContacts.TotalPages) // Expecting 1 page (10 contacts total)

	// // Test filtering by first name: Search for "John"
	// resp, err = http.Get(testServer.URL + "/getContacts?page=1&first_name=John")
	// assert.NoError(t, err)
	// assert.Equal(t, http.StatusOK, resp.StatusCode)

	// err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	// assert.NoError(t, err)
	// assert.Equal(t, 1, len(paginatedContacts.Contacts))
	// assert.Equal(t, "John", paginatedContacts.Contacts[0].FirstName)

	// // Test filtering by last name: Search for "Doe"
	// resp, err = http.Get(testServer.URL + "/getContacts?page=1&last_name=Doe")
	// assert.NoError(t, err)
	// assert.Equal(t, http.StatusOK, resp.StatusCode)

	// err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	// assert.NoError(t, err)
	// assert.Equal(t, 1, len(paginatedContacts.Contacts))
	// assert.Equal(t, "Doe", paginatedContacts.Contacts[0].LastName)

	// // Test filtering with a name that should not exist
	// resp, err = http.Get(testServer.URL + "/getContacts?page=1&first_name=Nonexistent")
	// assert.NoError(t, err)
	// assert.Equal(t, http.StatusOK, resp.StatusCode)

	// err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	// assert.NoError(t, err)
	// assert.Equal(t, 0, len(paginatedContacts.Contacts)) // Should return 0 results
	// assert.Equal(t, 0, paginatedContacts.TotalPages)    // Should return 0 pages
}

func setupTestServer() {
	// Initialize the test server once for all tests
	if testServer == nil {
		router := setupRouter()
		testServer = httptest.NewServer(router)
	}

	// Reset the database for each test
	resetDatabase()
}

func resetDatabase() {
	db, err := db.DBInit()
	if err != nil {
		internal.Logger.Error("Failed to initialize test database")
		panic(err)
	}

	// Drop and recreate the schema
	db.Exec("DROP SCHEMA public CASCADE;")
	db.Exec("CREATE SCHEMA public;")

	// Run migrations to create the table
	err = db.AutoMigrate(&contacts.Contact{})
	if err != nil {
		internal.Logger.Error("Failed to migrate schema for test database")
		panic(err)
	}
}
