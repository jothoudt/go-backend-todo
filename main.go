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

// const (
// 	DB_USER     = os.Setenv("DB_USER")
// 	DB_PASSWORD = os.Setenv("DB_PASSWORD")
// 	DB_NAME     = os.Setenv("DB_PASSWORD")
// )

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

type Todo struct {
	Id             int        `json:"id"`
	Task           string     `json:"task"`
	Completed      bool       `json:"completed"`
	Date_added     time.Time  `json:"date_added"`
	Date_completed NullString `json:"date_completed"`
}

type JsonResponse struct {
	Data []Todo `json:"data"`
}

func main() {
	router := mux.NewRouter()
	// router.StrictSlash(true)

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	router.Use(cors)
	// Get all tasks
	router.HandleFunc("/api/tasks/", GetTasks).Methods("GET")

	// Create a task
	router.HandleFunc("/api/tasks/", CreateTask).Methods("POST")
	router.HandleFunc("/", HelloServer)
	log.Fatal(http.ListenAndServe(":5000", router))

}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
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
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	fmt.Println(dbinfo)
	db, err := sql.Open("postgres", dbinfo)

	checkErr(err)

	return db
}

// GET all tasks from the database
func GetTasks(w http.ResponseWriter, r *http.Request) {
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

		tasks = append(tasks, Todo{Id: Id, Task: Task, Completed: Completed, Date_added: Date_added, Date_completed: Date_completed})
	}
	if tasks == nil {
		fmt.Println("no tasks")
	} else {
		json.NewEncoder(w).Encode(tasks)
	}
}

//POST---adds new task to the database
func CreateTask(w http.ResponseWriter, r *http.Request) {
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
	var response = JsonResponse{}

	db := setupDB()
	var lastInsertID int
	err := db.QueryRow("INSERT INTO todo(task) VALUES($1) returning id;", task).Scan(&lastInsertID)

	// check errors
	checkErr(err)
	json.NewEncoder(w).Encode(response)
}

//to test server
func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}
