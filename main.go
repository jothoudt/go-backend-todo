package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

//-----This is to prevent an error for a NULL SQL VALUE------//
type NullString string

func (s *NullString) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	strVal, ok := value.(string)
	if !ok {
		return errors.New("Column is not a string")
	}
	*s = NullString(strVal)
	return nil
}
func (s NullString) Value() (driver.Value, error) {
	if len(s) == 0 { // if nil or empty string
		return nil, nil
	}
	return string(s), nil
}

//------------------------------------------------------------//
//struct for each task
type Todo struct {
	Id             int        `json:"id"`
	Task           string     `json:"task"`
	Completed      bool       `json:"completed"`
	Date_added     time.Time  `json:"date_added"`
	Date_completed NullString `json:"date_completed"`
} //end Todo

//for sending data to the front-end
type JsonResponse struct {
	Type    string `json:"type"`
	Data    []Todo `json:"data"`
	Message string `json:"message"`
} //end JsonResponse

func main() {
	router := mux.NewRouter()
	//to prevent CORS errors
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	) //end cors
	router.Use(cors)
	// Get all tasks
	router.HandleFunc("/api/tasks/", GetTasks).Methods("GET")
	// Create a task
	router.HandleFunc("/api/tasks/", CreateTask).Methods("POST")
	//Delete a task
	router.HandleFunc("/api/tasks/{id}", DeleteTask).Methods("DELETE")
	// to update a task to complete
	router.HandleFunc("/api/tasks/{id}", CompleteTask).Methods("PUT")
	//test server
	router.HandleFunc("/", HelloServer)
	//spin up server
	log.Fatal(http.ListenAndServe(":5000", router))
} //end main
//this function checks for errors
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
} //end checkErr
func setupDB() *sql.DB {
	//------Load .env variables-----------------------//
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_NAME := os.Getenv("DB_NAME")
	//-----------------------------------------------//
	//information used to setup database connection
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	//open database
	db, err := sql.Open("postgres", dbinfo)
	//check error
	checkErr(err)
	//return database
	return db
} //end setupDB

// GET all tasks from the database
func GetTasks(w http.ResponseWriter, r *http.Request) {
	//setup connection with database
	db := setupDB()
	// Get all tasks
	rows, err := db.Query("SELECT * FROM todo")
	// check errors
	checkErr(err)
	var tasks []Todo
	// For each task
	for rows.Next() {
		var Id int
		var Task string
		var Completed bool
		var Date_added time.Time
		var Date_completed NullString

		err = rows.Scan(&Id, &Task, &Completed, &Date_added, &Date_completed)

		// check errors
		checkErr(err)
		//define tasks
		tasks = append(tasks, Todo{Id: Id, Task: Task, Completed: Completed, Date_added: Date_added, Date_completed: Date_completed})
	} //end for
	//this prevents an error on the front end by not allowing it to send the value of null
	if tasks == nil {
		fmt.Println("no tasks")
	} else {
		json.NewEncoder(w).Encode(tasks)
	} //end else
} //end GetTask

//POST---adds new task to the database
func CreateTask(w http.ResponseWriter, r *http.Request) {
	//get the data from r.Body
	decoder := json.NewDecoder(r.Body)
	var data Todo
	errr := decoder.Decode(&data)
	if errr != nil {
		panic(errr)
	}
	//define task to add to the database
	task := data.Task
	//to test that we are recieving the task
	fmt.Println("tasks", task)
	//define response
	var response = JsonResponse{}
	// set up connection with database
	db := setupDB()
	var lastInsertID int
	//database query
	err := db.QueryRow("INSERT INTO todo(task) VALUES($1) returning id;", task).Scan(&lastInsertID)
	// check errors
	checkErr(err)
	// send response
	json.NewEncoder(w).Encode(response)
} //end createTask

//DELETE---to delete a task
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	//get id to use for SQL query
	id := params["id"]
	//define response
	var response = JsonResponse{}
	//set up database connection
	db := setupDB()
	//database query
	_, err := db.Exec("DELETE FROM todo WHERE id= $1;", id)
	// check errors
	checkErr(err)
	response = JsonResponse{Type: "success", Message: "This task has been deleted successfully!"}
	//send response
	json.NewEncoder(w).Encode(response)
} //end DeleteTask

//PUT---to update a task to complete
func CompleteTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	//get id to use for SQL query
	id := params["id"]
	//define completed time to add to the database
	completedTime := time.Now()
	//define response
	var response = JsonResponse{}
	//set up database connection
	db := setupDB()
	//database query
	_, err := db.Exec("UPDATE todo SET completed=$1, date_completed=$2 WHERE id= $3;", true, completedTime, id)
	// check errors
	checkErr(err)
	response = JsonResponse{Type: "success", Message: "This task has been updated to complete successfully!"}
	//send response
	json.NewEncoder(w).Encode(response)
} //end CompleteTask

//to test server
func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
} //end HelloServer
