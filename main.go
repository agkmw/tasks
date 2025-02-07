package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
	"time"
)

type Task struct {
    Id int
    Todo string
    CreatedAt time.Time
    Done bool
}

func (t Task) String() string {
    return fmt.Sprintf(`{Id: %v, Todo: %v, CreatedAt: %v, Done: %v,}`,
        t.Id, t.Todo, t.CreatedAt, t.Done)
}

func main() {
	commands := []string{"add", "list", "complete", "delete"}
	all := flag.Bool("all", true, "Display all tasks")

	if len(os.Args) < 2 {
		fmt.Println("not enought argument")
		os.Exit(1)
	}

	arg := os.Args[1]
	if !slices.Contains(commands, arg) {
		fmt.Println("unavailable command")
		os.Exit(1)
	}

	commandCounts := 0
	for _, arg := range os.Args {
		if slices.Contains(commands, arg) {
			commandCounts++
		}
	}

	if commandCounts > 1 {
		fmt.Println("pass one command at a time")
		os.Exit(1)
	}

	if arg != "list" && slices.Contains(os.Args, "--all") {
		fmt.Printf("\"%v\" does not accept any flags\n", arg)
		os.Exit(1)
	}

	os.Args = os.Args[1:]
	flag.Parse()

	fmt.Println(*all)


	switch arg {
	case "add":
        var tasks []Task

		if len(os.Args[1:]) > 1 {
			fmt.Println("too many tasks. add one task at a time")
			os.Exit(1)
		}

        byteTasks, err := os.ReadFile("./tasks.json")
        if err != nil {
            fmt.Println(err.Error())
        }

        f, err := os.OpenFile("./tasks.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
        if err != nil {
            fmt.Println(err.Error())
        }
        defer f.Close()


        err = json.Unmarshal(byteTasks, &tasks)
        if err != nil {
            fmt.Println(err.Error())
        }

        var latestTaskId int
        latestTaskId = 1
        if len(tasks) != 0 {
            latestTaskId = tasks[len(tasks) - 1].Id + 1
        }

        newTask := Task{latestTaskId, os.Args[1], time.Unix(time.Now().Unix(), 0).Local(), false}

        tasks = append(tasks, newTask)
        for _, t := range tasks {
            fmt.Println(t)
        }
        output, err := json.Marshal(tasks)
        if err != nil {
            fmt.Println(err.Error())
        }
        fmt.Println(string(output))
        f.Write(output)

	case "complete":
	case "delete":
	case "list":
	}

}
