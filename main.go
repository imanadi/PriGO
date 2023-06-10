package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type task struct {
	Name        string
	Description string
	Priority    int
	Deadline    time.Time
	Duration    int
}

func main() {
	fmt.Println("Welcome")
	username, password := getCredentials()
	db, err := sql.Open("mysql", username+":"+password+"@tcp(localhost:3306)/task_picker")
	if err != nil {
		fmt.Println("Failed to connect to the database:", err)
		return
	}
	defer db.Close()

	for {
		fmt.Println("Enter 1 to enter a task, 2 to return upcoming tasks, 3 to return old tasks, anything else to exit")
		var input int
		fmt.Scanln(&input)
		switch input {
		case 1:
			_, err := createTask(db)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
		case 2:
			err := sortTasks(db)
			if err != nil {
				fmt.Println("Error sorting tasks:", err)
			}
			returnTasks(db, "tasks")
		case 3:
			err := sortTasks(db)
			if err != nil {
				fmt.Println("Error sorting tasks:", err)
			}
			returnTasks(db, "oldTasks")
		default:
			return
		}
	}
}

func createTask(db *sql.DB) (task, error) {
	err := sortTasks(db)
	if err != nil {
		fmt.Println("Error sorting tasks:", err)
	}
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter task name: ")
	scanner.Scan()
	name := scanner.Text()

	fmt.Print("Enter task description: ")
	scanner.Scan()
	description := scanner.Text()

	fmt.Print("Enter task priority (integer): ")
	scanner.Scan()
	priorityStr := scanner.Text()
	priority, err := strconv.Atoi(priorityStr)
	if err != nil {
		return task{}, fmt.Errorf("failed to convert priority to integer: %v", err)
	}

	fmt.Print("Enter task deadline (yyyy-mm-dd): ")
	scanner.Scan()
	deadlineStr := scanner.Text()
	deadline, err := time.Parse("2006-01-02", deadlineStr)
	if err != nil {
		return task{}, fmt.Errorf("failed to parse deadline: %v", err)
	}

	fmt.Print("Enter task duration (in days): ")
	scanner.Scan()
	durationStr := scanner.Text()
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		return task{}, fmt.Errorf("failed to convert duration to integer: %v", err)
	}

	newTask := task{
		Name:        name,
		Description: description,
		Priority:    priority,
		Deadline:    deadline,
		Duration:    duration,
	}

	// Insert the task into the MySQL table
	_, err = db.Exec("INSERT INTO tasks (name, description, priority, deadline, duration) VALUES (?, ?, ?, ?, ?)",
		newTask.Name, newTask.Description, newTask.Priority, newTask.Deadline, newTask.Duration)
	if err != nil {
		return task{}, fmt.Errorf("failed to insert task: %v", err)
	}

	return newTask, nil
}

func returnTasks(db *sql.DB, tableName string) {

	rows, err := db.Query("SELECT name, description, priority, deadline, duration FROM " + tableName)
	if err != nil {
		fmt.Println("Failed to retrieve tasks:", err)
		return
	}
	defer rows.Close()

	fmt.Println("Tasks from " + tableName)
	for rows.Next() {
		var name, description string
		var priority, duration int
		var deadlineStr string

		err := rows.Scan(&name, &description, &priority, &deadlineStr, &duration)
		if err != nil {
			fmt.Println("Error reading task:", err)
			return
		}

		deadline, err := time.Parse("2006-01-02", deadlineStr)
		if err != nil {
			fmt.Println("Error parsing deadline:", err)
			return
		}

		t := task{
			Name:        name,
			Description: description,
			Priority:    priority,
			Deadline:    deadline,
			Duration:    duration,
		}

		fmt.Printf("Name: %s\nDescription: %s\nPriority: %d\nDeadline: %s\nDuration: %d days\n\n",
			t.Name, t.Description, t.Priority, t.Deadline.Format("2006-01-02"), t.Duration)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Error:", err)
	}
}

func sortTasks(db *sql.DB) error {
	// Move tasks whose deadlines are crossed to the "oldTasks" table
	_, err := db.Exec("INSERT INTO oldTasks SELECT * FROM tasks WHERE deadline < NOW()")
	if err != nil {
		return fmt.Errorf("failed to move tasks to oldTasks table: %v", err)
	}

	// Delete the moved tasks from the "tasks" table
	_, err = db.Exec("DELETE FROM tasks WHERE deadline < NOW()")
	if err != nil {
		return fmt.Errorf("failed to delete tasks from tasks table: %v", err)
	}

	// Sort the remaining tasks by priority and duration
	_, err = db.Exec("ALTER TABLE tasks ORDER BY priority, deadline, duration")
	if err != nil {
		return fmt.Errorf("failed to sort tasks: %v", err)
	}
	_, err = db.Exec("ALTER TABLE oldTasks ORDER BY priority, deadline, duration")
	if err != nil {
		return fmt.Errorf("failed to sort tasks: %v", err)
	}

	return nil
}

func getCredentials() (username string, password string) {
	fmt.Println("Enter MySQL username:")
	fmt.Scanln(&username)
	fmt.Println("Enter MySQL password:")
	fmt.Scanln(&password)
	return username, password
}