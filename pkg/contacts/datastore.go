// Handle CRUD ops for the database
package contacts

import "golangphonebook/internal"

// Initialize contact list as slice, temporary solution for now
var MyContactList = ContactList{
	contacts: []Contact{},
	count:    0,
}

func addContact(contact Contact) error {
	internal.Logger.Info("Made it to the add method!")
	// Check for duplicates here, by same name and phone number
	return nil
}

func getContacts(page int) {

}

func updateContact(contact Contact) error {
	internal.Logger.Info("Made it to the update method!")
	return nil
}

func deleteContact(id int) error {
	internal.Logger.Info("Made it to the delete contact method")
	return nil
}

func mergeDuplicates() {

}

func getSize() int {
	return 100
}

func contactExists() bool {
	// TODO placeholder
	return true
}
