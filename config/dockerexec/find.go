package dockerexec

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

// ErrContainerNotUnique is an error that is returned when a container is not unique
var ErrContainerNotUnique = errors.New("No unique container found")

// FindUniqueContainer finds a unique running container with the given label key and value
//
// If there is no unique runing container, returns ErrContainerNotUnique.
// If something goes wrong, other errors may be returned.
func FindUniqueContainer(cli client.APIClient, key string, value string) (container_ types.Container, err error) {
	// Setup a filter for a running container with the given key/value label
	Filters := filters.NewArgs()
	Filters.Add("label", fmt.Sprintf("%s=%s", key, value))
	Filters.Add("status", "running")

	// do the list
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: Filters,
	})
	if err != nil {
		return types.Container{}, errors.Wrap(err, "Unable to list containers")
	}

	// if there isn't exactly one container
	// we bail out

	if len(containers) != 1 {
		err = ErrContainerNotUnique
		return
	}

	return containers[0], nil
}
