package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"text/tabwriter"
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

type Color string

const (
	ColorRed   Color = "\u001b[31m"
	ColorGreen       = "\u001b[32m"
	ColorReset       = "\u001b[0m"
)

func logError(err error) {
	log.Fatal(string(ColorRed), err, string(ColorReset))
}

func logSuccess(message string) {
	fmt.Println(string(ColorGreen), message, string(ColorReset))
}


func run() error {
	commands := []string{"add", "list", "complete", "delete"}
	all := flag.Bool("all", false, "Display all tasks")

	var err error
	os.Args, err = validateArgs(os.Args, commands)
	if err != nil {
		return err
	}

	flag.Parse()

  fmt.Println(time.Unix(time.Now().Unix(), 0).UTC().Location())
	command := os.Args[0]
	switch command {
	case "add":
		if err := addHandler(); err != nil {
			return err
		}
	case "complete":
		if err := completeHandler(); err != nil {
			return err
		}
	case "delete":
		if err := deleteHandler(*all); err != nil {
			return err
		}
	case "list":
    if err := listHandler(*all); err != nil {
      return err
    }
	}

	return nil
}

func validateArgs(args, commands []string) ([]string, error) {
	if len(args) < 2 {
		return nil, &taskError{"tasks: not enough arguments\n"}
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

	if arg != "list" && arg != "delete" && slices.Contains(args, "--all") {
		return nil, &taskError{fmt.Sprintf(`tasks: "%v" command does not accepts any flags\n`, arg)}
	}

	return args[1:], nil
}

func parseCmdArg(cmd string, args []string) (string, error) {
	if len(args[1:]) > 1 {
		return "", &taskError{fmt.Sprintf("tasks: too many arguments for \"%v\" command", cmd)}
	}

	if len(args[1:]) < 1 {
		return "", &taskError{fmt.Sprintf(`tasks: argument missing for "%v" command`, cmd)}
	}

	return args[1], nil
}

func addHandler() error {
	arg, err := parseCmdArg("add", os.Args)
	if err != nil {
		return err
	}

	tasks, err := getTasks()
	if err != nil {
		return err
	}

	var latestTaskId int
	latestTaskId = 1
	if len(tasks) != 0 {
		latestTaskId = tasks[len(tasks)-1].Id + 1
	}

	newTask := task{
		latestTaskId,
		arg,
		time.Unix(time.Now().Unix(), 0).Local(),
		false,
	}

	tasks = append(tasks, newTask)
	err = storeTasks(tasks)
	if err != nil {
		return err
	}

	logSuccess(fmt.Sprintf("added \"%v\"\n", newTask.Task))
	return nil
}

func completeHandler() error {
	id, err := parseCmdArg("complete", os.Args)
	if err != nil {
		return err
	}

	intId, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	tasks, err := getTasks()
	if err != nil {
		return err
	}

	index := slices.IndexFunc(tasks, func(t task) bool {
		return t.Id == intId
	})
  if index == -1 {
    return &taskError{fmt.Sprintf("tasks: not found task with %v", intId)}
  }

  taskToComplete := &tasks[index]
  taskToComplete.Done = true

	err = storeTasks(tasks)
	if err != nil {
		return err
	}

	logSuccess(fmt.Sprintf("completed \"%v\"\n", taskToComplete.Task))
	return nil
}

func deleteHandler(all bool) error {
	if all {
		if err := deleteAllTasks(); err != nil {
			return err
		}
		return nil
	}

	if err := deleteTask(); err != nil {
		return err
	}
	return nil
}

func deleteAllTasks() error {
	err := storeTasks([]task{})
	if err != nil {
		return err
	}
	return nil
}

func deleteTask() error {
	id, err := parseCmdArg("complete", os.Args)
	if err != nil {
		return err
	}

	intId, err := strconv.Atoi(id)
	if err != nil {
		return &taskError{"tasks: unvalid task id for \"delete\" command"}
	}

	tasks, err := getTasks()
	if err != nil {
		return err
	}

	index := slices.IndexFunc(tasks, func(t task) bool {
		return t.Id == intId
	})
  if index == -1 {
    return &taskError{fmt.Sprintf("tasks: not found task with %v", intId)}
  }
	taskToDelete := tasks[index]

	tasks = slices.DeleteFunc(tasks, func(t task) bool {
		return t.Id == intId
	})

	err = storeTasks(tasks)
	if err != nil {
		return err
	}

	logSuccess(fmt.Sprintf("deleted \"%v\"\n", taskToDelete.Task))
	return nil
}

func listHandler(all bool) error {
  tasks, err := getTasks()
  if err != nil {
    return err
  }

  tasksToDisplay := tasks
  if !all {
    tasksToDisplay = []task {}
    for _, t := range tasks {
      if !t.Done {
        tasksToDisplay = append(tasksToDisplay, t)
      }
    }
  }

  w := tabwriter.NewWriter(os.Stdout, 0, 0, 7, ' ', 0)
  _, err = fmt.Fprintln(w, "Id\tTask\tCreated\tDone")
  if err != nil {
    return err
  }

  for _, t := range tasksToDisplay {
    taskOutput := fmt.Sprintf("%v\t%v\t%v\t%v", t.Id, t.Task, t.CreatedAt, t.Done)
    _, err := fmt.Fprintln(w, taskOutput)
    if err != nil {
      return err
    }
  }

  err = w.Flush()
  if err != nil {
    return err
  }
  return nil
}

func storeTasks(tasks []task) error {
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
	return f.Close()
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

func main() {
	if err := run(); err != nil {
		logError(err)
	}
}
