package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golangphonebook/db"
	"golangphonebook/internal"
	"golangphonebook/pkg/contacts"
	"io"
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
	router.HandleFunc("/addContacts", func(w http.ResponseWriter, r *http.Request) { contacts.PutContacts(w, r, repo) }).Methods("POST")
	// R
	router.HandleFunc("/getContacts", func(w http.ResponseWriter, r *http.Request) { contacts.GetContacts(w, r, repo) }).Methods("GET")
	// U
	router.HandleFunc("/updateContact/{id}", func(w http.ResponseWriter, r *http.Request) { contacts.UpdateContact(w, r, repo) }).Methods("POST")
	// D
	router.HandleFunc("/deleteContact/{id}", func(w http.ResponseWriter, r *http.Request) { contacts.DeleteContact(w, r, repo) }).Methods("DELETE")
	router.HandleFunc("/deleteContacts", func(w http.ResponseWriter, r *http.Request) { contacts.DeleteContacts(w, r, repo) }).Methods("DELETE")
	return router
}

func TestE2E(t *testing.T) {
	setupTestServer()
	t.Run("CreateContact", testCreateContact)

	setupTestServer()
	t.Run("PutContacts", testPutContacts)

	resetDatabase()
	t.Run("GetContact", testGetContact)

	resetDatabase()
	t.Run("UpdateContact", testUpdateContact)

	resetDatabase()
	t.Run("DeleteContact", testDeleteContact)

	resetDatabase()
	t.Run("DeleteContacts", testDeleteContacts)

	resetDatabase()
	t.Run("SearchContacts", testSearchContacts)

	resetDatabase()
	t.Run("SearchContactsWithUpdates", testSearchContactsWithUpdates)
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

	// First attempt to create the contact
	resp, err := http.Post(testServer.URL+"/addContact", "application/json", bytes.NewBuffer(contactJSON))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Try entering the same contact again
	resp, err = http.Post(testServer.URL+"/addContact", "application/json", bytes.NewBuffer(contactJSON))
	assert.NoError(t, err)                                  // No network or request error
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // Expecting a 400 Bad Request

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	// Check if the body contains the expected error message
	assert.Contains(t, string(body), "contact with the same full name and phone number already exists")

	// Create a second contact with a different first name, should enter fine.
	// People have the same numbers sometimes (work number, some couples have calls routed to each of their phones)
	contact.FirstName = "Allison"
	contactJSON, err = json.Marshal(contact)
	assert.NoError(t, err)

	resp, err = http.Post(testServer.URL+"/addContact", "application/json", bytes.NewBuffer(contactJSON))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func testPutContacts(t *testing.T) {
	router := setupRouter()

	// Define valid contacts
	validContacts := []contacts.Contact{
		{
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
			Address:   "123 Main St",
		},
		{
			FirstName: "Jane",
			LastName:  "Smith",
			Phone:     "+9876543210",
			Address:   "456 Elm St",
		},
		{
			FirstName: "Emily",
			LastName:  "Johnson",
			Phone:     "+1122334455",
			Address:   "789 Oak St",
		},
	}

	// Define an invalid contact (missing FirstName and Phone)
	invalidContact := contacts.Contact{
		LastName: "Doe",
		Address:  "No Address",
	}

	// Convert valid contacts to JSON
	validContactsJSON, err := json.Marshal(validContacts)
	assert.NoError(t, err)

	// Test adding valid contacts
	t.Run("AddValidContacts", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/addContacts", bytes.NewBuffer(validContactsJSON))
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result map[string]interface{}
		err = json.NewDecoder(rr.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, 3.0, result["successful_contacts"]) // JSON decoding converts numbers to float64
		assert.Empty(t, result["failed_contacts"])
	})
	// If you don't reset the db then all of the contacts are invalid, since they're already in the DB
	resetDatabase()
	// Combine valid and invalid contacts for testing
	mixedContacts := append(validContacts[:2], invalidContact)
	mixedContactsJSON, err := json.Marshal(mixedContacts)
	assert.NoError(t, err)

	// // Test adding mixed valid and invalid contacts
	t.Run("AddMixedContacts", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/addContacts", bytes.NewBuffer(mixedContactsJSON))
		assert.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusPartialContent, rr.Code)

		var result map[string]interface{}
		err = json.NewDecoder(rr.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, 2.0, result["successful_contacts"]) // Two valid contacts should be added successfully
		assert.NotEmpty(t, result["failed_contacts"])
		assert.NotEmpty(t, result["errors"])

		// Verify that the invalid contact caused the failure
		failedContacts := result["failed_contacts"].([]interface{})
		assert.Equal(t, 1, len(failedContacts))
		failedContactJSON, _ := json.Marshal(invalidContact)
		assert.Contains(t, string(failedContacts[0].(string)), string(failedContactJSON))
	})
}

func testGetContact(t *testing.T) {
	// Retrieve nonexistent contact
	resp, err := http.Get(testServer.URL + "/getContacts?page=1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var paginatedContacts contacts.PaginatedContacts
	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, len(paginatedContacts.Contacts), 0)

	// Create a contact first
	contact := contacts.Contact{
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+1234567890",
		Address:   "123 Main St",
	}

	contactJSON, err := json.Marshal(contact)
	assert.NoError(t, err)

	resp, err = http.Post(testServer.URL+"/addContact", "application/json", bytes.NewBuffer(contactJSON))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Retrieve the contact
	resp, err = http.Get(testServer.URL + "/getContacts?page=1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

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

func testDeleteContacts(t *testing.T) {
	router := setupRouter()

	// Step 1: Add four contacts to have something to delete
	t.Run("AddContactsForDeletion", func(t *testing.T) {
		contactsToAdd := []contacts.Contact{
			{
				FirstName: "John",
				LastName:  "Doe",
				Phone:     "+1234567890",
				Address:   "123 Main St",
			},
			{
				FirstName: "Jane",
				LastName:  "Smith",
				Phone:     "+9876543210",
				Address:   "456 Elm St",
			},
			{
				FirstName: "Emily",
				LastName:  "Johnson",
				Phone:     "+1122334455",
				Address:   "789 Oak St",
			},
			{
				FirstName: "Alice",
				LastName:  "Brown",
				Phone:     "+4455667788",
				Address:   "101 Pine St",
			},
		}

		contactsJSON, err := json.Marshal(contactsToAdd)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", "/addContacts", bytes.NewBuffer(contactsJSON))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	// Step 2: Test deleting contacts 1, 2, 3
	t.Run("DeleteValidContacts", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "/deleteContacts?ids=1,2,3", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "Contacts deleted successfully", rr.Body.String())
	})

	// Step 3: Test deleting contacts 1, 4 (with 1 failing)
	t.Run("DeleteMixedContacts", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "/deleteContacts?ids=1,4", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "No contact found with ID 1")
	})
}

func testSearchContacts(t *testing.T) {
	// Create 35 contacts
	contactsToCreate := []contacts.Contact{
		{FirstName: "John", LastName: "Doe", Phone: "+1234567891", Address: "123 Main St"},
		{FirstName: "Jane", LastName: "Smith", Phone: "+1234567892", Address: "456 Elm St"},
	}

	// Fill up the list with 33 more unique contacts (total 35)
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
		assert.NoError(t, err)

		resp, err := http.Post(testServer.URL+"/addContact", "application/json", bytes.NewBuffer(contactJSON))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Test pagination: Get the first page (10 contacts)
	resp, err := http.Get(testServer.URL + "/getContacts?page=1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var paginatedContacts contacts.PaginatedContacts
	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(paginatedContacts.Contacts))
	assert.Equal(t, 4, paginatedContacts.TotalPages) // Expecting 4 pages (35 contacts total)

	// Test pagination: Get the second page (next 10 contacts)
	resp, err = http.Get(testServer.URL + "/getContacts?page=2")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(paginatedContacts.Contacts))
	assert.Equal(t, 4, paginatedContacts.TotalPages) // Still expecting 4 pages

	// Test pagination: Get the third page (next 10 contacts)
	resp, err = http.Get(testServer.URL + "/getContacts?page=3")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(paginatedContacts.Contacts))
	assert.Equal(t, 4, paginatedContacts.TotalPages) // Still expecting 4 pages

	// Test pagination: Get the fourth page (remaining 5 contacts)
	resp, err = http.Get(testServer.URL + "/getContacts?page=4")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(paginatedContacts.Contacts)) // Remaining 5 contacts
	assert.Equal(t, 4, paginatedContacts.TotalPages)    // Expecting 4 pages

	// Test filtering by first name: Search for "John"
	resp, err = http.Get(testServer.URL + "/getContacts?page=1&first_name=John")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(paginatedContacts.Contacts))
	assert.Equal(t, "John", paginatedContacts.Contacts[0].FirstName)

	// Test filtering by last name: Search for "Doe"
	resp, err = http.Get(testServer.URL + "/getContacts?page=1&last_name=Doe")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(paginatedContacts.Contacts))
	assert.Equal(t, "Doe", paginatedContacts.Contacts[0].LastName)

	// Test pagination with a filter: Search for "Person" in first name (should get multiple pages)
	resp, err = http.Get(testServer.URL + "/getContacts?page=1&first_name=Person")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(paginatedContacts.Contacts)) // Should return the first 10 matching contacts
	assert.Equal(t, 4, paginatedContacts.TotalPages)     // Expecting 4 pages for "Person" results

	// Test second page of filtered results for "Person"
	resp, err = http.Get(testServer.URL + "/getContacts?page=2&first_name=Person")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(paginatedContacts.Contacts)) // Another 10 matching contacts
	assert.Equal(t, 4, paginatedContacts.TotalPages)     // Expecting 4 pages

	// Test third page of filtered results for "Person"
	resp, err = http.Get(testServer.URL + "/getContacts?page=3&first_name=Person")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(paginatedContacts.Contacts)) // next 10 matching contacts
	assert.Equal(t, 4, paginatedContacts.TotalPages)     // Expecting 4 pages

	// Test fourth page of filtered results for "Person"
	resp, err = http.Get(testServer.URL + "/getContacts?page=4&first_name=Person")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(paginatedContacts.Contacts)) // Remaining 3 matching contacts
	assert.Equal(t, 4, paginatedContacts.TotalPages)    // Expecting 4 pages

	// Test second page of filtered results for "Person" again
	resp, err = http.Get(testServer.URL + "/getContacts?page=2&first_name=Person")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(paginatedContacts.Contacts)) // Another 10 matching contacts
	assert.Equal(t, 4, paginatedContacts.TotalPages)     // Expecting 4 pages

	// Jump to fourth page of filtered results for "Person" (not using cache)
	resp, err = http.Get(testServer.URL + "/getContacts?page=4&first_name=Person")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(paginatedContacts.Contacts)) // Remaining 3 matching contacts
	assert.Equal(t, 4, paginatedContacts.TotalPages)    // Expecting 4 pages
}

func testSearchContactsWithUpdates(t *testing.T) {
	// Create 35 contacts
	contactsToCreate := []contacts.Contact{
		{FirstName: "John", LastName: "Doe", Phone: "+1234567891", Address: "123 Main St"},
		{FirstName: "Jane", LastName: "Smith", Phone: "+1234567892", Address: "456 Elm St"},
	}

	// Fill up the list with 29 more unique contacts (total 31)
	for i := 3; i <= 31; i++ {
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
		assert.NoError(t, err)

		resp, err := http.Post(testServer.URL+"/addContact", "application/json", bytes.NewBuffer(contactJSON))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Test pagination: Get the first page (10 contacts)
	resp, err := http.Get(testServer.URL + "/getContacts?page=1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var paginatedContacts contacts.PaginatedContacts
	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(paginatedContacts.Contacts))
	assert.Equal(t, 4, paginatedContacts.TotalPages) // Expecting 4 pages (31 contacts total)

	// Test pagination: test negative page
	resp, err = http.Get(testServer.URL + "/getContacts?page=-1423")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test negative page, should receive page 1
	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(paginatedContacts.Contacts))
	assert.Equal(t, 4, paginatedContacts.TotalPages)  // Expecting 4 pages (31 contacts total)
	assert.Equal(t, 1, paginatedContacts.CurrentPage) // Check that page 1 is served in this case

	// Test pagination: test invalid format page
	resp, err = http.Get(testServer.URL + "/getContacts?page=hahathisshouldjustgivemepage1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test negative page, should receive page 1
	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(paginatedContacts.Contacts))
	assert.Equal(t, 4, paginatedContacts.TotalPages)  // Expecting 4 pages (31 contacts total)
	assert.Equal(t, 1, paginatedContacts.CurrentPage) // Check that page 1 is served in this case

	// Test pagination: get page 4, then delete a thing to bring it to 3 pages. It should return page 1
	resp, err = http.Get(testServer.URL + "/getContacts?page=4")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Should receive page 4
	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(paginatedContacts.Contacts)) // Page 4 has 1 contact on it currently
	assert.Equal(t, 4, paginatedContacts.TotalPages)    // Expecting 4 pages (31 contacts total)
	assert.Equal(t, 4, paginatedContacts.CurrentPage)   // Check that page 1 is served in this case

	// Delete a contact :(, bye bye Person3
	req, err := http.NewRequest(http.MethodDelete, testServer.URL+"/deleteContact/3", nil)
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test pagination: Get the second page (next 10 contacts), should not refer to cache
	resp, err = http.Get(testServer.URL + "/getContacts?page=4")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&paginatedContacts)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(paginatedContacts.Contacts)) // 10 contacts since it's returning page 1
	assert.Equal(t, 3, paginatedContacts.TotalPages)     // Now expecting 3 pages for 30 contacts
	assert.Equal(t, 1, paginatedContacts.CurrentPage)    // Check that page 1 is served in this case

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
