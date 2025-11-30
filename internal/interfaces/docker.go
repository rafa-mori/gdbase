package interfaces

import (
	"context"
	"io"

	"github.com/docker/go-connections/nat"
	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"

	c "github.com/docker/docker/api/types/container"
	i "github.com/docker/docker/api/types/image"
	n "github.com/docker/docker/api/types/network"
	v "github.com/docker/docker/api/types/volume"
	o "github.com/opencontainers/image-spec/specs-go/v1"
)

// Services represents a Docker service configuration.
// It preceeds the DockerService struct and is used to manage
// the state of various services running in Docker containers.
type Services struct {
	Name     string
	Image    string
	Env      map[string]string
	Ports    []nat.PortMap
	Volumes  map[string]struct{}
	StateMap map[string]any
}

// StructuredVolume represents a structured volume configuration
type StructuredVolume struct {
	Name          string
	HostPath      string
	ContainerPath string
}

type IDockerClient interface {
	ContainerStop(ctx context.Context, containerID string, options c.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options c.RemoveOptions) error
	ContainerList(ctx context.Context, options c.ListOptions) ([]c.Summary, error)
	ContainerCreate(ctx context.Context, config *c.Config, hostConfig *c.HostConfig, networkingConfig *n.NetworkingConfig, platform *o.Platform, containerName string) (c.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options c.StartOptions) error
	VolumeCreate(ctx context.Context, options v.CreateOptions) (v.Volume, error)
	VolumeList(ctx context.Context, options v.ListOptions) (v.ListResponse, error)
	ImagePull(ctx context.Context, image string, options i.PullOptions) (io.ReadCloser, error)
}

type IDockerService interface {
	Error() string
	IDockerUtils
	IContainerVolumeReport
	IContainerImageReport
	IContainerNameReport

	Initialize() error
	InitializeWithConfig(ctx context.Context, config *kbx.RootConfig) error
	StartContainer(serviceName, image string, envVars map[string]string, portBindings map[nat.Port]struct{}, volumes map[string]struct{}) error
	CreateVolume(volumeName, devicePath string) error
	GetContainerLogs(ctx context.Context, containerName string, follow bool) error
	GetProperty(name string) any
	GetContainersList() ([]c.Summary, error)
	GetVolumesList() ([]*v.Volume, error)
	StartContainerByName(containerName string) error
	StopContainerByName(containerName string, options c.StopOptions) error
	On(name string, event string, callback func(...any))
	Off(name string, event string)
	AddService(name string, image string, env map[string]string, ports []nat.PortMap, volumes map[string]struct{}) *Services
}
type IContainerVolumeReport interface {
	GetStructuredVolume(volumeName, pathsForBind string) (StructuredVolume, error)
	GetStructuredVolumes(volumes []string) ([]StructuredVolume, error)
	GetStructuredVolumesMap(volumes []string) (map[string]StructuredVolume, error)
}

type IContainerImageReport interface {
	// GetImageMap returns a map of container images with their names as keys.
	GetImageMap(images []string) map[string]struct{}
}

type IDockerUtils interface {
	GetStateMap(stateMap map[string]any) map[string]any
	MapPorts(hostPort, containerPort string) nat.PortMap
	MapPortsSlice(hostPorts, containerPorts []string) []nat.PortMap
	ExtractPorts(portBindings map[nat.Port]struct{}) nat.PortSet
	GetPortsMap(ports []nat.PortMap) map[string][]nat.PortBinding
	GetEnvSlice(envMap map[string]struct{}) []string
	GetEnvMap(envs []string) map[string]struct{}
}

type IContainerNameReport interface {
	// GetName returns the container name based on the service name.
	GetName(args ...any) string
	// GetNames returns a slice of container names based on the provided service names.
	GetNames(args ...any) []string
	// GetNamesMap returns a map of container names based on the provided service names.
	GetNamesMap(services []string) map[string]struct{}
}

// IDockerServiceReport defines the interface for reporting Docker service statistics.
type IDockerServiceReport interface {
	// GeneralStats returns general statistics of the Docker service.
	GeneralStats() (map[string]any, error)
	// ServiceStats returns statistics of a specific Docker service.
	ServiceStats(serviceName string) (map[string]any, error)
	// ServiceStatsStream returns a stream of statistics for a specific Docker service.
	ServiceStatsStream(serviceName string) (map[string]any, error)
	// ServiceStatsMap returns a map of statistics for multiple Docker services.
	ServiceStatsMap(services []string) (map[string]map[string]any, error)
	// ServiceStatsStreamMap returns a map of streamed statistics for multiple Docker services.
	ServiceStatsStreamMap(services []string) (map[string]map[string]any, error)
	// Note: The actual implementation of these methods would depend on the Docker client library being used.
}
