# Phonebook API Documentation

Experimenting with Go to create a WebServer API with basic CRUD operations

## Setup

To build and test this project, run the following from the root directory
```bash
docker-compose up --build test
```
Next, close the docker build once the tests are complete by pressing Ctrl+C, and then removing the test containers with the following command

```bash
docker-compose down
```
To then run the resulting Docker build and expose port 8443, please run the following, also from the root directory

```bash
docker-compose up db phonebook
```

This application is secured using ca signed certificates, so you'll need to import those into Postman. These certs were generated for this project only and are not meant to be used anywhere else. That would not be secure :)

The certificate files are located at the paths /certs/ca.crt and /certs/ca.key, you should set them to be used when https://localhost:8443/ is hit from Postman. I provide a Postman collection in the root directory here that you can import to Postman that has all of the requests and parameters you could pass into this API.


Then you can access the application by sending CURL requests to [https://localhost:8443/](https://localhost:8443/) from the terminal, via [Postman](https://www.postman.com/), or you can navigate to the same URL in your browser.


## Table of Contents

1. [Some Basic Constraints](#constraints)
2. [Endpoints](#endpoints)
    - [Add Contact](#add-contact)
    - [Add Contacts](#add-contacts)
    - [Get Contacts](#get-contacts)
    - [Update Contact](#update-contact)
    - [Delete Contact](#delete-contact)
    - [Delete Contacts](#delete-contacts)
  

## Constraints

    - first_name: Pretty open, needs to exist on any JSON calls to update or add a contact
    - last_name: Pretty open, optional in most cases
    - phone: An optional + sign followed by between 4 and 20 digits 0-9
    - address: Pretty open, optional in most cases
    - sort_by: only for getContacts function, can be first_name, last_name, or last_modified depedning on how you want to sort your results
    - asc_dec: only for getContacts function, can be asc or dec depending on whether you want results to be ascending or descending, by default ascending, except for last_modified, which by default is descending (so you can see most recent contacts)
    - page: only for getContacts function, used to tell the server what page of the results you want. If undefined or out of bounds due to a filter will automatically be set to 1

## Endpoints

---

### Add Contact

- **Endpoint**: `/addContact`
- **Method**: PUT
- **Description**: Creates multiple new contacts based on the provided JSON body.

#### Request Body

- An array of JSON objects representing the contacts to add. Each object should include at least the 'first_name' and 'phone' fields. Optional fields that can also be populated later using an [Update Contact](#update-contact) call are 'last_name' and 'address'. The first_name, last_name, and phone cannot be the same as a contact already in the database.

**Example Request Body**:

```json
{
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "address": "123 Main St"
}
```

**Responses:**
- 200 OK: Contact added successfully.
- 400 Bad Request: Invalid request body, first name and phone must be correctly defined
- 400 Bad Request: contact with the same full name and phone number already exists
- 500 Internal Server Error: Failed to insert contact into the database.


### Add Contacts

- **Endpoint**: `/addContacts`
- **Method**: PUT
- **Description**: Creates multiple new contacts based on the provided JSON array.

#### Request Body

- An array of JSON objects representing the contacts to add. Each object should include 'first_name', 'last_name', and 'phone' fields.

**Example Request Body**:

```json
[
  {
    "first_name": "John",
    "last_name": "Doe",
    "phone": "1234567890"
  },
  {
    "first_name": "Jane",
    "last_name": "Smith",
    "phone": "9876543210"
  }
]
```
- 200 OK: Contacts added successfully.
- 400 Bad Request: Invalid request body. Please provide a valid JSON array of contacts.
- 400 Bad Request: Cannot add more than 20 contacts at a time.
- 400 Bad Request: Failed to create request for contact
- 400 Bad Request: Failed to create request for contact
