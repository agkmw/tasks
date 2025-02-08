package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"time"
)

type taskError struct {
	message string
}

func (e *taskError) Error() string {
	return e.message
}

type task struct {
	Id        int
	Task      string
	CreatedAt time.Time
	Done      bool
}

func (t task) String() string {
	return fmt.Sprintf(`
    {
    Id: %v,
    Task: %v,
    CreatedAt: %v,
    Done: %v,
    }`, t.Id, t.Task, t.CreatedAt, t.Done)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	commands := []string{"add", "list", "complete", "delete"}
	all := flag.Bool("all", true, "Display all tasks")

	var err error
	os.Args, err = validateArgs(os.Args, commands)
	if err != nil {
		return err
	}
	flag.Parse()

	tasks, err := getTasks()
	command := os.Args[0]
	switch command {
	case "add":
    if err := addTask(tasks); err != nil {
      return err
    }
	case "complete":
	case "delete":
	case "list":
	}
	fmt.Println(*all)
	return nil
}

func addTask(tasks []task) error {
	if len(os.Args[1:]) > 1 {
		return &taskError{"tasks: too many task. add one task at a time"}
	}

  if len(os.Args[1:]) <= 1 {
    return &taskError{`tasks: please specify the task for "add" command`}
  }

	var latestTaskId int
	latestTaskId = 1
	if len(tasks) != 0 {
		latestTaskId = tasks[len(tasks)-1].Id + 1
	}

	newTask := task{
		latestTaskId,
		os.Args[1],
		time.Unix(time.Now().Unix(), 0).Local(),
		false,
	}

	tasks = append(tasks, newTask)

	f, err := os.OpenFile("./tasks.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	err = encoder.Encode(tasks)
	if err != nil {
		return err
	}
	return nil
}

func validateArgs(args, commands []string) ([]string, error) {
	if len(args) < 2 {
		return nil, &taskError{fmt.Sprintf("tasks: not enough arguments\n")}
	}

	arg := args[1]
	if !slices.Contains(commands, arg) {
		return nil, &taskError{fmt.Sprintf(`tasks: command "%v" unavailable\n`, arg)}
	}

	commandCounts := 0
	for _, arg := range args {
		if slices.Contains(commands, arg) {
			commandCounts++
		}
	}

	if commandCounts > 1 {
		return nil, &taskError{fmt.Sprintf(`tasks: expected 1 command, got %v\n`, commandCounts)}
	}

	if arg != "list" && slices.Contains(args, "--all") {
		return nil, &taskError{fmt.Sprintf(`tasks: "%v" command does not accepts any flags\n`, arg)}
	}

	return args[1:], nil
}

func getTasks() ([]task, error) {
	var tasks []task

	contents, err := os.ReadFile("./tasks.json")
	if err != nil {
		return nil, &taskError{fmt.Sprintf(`tasks: %v\n`, err.Error())}
	}

	err = json.Unmarshal(contents, &tasks)
	if err != nil {
		return nil, &taskError{fmt.Sprintf(`tasks: %v\n`, err.Error())}
	}

	return tasks, nil
}
