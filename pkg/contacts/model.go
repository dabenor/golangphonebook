// Define structures
package contacts

type Contact struct {
	id        int
	FirstName string
	LastName  string
	Phone     string
	Address   string
}

type ContactList struct {
	contacts []Contact
	count    int
}

// Addtl. structures as needed will go here
