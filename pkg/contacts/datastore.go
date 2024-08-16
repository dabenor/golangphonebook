// Handle CRUD ops for the database
package contacts

import (
	"errors"
	"fmt"
	"golangphonebook/internal"

	"gorm.io/gorm"
)

// Initialize contact list as slice, temporary solution for now
var MyContactList = ContactList{
	contacts: []Contact{},
	count:    0,
}

type SQLContactRepository struct {
	DB *gorm.DB
}

// NewSQLContactRepository creates a new instance of SQLContactRepository
func NewSQLContactRepository(db *gorm.DB) *SQLContactRepository {
	return &SQLContactRepository{DB: db}
}

func (repo *SQLContactRepository) AddContact(contact Contact) error {
	var existingContact Contact

	// Check if a contact with the same FirstName, LastName and Phone already exists
	err := repo.DB.Where("first_name = ? AND last_name = ? AND phone = ?", contact.FirstName, contact.LastName, contact.Phone).First(&existingContact).Error
	if err == nil {
		// Contact already exists
		internal.Logger.Warn("contact with the same full name and phone number already exists")
		return errors.New("contact with the same full name and phone number already exists")
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Contact does not exist
		err = repo.DB.Create(&contact).Error
		return err
	} else {
		// We got some other error
		return err

	}
}

func (repo *SQLContactRepository) GetContacts(page int) error {
	internal.Logger.Info("Made it to getContacts")
	return nil

}

func (repo *SQLContactRepository) GetContact(contact Contact) error {
	internal.Logger.Info("Made it to getContacts")
	return nil

}

func (repo *SQLContactRepository) GetAllContacts() {
	var contacts []Contact

	// Query the database to retrieve all contacts
	err := repo.DB.Find(&contacts).Error
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Error on getting contacts %v", err))
	}

	internal.Logger.Info(fmt.Sprintf("Retrieved %d contacts", len(contacts)))
	for _, contact := range contacts {
		fmt.Println(contact)
	}
	return
}

func (repo *SQLContactRepository) UpdateContact(contact Contact) error {
	internal.Logger.Info("Made it to the update method!")
	return nil
}

func (repo *SQLContactRepository) DeleteContact(id int) error {
	result := repo.DB.Delete(&Contact{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no contact found with the given ID")
	}

	internal.Logger.Info(fmt.Sprintf("Contact deleted successfully, %d row(s) affected", result.RowsAffected))

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
