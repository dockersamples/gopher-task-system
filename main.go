package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	taskrunner "github.com/dockersamples/gopher-task-system/internal/task-runner"
	"github.com/dockersamples/gopher-task-system/internal/types"

	yaml "gopkg.in/yaml.v2"
)

const (
	argRun      = "run"
	helpMessage = `You must provide valid arguments to gopher.
Example:
	./gopher run tasks.yaml`

	errReadTaskDef = "failed to read tasks due to error: %v\n"
	errNewRunner   = "failed to create new runner: %v\n"
	errTaskRun     = "failed to run the task: %v\n"
)

func main() {
	args := os.Args[1:]

	if len(args) < 2 || args[0] != argRun {
		fmt.Println(helpMessage)
		return
	}

	// read the task definition file
	def, err := readTaskDefinition(args[1])
	if err != nil {
		fmt.Printf(errReadTaskDef, err)
	}

	// create a task runner for the task definition
	ctx := context.Background()
	runner, err := taskrunner.NewRunner(def)
	if err != nil {
		fmt.Printf(errNewRunner, err)
	}

	doneCh := make(chan bool)
	go runner.Run(ctx, doneCh)

	<-doneCh
}

func readTaskDefinition(fileName string) (types.TaskDefinition, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return types.TaskDefinition{}, err
	}

	var def types.TaskDefinition
	err = yaml.Unmarshal(data, &def)
	if err != nil {
		return def, err
	}

	return def, nil
}
