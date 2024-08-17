// handle HTTP function requests GET POST DELETE etc.
package contacts

import (
	"encoding/json"
	"fmt"
	"golangphonebook/internal"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func PutContact(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	defer internal.Timer("PutContact")()

	contact, err := decodeBodyToContact(r)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Received invalid body in addContact method %s", err))
		http.Error(w, "Invalid request body, first name and phone must be correctly defined", http.StatusBadRequest)
		return
	} else {
		internal.Logger.Info(fmt.Sprintf("Received valid body in addContact method %s", contact))
	}

	err = repo.AddContact(*contact)
	if err != nil {
		if err.Error() == "contact with the same full name and phone number already exists" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, fmt.Sprintf("Failed to insert contact to db with error %v", err), http.StatusInternalServerError)
		}
		return
	}

	internal.Logger.Info("Contact added to DB successfully")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Contact added to DB successfully"))
	// Return an updated first page, maybe. Return contact itself. Or when user adds a contact, send them back to their origin page, or to page 1 of their contacts
}

func GetContacts(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	// default page is page 1
	page := 1

	pageHeader := r.Header.Get("page")
	if parsedPage, err := strconv.Atoi(pageHeader); err == nil {
		// TODO: Add logic for upper bound of parsed page: Old logic && parsedPage <= repo.GetContactCount()/11+1
		if parsedPage >= 1 {
			page = parsedPage
		} else {
			internal.Logger.Warn("Invalid page number, defaulting to page 1")
		}
	}

	repo.GetContacts(page)
}

func GetAllContacts(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	repo.GetAllContacts()
}

func UpdateContact(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	// Extract ID from  URL path /deleteContact/{id}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(w, "Invalid ID, IDs can only be integers", http.StatusBadRequest)
		return
	}
	internal.Logger.Info(fmt.Sprintf("ID to update detected as %d", id))

	contact, err := decodeBodyToContact(r)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Received invalid body in updateContact method %s", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	internal.Logger.Info(fmt.Sprintf("Received valid body in updateContact method %s", contact))

	// Update contact in db
	err = repo.UpdateContact(id, *contact)
	if err != nil {
		// Handle duplicate data error
		if err.Error() == "another contact with the same first name, last name, and phone number already exists" {
			internal.Logger.Error(fmt.Sprintf("Duplicate data error: %s", err))
			http.Error(w, "Duplicate contact with the same first name, last name, and phone number already exists", http.StatusBadRequest)
			return
		}
		// Handle other errors as internal server errors
		internal.Logger.Error(fmt.Sprintf("Failed to update contact to db: %s", err))
		http.Error(w, "Failed to update contact due to an internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Contact updated successfully"))
}

func DeleteContact(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	defer internal.Timer("DeleteContact")()

	// Extract ID from  URL path /deleteContact/{id}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(w, "Invalid ID, IDs can only be integers", http.StatusBadRequest)
		return
	}
	internal.Logger.Info(fmt.Sprintf("ID to delete detected as %d", id))

	err = repo.DeleteContact(id)
	if err != nil {
		if err.Error() == "no contact found with the given ID" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete contact", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Contact deleted successfully"))
}

// Helper method(s)
// Decode JSON body into a Contact
func decodeBodyToContact(r *http.Request) (*Contact, error) {
	// Read body from request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Unable to read request body: %v", err))
		return nil, fmt.Errorf("unable to read request body: %v", err)
	}
	defer r.Body.Close()
	internal.Logger.Info(fmt.Sprintf("Received JSON: %s", string(body)))

	// Decode JSON body into a Contact
	var contact Contact
	if err := json.Unmarshal(body, &contact); err != nil {
		internal.Logger.Error(fmt.Sprintf("Unable to unmarshal JSON into Contact: %v", err))
		return nil, fmt.Errorf("unable to unmarshal JSON into Contact: %v", err)
	}

	// Validate the Contact struct
	err = validate.Struct(contact)
	if err != nil {
		return nil, fmt.Errorf("Err(s):\n%+v", err)
	}
	// Return Contact object
	return &contact, nil
}
