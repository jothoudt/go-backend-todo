# go-backend-todo

This is one of the REST-api applications for my to-do list. It allows user to GET, POST, DELETE, and UPDATE tasks. This is my first time working with Go and it is just to get some practice. The front-end of this application can be found at [front-end to-do](https://github.com/jothoudt/frontend-to-do).

# Technologies Used

- Go
- PostgreSQL

# Requirements

- Add .env file to the project. Add the following and update the italic text to match the postgreSQL information for your environment:
    DB_USER :=<em>DB_USER</em>
	DB_PASSWORD :=<em>DB_PASSWORD</em>
	DB_NAME := <em>DB_NAME</em>
- Must have postgreSQL installed
- Create a database "go_todo"
- Create the database table found in the database.sql file
- Get the front-end of the to-do list [here](https://github.com/jothoudt/frontend-to-do).

- npm start on the front-end

- go run main.go