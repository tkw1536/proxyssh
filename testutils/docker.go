package testutils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

// RunComposeTest starts the docker-compose configuration contained in the string, waits for it to become available, and then runs f.
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
// The testCode function gets two additional parameter.
// The first parameter is a valid docker cli client that can be used during testing.
// The second is called 'stopService' and is a function that can be used to stop a particular service from the compose-file.
//
// This function has two kinds of error conditions, those that occur during setting up the project, and those that occur in the testcode.
// If an error occurs during setup or teardown, panic() is called.
// If an error occurs during the testcase, and testcase does not call panic(), the error is returned by this function
func RunComposeTest(config string, testcode func(cli *client.Client, stopService func(string)) error) error {
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
	if ioutil.WriteFile(dockerComposeYML, []byte(config), os.ModePerm); err != nil {
		err = errors.Wrap(err, "Unable to write docker-compose.yml")
		panic(err)
	}

	// Setup running docker-compose up -d and the opposite docker-compose down -v
	cmd := exec.Command(dockerComposePath, "-f", dockerComposeYML, "up", "-d")
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		err = errors.Wrapf(err, "Unable to start docker-compose")
		panic(err)
	}
	defer func() {
		cmd := exec.Command(dockerComposePath, "-f", dockerComposeYML, "down", "-v")
		cmd.Dir = tmpDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			err = errors.Wrap(err, "Unable to stop docker-compose")
			panic(err)
		}
	}()

	// finally run the testcode
	return testcode(cli, func(service string) {
		cmd := exec.Command(dockerComposePath, "-f", dockerComposeYML, "stop", service)
		cmd.Dir = tmpDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			err = errors.Wrap(err, "Unable to stop service")
			panic(err)
		}
	})
}
