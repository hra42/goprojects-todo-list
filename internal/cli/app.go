package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	duckdb "github.com/hra42/goprojects-todo-list/internal/sqlite"
	"github.com/mergestat/timediff"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Run         func([]string) error
}

var commands = []Command{
	{"add", "Add a new task", handleAdd},
	{"list", "List tasks", handleList},
	{"complete", "Mark a task as complete", handleComplete},
	{"delete", "Delete a task", handleDelete},
}

// Start initializes and runs the CLI application
func Start() error {
	err := duckdb.InitDB("tasks.db")
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	defer duckdb.CloseDB()

	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	commandName := os.Args[1]
	for _, cmd := range commands {
		if cmd.Name == commandName {
			return cmd.Run(os.Args[2:])
		}
	}

	fmt.Printf("Unknown command: %s\n", commandName)
	printUsage()
	return nil
}

func printUsage() {
	fmt.Println("Usage: tasks <command> [arguments]")
	fmt.Println("Available commands:")
	for _, cmd := range commands {
		fmt.Printf("  %s\t%s\n", cmd.Name, cmd.Description)
	}
}

func handleAdd(args []string) error {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	description := addCmd.String("desc", "", "Task description")
	addCmd.Parse(args)

	if *description == "" {
		return fmt.Errorf("task description is required")
	}

	err := duckdb.AddTask(*description)
	if err != nil {
		return fmt.Errorf("failed to add task: %v", err)
	}
	fmt.Println("Task added successfully")
	return nil
}

func handleList(args []string) error {
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	showAll := listCmd.Bool("all", false, "Show all tasks including completed")
	listCmd.Parse(args)

	tasks, err := duckdb.ListTasks(*showAll)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTask\tCreated")

	for _, task := range tasks {
		status := ""
		if *showAll {
			if task.IsComplete {
				status = "[x] "
			} else {
				status = "[ ] "
			}
		}
		createdAgo := timediff.TimeDiff(task.CreatedAt)

		// Truncate description if it's too long
		description := task.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		fmt.Fprintf(w, "%s%d\t%s\t%s\n", status, task.ID, description, createdAgo)
	}

	w.Flush()
	return nil
}

func handleComplete(args []string) error {
	completeCmd := flag.NewFlagSet("complete", flag.ExitOnError)
	completeCmd.Parse(args)

	if completeCmd.NArg() != 1 {
		return fmt.Errorf("usage: tasks complete <task_id>")
	}

	taskID, err := strconv.Atoi(completeCmd.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid task ID: %v", err)
	}

	err = duckdb.CompleteTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to complete task: %v", err)
	}
	fmt.Println("Task marked as complete")
	return nil
}

func handleDelete(args []string) error {
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteCmd.Parse(args)

	if deleteCmd.NArg() != 1 {
		return fmt.Errorf("usage: tasks delete <task_id>")
	}

	taskID, err := strconv.Atoi(deleteCmd.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid task ID: %v", err)
	}

	err = duckdb.DeleteTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}
	fmt.Println("Task deleted successfully")
	return nil
}
