package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh/terminal"
)

type task struct {
	Name        string
	Description string
	Priority    int
	Deadline    time.Time
	Duration    int
}

type taskManager struct {
	db *sql.DB
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

	manager := taskManager{db: db}

	for {
		fmt.Println("Enter 1 to enter a task, 2 to return upcoming tasks, 3 to return old tasks, anything else to exit")
		var input int
		fmt.Scanln(&input)
		switch input {
		case 1:
			_, err := manager.createTask()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
		case 2:
			err := manager.sortTasks()
			if err != nil {
				fmt.Println("Error sorting tasks:", err)
			}
			manager.returnTasks("tasks")
		case 3:
			err := manager.sortTasks()
			if err != nil {
				fmt.Println("Error sorting tasks:", err)
			}
			manager.returnTasks("oldTasks")
		default:
			return
		}
	}
}

func (tm *taskManager) createTask() (task, error) {
	err := tm.sortTasks()
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
	_, err = tm.db.Exec("INSERT INTO tasks (name, description, priority, deadline, duration) VALUES (?, ?, ?, ?, ?)",
		newTask.Name, newTask.Description, newTask.Priority, newTask.Deadline, newTask.Duration)
	if err != nil {
		return task{}, fmt.Errorf("failed to insert task: %v", err)
	}

	return newTask, nil
}

func (tm *taskManager) returnTasks(tableName string) {
	rows, err := tm.db.Query("SELECT name, description, priority, deadline, duration FROM " + tableName)
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

func (tm *taskManager) sortTasks() error {
	// Move tasks whose deadlines are crossed to the "oldTasks" table
	_, err := tm.db.Exec("INSERT INTO oldTasks SELECT * FROM tasks WHERE deadline < NOW()")
	if err != nil {
		return fmt.Errorf("failed to move tasks to oldTasks table: %v", err)
	}

	// Delete the moved tasks from the "tasks" table
	_, err = tm.db.Exec("DELETE FROM tasks WHERE deadline < NOW()")
	if err != nil {
		return fmt.Errorf("failed to delete tasks from tasks table: %v", err)
	}

	// Sort the remaining tasks by priority, deadline, and duration
	_, err = tm.db.Exec("ALTER TABLE tasks ORDER BY priority, deadline, duration")
	if err != nil {
		return fmt.Errorf("failed to sort tasks: %v", err)
	}
	_, err = tm.db.Exec("ALTER TABLE oldTasks ORDER BY priority, deadline, duration")
	if err != nil {
		return fmt.Errorf("failed to sort tasks: %v", err)
	}

	return nil
}

func getCredentials() (username string, password string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter MySQL username: ")
	username, _ = reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Enter MySQL password: ")
	passwordBytes, _ := terminal.ReadPassword(int(syscall.Stdin))
	password = string(passwordBytes)
	fmt.Println()

	return username, password
}

