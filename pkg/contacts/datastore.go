// Handle CRUD ops for the database
package contacts

import (
	"database/sql"
	"golangphonebook/internal"
)

// Initialize contact list as slice, temporary solution for now
var MyContactList = ContactList{
	contacts: []Contact{},
	count:    0,
}

type SQLContactRepository struct {
	DB *sql.DB
}

// NewSQLContactRepository creates a new instance of SQLContactRepository
func NewSQLContactRepository(db *sql.DB) *SQLContactRepository {
	return &SQLContactRepository{DB: db}
}

func (repo *SQLContactRepository) AddContact(contact Contact) error {
	query := `INSERT INTO contacts (first_name, last_name, phone, address) VALUES (?, ?, ?, ?)`
	_, err := repo.DB.Exec(query, contact.FirstName, contact.LastName, contact.Phone, contact.Address)
	return err
}

func (repo *SQLContactRepository) GetContacts(page int) error {
	internal.Logger.Info("Made it to getContacts")
	return nil

}

func (repo *SQLContactRepository) UpdateContact(contact Contact) error {
	internal.Logger.Info("Made it to the update method!")
	return nil
}

func (repo *SQLContactRepository) DeleteContact(id int) error {
	internal.Logger.Info("Made it to the delete contact method")
	return nil
}

func mergeDuplicates() {

}

// Helper methods
func getSize() int {
	return 100
}

func contactExists() bool {
	// TODO placeholder
	return true
}
