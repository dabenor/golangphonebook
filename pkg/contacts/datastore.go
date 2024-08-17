// Handle CRUD ops for the database
package contacts

import (
	"errors"
	"fmt"
	"golangphonebook/internal"

	"gorm.io/gorm"
)

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

func (repo *SQLContactRepository) UpdateContact(id int, updatedContact Contact) error {
	// Check if contact exists
	var existingContact Contact
	err := repo.DB.First(&existingContact, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("contact not found")
		}
		return err
	}

	// Check for duplicate contact
	var duplicateContact Contact
	err = repo.DB.Where("first_name = ? AND last_name = ? AND id != ? AND phone = ?",
		updatedContact.FirstName, updatedContact.LastName, id, updatedContact.Phone).First(&duplicateContact).Error
	if err == nil {
		// Duplicate exists
		return errors.New("another contact with the same first name, last name, and phone number already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other error
		return err
	}
	// Update fields
	if updatedContact.FirstName != "" {
		existingContact.FirstName = updatedContact.FirstName
	}
	if updatedContact.LastName != "" {
		existingContact.LastName = updatedContact.LastName
	}
	if updatedContact.Phone != "" {
		existingContact.Phone = updatedContact.Phone
	}
	if updatedContact.Address != "" {
		existingContact.Address = updatedContact.Address
	}

	// Save contact back to db
	err = repo.DB.Save(&existingContact).Error
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Encountered err while saving updated contact back to DB: %v", err))
		return err
	}

	internal.Logger.Info(fmt.Sprintf("Contact with ID %d updated successfully", id))
	return nil
}

func (repo *SQLContactRepository) DeleteContact(id int) error {
	result := repo.DB.Delete(&Contact{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		internal.Logger.Error(fmt.Sprintf("no contact found with ID: %d", id))
		return errors.New("no contact found with the given ID")
	}

	internal.Logger.Info(fmt.Sprintf("Contact deleted successfully, %d row(s) affected", result.RowsAffected))

	return nil
}

func mergeDuplicates() {

}

// Helper methods
func (repo *SQLContactRepository) GetContactCount() (int64, error) {
	var count int64
	err := repo.DB.Model(&Contact{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func contactExists() bool {
	// TODO placeholder
	return true
}
