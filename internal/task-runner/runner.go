package taskrunner

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	cm "github.com/dockersamples/gopher-task-system/internal/container-manager"
	"github.com/dockersamples/gopher-task-system/internal/types"
)

type Runner interface {
	Run(ctx context.Context, doneCh chan<- bool)
}

type runner struct {
	def              types.TaskDefinition
	containerManager cm.ContainerManager
}

func NewRunner(def types.TaskDefinition) (Runner, error) {
	client, err := initDockerClient()
	if err != nil {
		return nil, err
	}

	return &runner{
		def:              def,
		containerManager: cm.NewContainerManager(client),
	}, nil
}

func initDockerClient() (cm.DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func (r *runner) Run(ctx context.Context, doneCh chan<- bool) {
	taskDoneCh := make(chan bool)
	for _, task := range r.def.Tasks {
		go r.run(ctx, task, taskDoneCh)
	}

	taskCompleted := 0
	for {
		if <-taskDoneCh {
			taskCompleted++
		}

		if taskCompleted == len(r.def.Tasks) {
			doneCh <- true
			return
		}
	}
}

func (r *runner) run(ctx context.Context, task types.Task, taskDoneCh chan<- bool) {
	defer func() {
		taskDoneCh <- true
	}()

	fmt.Println("preparing task - ", task.Name)
	if err := r.containerManager.PullImage(ctx, task.Runner); err != nil {
		fmt.Println(err)
		return
	}

	id, err := r.containerManager.CreateContainer(ctx, task)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("starting task - ", task.Name)
	err = r.containerManager.StartContainer(ctx, id)
	if err != nil {
		fmt.Println(err)
		return
	}

	statusSuccess, err := r.containerManager.WaitForContainer(ctx, id)
	if err != nil {
		fmt.Println(err)
		return
	}

	if statusSuccess {
		fmt.Println("completed task - ", task.Name)

		// cleanup by removing the task container
		if task.Cleanup {
			fmt.Println("cleanup task - ", task.Name)
			err = r.containerManager.RemoveContainer(ctx, id)
			if err != nil {
				fmt.Println(err)
			}
		}
	} else {
		fmt.Println("failed task - ", task.Name)
	}
}
