package containermanager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	api "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/dockersamples/gopher-task-system/internal/types"
	"github.com/pkg/errors"
)

type ContainerManager interface {
	PullImage(ctx context.Context, image string) error
	CreateContainer(ctx context.Context, task types.Task) (string, error)
	StartContainer(ctx context.Context, id string) error
	WaitForContainer(ctx context.Context, id string) (bool, error)
	RemoveContainer(ctx context.Context, id string) error
}

type DockerClient interface {
	client.ImageAPIClient
	client.ContainerAPIClient
}

type ImagePullStatus struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

type containermanager struct {
	cli DockerClient
}

func NewContainerManager(cli DockerClient) ContainerManager {
	return &containermanager{
		cli: cli,
	}
}

// PullImage outputs to stdout the contents of the runner image.
func (m *containermanager) PullImage(ctx context.Context, image string) error {
	out, err := m.cli.ImagePull(ctx, image, api.ImagePullOptions{})
	if err != nil {
		return errors.Wrap(err, "DOCKER PULL")
	}

	defer func() {
		if err := out.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	fd := json.NewDecoder(out)
	var status *ImagePullStatus
	for {
		if err := fd.Decode(&status); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return errors.Wrap(err, "DOCKER PULL")
		}

		if status.Error != "" {
			return errors.Wrap(errors.New(status.Error), "DOCKER PULL")
		}

		// uncomment to log image pull status
		// fmt.Println(status)
	}

	return nil
}

// CreateContainer creates a new container and returns it ID.
func (m *containermanager) CreateContainer(ctx context.Context, task types.Task) (string, error) {
	config := &container.Config{
		Image: task.Runner,
		Cmd:   task.Command,
	}

	res, err := m.cli.ContainerCreate(ctx, config, &container.HostConfig{}, nil, nil, task.Name)
	if err != nil {
		return "", err
	}

	return res.ID, nil
}

// StartContainer starts the container created with given ID.
func (m *containermanager) StartContainer(ctx context.Context, id string) error {
	return m.cli.ContainerStart(ctx, id, api.ContainerStartOptions{})
}

// WaitForContainer waits for the running container to finish.
func (m *containermanager) WaitForContainer(ctx context.Context, id string) (bool, error) {
	// check if the container is in running state
	if _, err := m.cli.ContainerInspect(ctx, id); err != nil {
		return true, nil
	}

	// send API call to wait for the container completion
	wait, errC := m.cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	// check if container exit code is 0, and return accordingly
	select {
	case status := <-wait:
		if status.StatusCode == 0 {
			return true, nil
		}

		return false, nil
	case err := <-errC:
		return false, err
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

// RemoveContainer removes the given container../0+
func (m *containermanager) RemoveContainer(ctx context.Context, id string) error {
	return m.cli.ContainerRemove(ctx, id, api.ContainerRemoveOptions{})
}
