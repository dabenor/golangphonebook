# golangphonebook
Experimenting with Go to create a WebServer API with basic CRUD operations

For now, placed basic skeleton of the web server into the file structure. Will set up a basic entry point, configure the dockerfile, and then work on creating the data storage model. At first I will create it as an in-memory storage solution, and once I complete that, I will set up a database or a save state option for the service.

To build this project, run the following from the root directory
```
docker build --tag docker-phonebook .
```

To then run the resulting Docker build and expose port 8080, run the following, also from the root directory

```
docker run -p 8080:8080 docker-phonebook
```

Then you can access the application by sending CURL requests to [http://localhost:8080/](http://localhost:8080/) from the terminal, or to navigate to the same URL in your browser.