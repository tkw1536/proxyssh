package integrationtest

import (
	"context"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

// ComposeTestCode is a function that runs tests inside a docker-compose file context.
// It returns an error if the test cases failed, and nil otherwise.
// See also RunComposeTest.
//
// This function receives three parameters from the testrunner.
//
// cli is a valid Docker Client that can be used for arbitary interactions with docker.
//
// findService is a function that is provided a name of a service, and returns the container within the current docker-compose context.
//
// stopService and is a function that can be used to stop a particular service in the docker-compose service.
type ComposeTestCode func(cli client.APIClient, findService func(string) types.Container, stopService func(string)) error

// RunComposeTest runs a docker-compose based test.
// It first starts the docker-compose configuration contained in the string.
// Thenm it writes out addtional data files in the provided map and waits for it to become available.
// Finally, it runs the testcode function.
//
// Calling this function is roughly equivalent to the following bash pseudo-code:
//
//	mktmp -d
//	echo $config > tmpdir/docker-compose.yml
//	docker-compose pull
//	docker-compose up -d tmpdir/docker-compose.yml
//	// ... call f() ...
//	docker-compose down -v
//	rm -rf tmpdir
//
// For source code compatibility purposes, all occurences of '\t' in the compose file will be replace with four spaces.
//
// This command depends on the 'docker-compose' executable to be available in path.
//
// This function has two kinds of error conditions, those that occur during setting up the project, and those that occur in the testcode.
// If an error occurs during setup or teardown, panic() is called.
// If an error occurs during the testcase, and testcase does not call panic(), the error is returned by this function.
//
// This function is itself untested.
func RunComposeTest(config string, files map[string]string, testcode ComposeTestCode) error {
	// TODO: Make a simple test for this function

	// create a new docker client or panic
	cli, err := client.NewEnvClient()
	if err != nil {
		err = errors.Wrap(err, "Unable to create docker client")
		panic(err)
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	// check that docker-compose exists
	dockerComposePath, err := exec.LookPath("docker-compose")
	if err != nil {
		err = errors.Wrap(err, "Unable to find 'docker-compose' in $PATH")
		panic(err)
	}

	// create a temporary directory to work in
	// also setup deletion of the temporary directory after this test ends
	tmpDir, err := os.MkdirTemp("", "docker-compose-test")
	if err != nil {
		err = errors.Wrap(err, "Unable to create a temporary directory")
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	// create a docker-compose yml and place in the content of the configuration.
	// This file will be automatically deleted with everything else in the directory
	dockerComposeYML := path.Join(tmpDir, "docker-compose.yml")
	if os.WriteFile(dockerComposeYML, []byte(strings.ReplaceAll(config, "	", "    ")), os.ModePerm); err != nil {
		err = errors.Wrap(err, "Unable to write docker-compose.yml")
		panic(err)
	}

	// write out all the extra files
	for filename, content := range files {
		if err := os.WriteFile(path.Join(tmpDir, filename), []byte(content), os.ModePerm); err != nil {
			err = errors.Wrap(err, "Unable to write file")
			panic(err)
		}
	}

	// generate a project name
	projectName := "test-" + strconv.Itoa(rand.Int())

	// function to run a docker-compose command
	compose := func(args ...string) {
		cmd := exec.Command(dockerComposePath,
			append([]string{"-f", dockerComposeYML, "-p", projectName}, args...)...,
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = tmpDir

		if err := cmd.Run(); err != nil {
			err = errors.Wrap(err, "docker-compose failed to run")
			panic(err)
		}
	}

	// Setup running docker-compose up -d and the opposite docker-compose down -v
	compose("up", "-d")
	defer compose("down", "-v")

	// finally run the testcode
	return testcode(cli, func(service string) types.Container {

		filters := filters.NewArgs()
		filters.Add("label", "com.docker.compose.service="+service)
		filters.Add("label", "com.docker.compose.project="+projectName)

		lst, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
			Filters: filters,
		})
		if len(lst) == 0 {
			err = errors.New("No service containers exist")
		}

		if err != nil {
			err = errors.Wrap(err, "Unable to find service")
			panic(err)
		}

		return lst[0]
	}, func(service string) {
		compose("stop", service)
	})
}
