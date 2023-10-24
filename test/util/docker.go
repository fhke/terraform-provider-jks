package util

import (
	"context"
	"fmt"
	"io"
	"os/user"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

// runContainer runs a one-shot container & removes it after completion.
func runContainer(ctx context.Context, t *testing.T, cli *client.Client, image string, workDir string, entrypoint ...string) {
	// working directory for container
	const containerCwd = "/var/tmp/workspace"

	// Get current UID
	currUsr, err := user.Current()
	require.NoError(t, err, "It should return current user")

	// Generate container name from UUID
	containerName, err := uuid.NewRandom()
	require.NoError(t, err, "It should generate UUID")

	// pull image
	r, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	require.NoError(t, err, "It should pull image ", image)
	_, err = io.ReadAll(r)
	require.NoError(t, err)
	r.Close()

	// create container
	crResp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:      image,
			Entrypoint: strslice.StrSlice(entrypoint),
			WorkingDir: containerCwd,
			User:       currUsr.Uid,
		},
		&container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:%s", workDir, containerCwd),
			},
		},
		&network.NetworkingConfig{},
		&v1.Platform{},
		containerName.String(),
	)
	require.NoError(t, err, "It should create container")
	defer removeContainer(t, cli, crResp.ID)

	// start container
	err = cli.ContainerStart(
		ctx,
		crResp.ID,
		types.ContainerStartOptions{},
	)
	require.NoError(t, err, "It should start container")

	// wait for container to complete
	st := ""
	for {
		ctr, err := cli.ContainerInspect(ctx, crResp.ID)
		if err == context.DeadlineExceeded {
			t.Fatalf("Timed out waiting for container with ID %s to complete, state: %s", crResp.ID, st)
		}
		require.NoError(t, err, "It should inspect container")

		switch st = ctr.State.Status; st {
		case "created":
		case "running":
		case "exited":
			if ctr.State.ExitCode != 0 {
				t.Fatalf("Container exit code is %d", ctr.State.ExitCode)
			}
			return
		default:
			t.Fatalf("Unexpected state for container %s: %s", crResp.ID, st)
		}
		time.Sleep(time.Second / 2)
	}
}

// removeContainer forcibly removes container with ID containerID.
func removeContainer(t *testing.T, cli *client.Client, containerID string) {
	err := cli.ContainerRemove(
		context.Background(),
		containerID,
		types.ContainerRemoveOptions{
			Force: true,
		},
	)
	require.NoErrorf(t, err, "It should remove container with ID %s", containerID)
}
