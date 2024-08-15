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
	return nil
}

func getContacts(page int) {

}

func updateContact() {

}

func deleteContacts() {

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
