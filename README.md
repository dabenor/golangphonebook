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

The certificate files are located at the paths /certs/client.crt and /certs/client.key, you should set them to be used when https://localhost:8443/ is hit from Postman. I provide a Postman collection in the root directory here that you can import to Postman that has all of the requests and parameters you could pass into this API.


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
- **Description**: Creates up to 20 new contacts based on the provided JSON array.

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

### Get Contacts

- **Endpoint**: `/getContacts`
- **Method**: GET
- **Description**: Filter and Search through contacts in the Phonebook.

#### Request Body

- blank/ignored

#### Parameters

Filter Parameters
- first_name
- last_name
- phone
- address

Pagination/Sorting Parameters
- page (default value is 1, can be any value up to the number of pages for the filter)
- sort_by ("first_name", "last_name" or "last_modified")
- asc_dec ("asc" or "dec" for ascending or descending sort)

**Example Request Parameters**:

Request 1: Find the first page of contacts with names like John, sorted descending by first name. If you pass in page=2, and there is only 1 page of results, you will receive page 1
first_name=john
page=1
sort_by=first_name
asc_dec=dec

Request 2: Find the people who live on main street, even without passing in a page parameter I will receive page 1 of the results
address=main street

- 200 OK: Contacts added successfully.
- 400 Bad Request: Invalid request body. Please provide a valid JSON array of contacts.
- 400 Bad Request: Cannot add more than 20 contacts at a time.
- 400 Bad Request: Failed to create request for contact


Response Format:
You will receive an array of no more than 10 contacts, followed by pagination metadata. 
- total_pages is the number of pages of contacts for the current query
- current_page is the page of results returned to the client
- total_count is the total number of records in the Phonebook that match the filters. That's only affected by adjusting the filter parameters
```json
{
    "contacts": [
        {
            "id": 62,
            "first_name": "Alice",
            "last_name": "Wonderland",
            "phone": "+3422220456",
            "address": "456 Elm St",
            "last_modified": "2024-08-18T23:02:29.101933Z"
        },
        {
            "id": 63,
            "first_name": "Bob",
            "last_name": "Builder",
            "phone": "+3422220789",
            "address": "789 Maple St",
            "last_modified": "2024-08-18T23:02:29.105321Z"
        },
        {
            "id": 64,
            "first_name": "Charlie",
            "last_name": "Chaplin",
            "phone": "+3422220110",
            "address": "101 Oak St",
            "last_modified": "2024-08-18T23:02:29.108775Z"
        },
        // 10 of these, cutting it off here for readability
    ],
    "total_pages": 5,
    "current_page": 1,
    "total_count": 41
}
```

### Update Contact

- **Endpoint**: `/updateContact/{id}`
- **Method**: POST
- **Description**: Update the contact with the specified ID.

#### Request Body

- An JSON contact to add. The JSON object should include at least the 'first_name' and 'phone' fields. Optional fields that can also be populated later are 'last_name' and 'address'. The first_name, last_name, and phone cannot be the same as a contact already in the database. You receive the IDs of a contact to update from the [Get Contacts](#get-contacts) endpoint. The ID is set by the database, and is unique to each contact. This way, you can be sure you are updating the right contact in the database.

**Example Request URL**:

To update contact 3, send a request to this URL with an updated Contact request body as shown below
https://localhost:8443/updateContact/3

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
- 200 OK: Contact updated successfully.
- 400 Bad Request: Invalid request body, first name and phone must be correctly defined
- 400 Bad Request: another contact with the same first name, last name, and phone number already exists
- 500 Internal Server Error: Failed to update contact due to an internal server error

### Delete Contact

- **Endpoint**: `/deleteContact/{id}`
- **Method**: DELETE
- **Description**: Delete the contact with the specified ID. ID must be an integer

#### Request Body

- Blank/ignored

**Example Request URL**:

To delete contact with ID 3, send a request to this URL
https://localhost:8443/deleteContact/3


**Responses:**
- 200 OK: Contact deleted successfully.
- 400 Bad Request: Invalid ID, IDs can only be integers
- 404 Not Found: no contact found with the given ID
- 500 Internal Server Error: failed to delete contact


### Delete Contacts

- **Endpoint**: `/deleteContacts/{ids}`
- **Method**: DELETE
- **Description**: Delete up to 20 contacts at once based on the the list of comma separated ints passed in as {ids} in the URL.

This function is less tolerant than [Add Contacts](#add-contacts) because we want to be sure that the user knows what they're deleting. Also, there is more room for error in the `/addContacts` endpoint above, as those have elaborate json requirements that do not exist for `/deleteContacts`. Once we encounter an invalid ID, we abort the method and return an error. We delete until that point though.

Say you pass in IDs 3, 5, 7, and 10, and ID 7 is not in the DB, the `/deleteContacts` method will only remove IDs 3 and 5, even if 10 is a valid ID to delete.

#### Request Body

- blank/ignored

**Example Request URL**:
To delete IDs 3, 5, 7, and 10, pass the following into the service
https://localhost:8443/deleteContact/3,5,7,10


- 200 OK: Contacts added successfully.
- 400 Bad Request: No IDs provided
- 400 Bad Request: Cannot delete more than 20 contacts at a time.
- 400 Bad Request: Invalid IDs: {list of invalid IDs}. IDs can only be integers.
- 404 Bad Request: No contact found with ID {id}
- 500 Internal Server Error: Failed to delete contact with ID {id}
