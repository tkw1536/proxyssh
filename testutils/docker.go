package testutils

import (
	"context"
	"io/ioutil"
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

// RunComposeTest starts the docker-compose configuration contained in the string, writes out addtional data files in the provided map, and then waits for it to become available. Finally, it runs f.
// Depends on the 'docker-compose' executable to be available in path.
//
// Calling this function is roughly equivalent to:
//
// mktmp -d
// echo $config > tmpdir/docker-compose.yml
// docker-compose pull
// docker-compose up -d tmpdir/docker-compose.yml
// ... run tests ...
// docker-compose down -v
// rm -rf tmpdir
//
// For source code compatibility purposes, all occurences of '\t' in the compose file will be replace with four spaces.
//
// The testCode function gets three additional parameters.
// The first parameter is a valid docker cli client that can be used during testing.
// The second is called findService and it returns the container belonging to a provided service in the compose-file.
// The third is called 'stopService' and is a function that can be used to stop a particular service from the compose-file.
//
// This function has two kinds of error conditions, those that occur during setting up the project, and those that occur in the testcode.
// If an error occurs during setup or teardown, panic() is called.
// If an error occurs during the testcase, and testcase does not call panic(), the error is returned by this function
func RunComposeTest(config string, files map[string]string, testcode func(cli *client.Client, findService func(string) types.Container, stopService func(string)) error) error {
	// create a new docker client or panic
	cli, err := client.NewEnvClient()
	if err != nil {
		err = errors.Wrap(err, "Unable to create docker client")
		panic(err)
	}
	defer cli.Close()

	// check that docker-compose exists
	dockerComposePath, err := exec.LookPath("docker-compose")
	if err != nil {
		err = errors.Wrap(err, "Unable to find 'docker-compose' in $PATH")
		panic(err)
	}

	// create a temporary directory to work in
	// also setup deletion of the temporary directory after this test ends
	tmpDir, err := ioutil.TempDir("", "docker-compose-test")
	if err != nil {
		err = errors.Wrap(err, "Unable to create a temporary directory")
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	// create a docker-compose yml and place in the content of the configuration.
	// This file will be automatically deleted with everything else in the directory
	dockerComposeYML := path.Join(tmpDir, "docker-compose.yml")
	if ioutil.WriteFile(dockerComposeYML, []byte(strings.ReplaceAll(config, "	", "    ")), os.ModePerm); err != nil {
		err = errors.Wrap(err, "Unable to write docker-compose.yml")
		panic(err)
	}

	// write out all the extra files
	for filename, content := range files {
		if err := ioutil.WriteFile(path.Join(tmpDir, filename), []byte(content), os.ModePerm); err != nil {
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
