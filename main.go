package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"
)

type task struct {
	Name        string
	Description string
	Priority    int
	Deadline    time.Time
	Duration    int
}

func main() {
	fmt.Println("Function started")
	for true {
		fmt.Println("Enter 1 to enter task, 2 to check tasks, anything else to exit")
		var input int
		fmt.Scanln(&input)
		switch input {
		case 1:
			newTask, err := createTask()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("Task is %v\n", newTask)
		case 2:
			returnTasks()
		default:
			return

		}
	}
}

func createTask() (task, error) {
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

	fmt.Print("Enter task duration (in hours): ")
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

	return newTask, nil
}

func returnTasks() {

}
