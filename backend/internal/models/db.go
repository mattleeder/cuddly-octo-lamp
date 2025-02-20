package models

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"reflect"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	rowsLog  *log.Logger
	perfLog  *log.Logger
}

type task func() (any, error)
type valueOnlyTask func() any
type errorOnlyTask func() error

type TaskResponse struct {
	val any
	err error
}

type Task struct {
	task    task
	channel chan TaskResponse
}

type TaskQueue struct {
	tasks chan Task
}

func wrapValueOnlyTask(task valueOnlyTask) task {
	return func() (any, error) {
		return task(), nil
	}
}

func wrapErrorOnlyTask(task errorOnlyTask) task {
	return func() (any, error) {
		return nil, task()
	}
}

func (taskQueue *TaskQueue) runWorker() {
	for {
		select {
		case task := <-taskQueue.tasks:
			channel := task.channel
			result, err := task.task()
			if channel != nil {
				channel <- TaskResponse{val: result, err: err}
			}
		}
	}
}

func (taskQueue *TaskQueue) EnQueue(task task) {
	taskQueue.tasks <- Task{task: task, channel: nil}
}

func (taskQueue *TaskQueue) EnQueueReturn(task task) (any, error) {
	channel := make(chan TaskResponse, 1)
	taskQueue.tasks <- Task{task: task, channel: channel}
	taskResponse := <-channel
	return taskResponse.val, taskResponse.err
}

func (taskQueue *TaskQueue) EnQueueValueOnlyTask(task valueOnlyTask) {
	taskQueue.tasks <- Task{task: wrapValueOnlyTask(task), channel: nil}
}

func (taskQueue *TaskQueue) EnQueueReturnValueOnlyTask(task valueOnlyTask) any {
	channel := make(chan TaskResponse, 1)
	taskQueue.tasks <- Task{task: wrapValueOnlyTask(task), channel: channel}
	taskResponse := <-channel
	return taskResponse.val
}

func (taskQueue *TaskQueue) EnQueueErrorOnlyTask(task errorOnlyTask) {
	taskQueue.tasks <- Task{task: wrapErrorOnlyTask(task), channel: nil}
}

func (taskQueue *TaskQueue) EnQueueReturnErrorOnlyTask(task errorOnlyTask) error {
	channel := make(chan TaskResponse, 1)
	taskQueue.tasks <- Task{task: wrapErrorOnlyTask(task), channel: channel}
	taskResponse := <-channel
	return taskResponse.err
}

var app *application

var DBTaskQueue *TaskQueue

const numdbTaskQueueWorkers = 5

func CreateClosure(fn any, args ...any) func() {
	return func() {
		// Use reflection to call the function with the provided arguments
		reflectFn := reflect.ValueOf(fn)
		if reflectFn.Kind() != reflect.Func {
			app.errorLog.Panicln("fn must be a function")
		}

		// Convert args to []reflect.Value
		reflectArgs := make([]reflect.Value, len(args))
		for i, arg := range args {
			reflectArgs[i] = reflect.ValueOf(arg)
		}

		// Call the function
		reflectFn.Call(reflectArgs)
	}
}

func init() {

	infoLog := log.New(os.Stdout, "DB INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "DB ERROR\t", log.Ldate|log.Ltime|log.Llongfile)
	rowsLog := log.New(os.Stdout, "DB ROW\t", 0)
	perfLog := log.New(os.Stdout, "DB PERF\t", log.Lshortfile)

	app = &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		rowsLog:  rowsLog,
		perfLog:  perfLog,
	}

	app.infoLog.Println("RAN MODELS INIT")

	for i := 0; i < numdbTaskQueueWorkers; i++ {
		app.infoLog.Printf("Starting DB Task Queue Worker Number %v\n", i)
		go DBTaskQueue.runWorker()
	}

}

func InitDatabase(driverName string, dataSourceName string) {
	os.Remove("./chess_site.db")

	db, err := sql.Open(driverName, dataSourceName)
	defer db.Close()
	if err != nil {
		app.errorLog.Fatal(err)
	}

	schemaPath := filepath.Join("internal", "models", "schema.sql")
	c, ioErr := os.ReadFile(schemaPath)
	if ioErr != nil {
		app.errorLog.Fatalf("%s", ioErr.Error())
	}
	sqlStmt := string(c)

	_, err = db.Exec(sqlStmt)
	if err != nil {
		app.errorLog.Fatalf("%q: %s\n", err, sqlStmt)
	}
}
