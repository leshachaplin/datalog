package clickhouse

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

const defaultPort = "9000/tcp"

var hostName = os.Getenv("OVERRIDE_HOSTNAME")

func init() {
	const defaultHostName = "localhost"

	if hostName == "" {
		hostName = defaultHostName
	}
}

type Container struct {
	resource *dockertest.Resource
}

func NewContainer(connectFn func(connURL string) error) (*Container, error) {
	hostPort, err := getFreePort()
	if err != nil {
		return nil, fmt.Errorf("could not get free hostPort1: %w", err)
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "clickhouse/clickhouse-server",
			Tag:        "latest-alpine",
			Env: []string{
				"CLICKHOUSE_DB=test_db",
				"CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1",
				"CLICKHOUSE_USER=su",
				"CLICKHOUSE_PASSWORD=su",
				"--ulimit nofile=262144:262144",
			},
			Entrypoint: nil,
			PortBindings: map[docker.Port][]docker.PortBinding{
				"9000/tcp": {{
					HostIP:   hostName,
					HostPort: strconv.Itoa(hostPort),
				}},
			},
		}, func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		})
	if err != nil {
		return nil, fmt.Errorf("could not create a container: %w", err)
	}

	container := &Container{
		resource: resource,
	}
	addr := fmt.Sprintf("%s:%s", hostName, resource.GetPort(defaultPort))
	if err := pool.Retry(func() error {
		return connectFn(addr)
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	return container, nil
}

func (c *Container) Purge() error {
	return c.resource.Close()
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
