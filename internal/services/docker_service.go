package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/go-connections/nat"

	c "github.com/docker/docker/api/types/container"
	i "github.com/docker/docker/api/types/image"
	v "github.com/docker/docker/api/types/volume"
	k "github.com/docker/docker/client"
	nl "github.com/docker/docker/libnetwork/netlabel"

	evs "github.com/kubex-ecosystem/gdbase/internal/events"
	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"
	it "github.com/kubex-ecosystem/gdbase/internal/types"
	"github.com/kubex-ecosystem/logz"
	gl "github.com/kubex-ecosystem/logz"
	l "github.com/kubex-ecosystem/logz"

	_ "embed"
)

func NewServices(name, image string, env []string, ports []nat.PortMap, volumes map[string]struct{}) *Services {
	if containersCache == nil {
		containersCache = make(map[string]*Services)
	}
	service := &Services{
		Name:     name,
		Image:    image,
		Env:      env,
		Ports:    ports,
		Volumes:  volumes,
		StateMap: make(map[string]any),
	}
	if _, ok := containersCache[name]; !ok {
		containersCache[name] = service
	} else {
		containersCache[name].Name = name
		containersCache[name].Image = image
		containersCache[name].Env = env
		containersCache[name].Ports = ports
		containersCache[name].Volumes = volumes
	}
	return service
}

type IDockerService interface {
	IDockerUtils
	IContainerVolumeReport
	IContainerImageReport
	IContainerNameReport

	Initialize() error
	StartContainer(serviceName, image string, envVars []string, portBindings map[nat.Port]struct{}, volumes map[string]struct{}) error
	CreateVolume(volumeName, devicePath string) error
	GetContainerLogs(ctx context.Context, containerName string, follow bool) error
	GetProperty(name string) any
	GetContainersList() ([]c.Summary, error)
	GetVolumesList() ([]*v.Volume, error)
	StartContainerByName(containerName string) error
	StopContainerByName(containerName string, options c.StopOptions) error
	On(name string, event string, callback func(...any))
	Off(name string, event string)
	AddService(name string, image string, env []string, ports []nat.PortMap, volumes map[string]struct{}) *Services
}
type DockerService struct {
	*ContainerNameReport
	*ContainerImageReport
	*ContainerVolumeReport
	*DockerUtils

	Logger    *logz.LoggerZ
	reference *it.Reference
	mutexes   *it.Mutexes

	services map[string]any

	Cli  IDockerClient
	pool *sync.Pool

	properties map[string]any
	eventBus   *evs.EventBus
}

func newDockerServiceBus(config *DBConfig, logger *logz.LoggerZ) (IDockerService, error) {
	EnsureDockerIsRunning()

	if logger == nil {
		logger = l.GetLogger("DockerService")
	}

	var propDBConfig ci.IProperty[*DBConfig]
	if config != nil {
		propDBConfig = it.NewProperty[*DBConfig]("dbConfig", &config, false, nil)
	}

	cli, err := k.NewClientWithOpts(k.FromEnv, k.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("❌ Error creating Docker client: %v", err)
	}
	dockerService := &DockerService{
		Logger:     logger,
		reference:  it.NewReference("DockerService").GetReference(),
		mutexes:    it.NewMutexesType(),
		pool:       &sync.Pool{},
		Cli:        cli,
		properties: nil,

		DockerUtils:           NewDockerUtils(),
		ContainerNameReport:   NewContainerNameReport(),
		ContainerImageReport:  NewContainerImageReport(),
		ContainerVolumeReport: NewContainerVolumeReport(),
	}
	if config != nil {
		dockerService.properties = map[string]any{"dbConfig": propDBConfig}
	}
	if dockerService.eventBus == nil {
		dockerService.eventBus = evs.NewEventBus()
	}
	return dockerService, nil
}
func newDockerService(config *DBConfig, logger *logz.LoggerZ) (IDockerService, error) {
	EnsureDockerIsRunning()
	return newDockerServiceBus(config, logger)
}
func NewDockerService(config *DBConfig, logger *logz.LoggerZ) (IDockerService, error) {
	return newDockerService(config, logger)
}

func (d *DockerService) GetContainerLogs(ctx context.Context, containerName string, follow bool) error {
	cli, err := k.NewClientWithOpts(k.FromEnv, k.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("error creating Docker client: %w", err)
	}

	logsReader, err := cli.ContainerLogs(ctx, containerName, c.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     follow,
	})
	if err != nil {
		return fmt.Errorf("error getting logs for container %s: %w", containerName, err)
	}
	defer func(logsReader io.ReadCloser) {
		_ = logsReader.Close()
	}(logsReader)

	scanner := bufio.NewScanner(logsReader)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if scannerErr := scanner.Err(); scannerErr != nil {
		return fmt.Errorf("error processing logs for container %s: %w", containerName, scannerErr)
	}
	return nil
}
func (d *DockerService) Initialize() error {
	if d.properties != nil {
		dbServiceConfigT, exists := d.properties["dbConfig"]
		if exists {
			if dbServiceConfig, ok := dbServiceConfigT.(*it.Property[*DBConfig]); !ok {
				return fmt.Errorf("❌ Error converting database configuration")
			} else {
				dbSrvCfg := dbServiceConfig.GetValue()
				if err := SetupDatabaseServices(context.Background(), d, dbSrvCfg); err != nil {
					return fmt.Errorf("❌ Error setting up database services: %v", err)
				}
				d.properties["dbConfig"] = dbServiceConfig
			}
		} else {
			return fmt.Errorf("❌ Database configuration not found")
		}
	}

	d.properties["volumes"] = make(map[string]map[string]struct{})
	d.properties["services"] = make(map[string]string)

	return nil
}
func (d *DockerService) StartContainer(serviceName, image string, envVars []string, portBindings map[nat.Port]struct{}, volumes map[string]struct{}) error {
	if !isDockerRunning() {
		gl.Log("fatal", "Docker is not running. Please start Docker and try again.")
		return fmt.Errorf("docker is not running")
	}

	if IsServiceRunning(serviceName) {
		fmt.Printf("✅ %s is already running!\n", serviceName)
		return nil
	}

	ctx := context.Background()

	fmt.Println("🔄 Pulling image...")
	reader, err := d.Cli.ImagePull(ctx, image, i.PullOptions{})
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error pulling image: %v", err))
		return fmt.Errorf("error pulling image: %w", err)
	}
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)
	_, _ = io.Copy(io.Discard, reader)

	fmt.Println("🚀 Creating container...")
	containerConfig := &c.Config{
		Image:        image,
		Env:          envVars,
		ExposedPorts: d.ExtractPorts(portBindings),
	}

	binds := []string{}

	for volume := range volumes {
		// Por enquanto coloquei os campos repetidos, mas depois PRECISAMOS melhorar isso
		structuredVolume, err := d.GetStructuredVolume(volume, volume)
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error getting structured volume: %v", err))
			return fmt.Errorf("error getting structured volume: %w", err)
		}

		binds = append(binds, fmt.Sprintf("%s:%s", structuredVolume.HostPath, structuredVolume.ContainerPath))
	}

	portBindingsT := make(nat.PortMap)
	for hostPort := range portBindings {
		containerPort := strings.TrimSuffix(hostPort.Port(), "/tcp")
		hostPortBinding := nat.PortBinding{
			HostIP:   nl.HostIPv4,
			HostPort: hostPort.Port(),
		}
		prtPort := nat.Port(containerPort + "/tcp")
		portBindingsT[prtPort] = []nat.PortBinding{hostPortBinding}
	}

	hostConfig := &c.HostConfig{
		Binds:        binds,
		PortBindings: portBindingsT,
		RestartPolicy: c.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	resp, err := d.Cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, serviceName)
	if err != nil {
		return fmt.Errorf("error creating container %s: %w", serviceName, err)
	}

	if err := d.Cli.ContainerStart(ctx, resp.ID, c.StartOptions{}); err != nil {
		return fmt.Errorf("error starting container %s: %w", serviceName, err)
	}

	fmt.Println("✅ Container started successfully!")
	return nil
}
func (d *DockerService) CreateVolume(volumeName, pathsForBind string) error {
	structuredVolume, err := d.GetStructuredVolume(volumeName, pathsForBind)
	if err != nil {
		return fmt.Errorf("error getting structured volume: %w", err)
	}
	ctx := context.Background()

	volumes, _ := d.Cli.VolumeList(ctx, v.ListOptions{})
	for _, vol := range volumes.Volumes {
		if vol.Name == volumeName {
			gl.Log("debug", fmt.Sprintf("Volume %s already exists, skipping creation", volumeName))
			return nil
		}
	}

	if filepath.IsAbs(structuredVolume.HostPath) {
		var createOpts v.CreateOptions
		if structuredVolume.HostPath == "" {
			createOpts = v.CreateOptions{
				Name:   structuredVolume.Name,
				Labels: map[string]string{"created_by": "gdbase"},
			}
		} else {
			// Ensure the host path exists
			// if err := ensureDirWithOwner(structuredVolume.HostPath, os.Getuid(), os.Getgid(), 0755); err != nil {
			// 	return fmt.Errorf("error ensuring host path %s exists: %w", structuredVolume.HostPath, err)
			// }
			createOpts = v.CreateOptions{
				Name:   structuredVolume.Name,
				Labels: map[string]string{"created_by": "gdbase"},
				Driver: "local",
				DriverOpts: map[string]string{
					"type":   "none",
					"device": structuredVolume.HostPath,
					"o":      "bind,rbind,rshared",
				},
			}
		}

		// Create the volume with the bind mount option1
		vol, err := d.Cli.VolumeCreate(ctx, createOpts)

		if err != nil {
			return err
		}

		gl.Log("info", fmt.Sprintf("Volume %s created at %s", vol.Name, structuredVolume.HostPath))
	}

	return nil
}
func (d *DockerService) GetContainersList() ([]c.Summary, error) {
	containers, err := d.Cli.ContainerList(context.Background(), c.ListOptions{All: true})
	if err != nil {
		panic(err)
	}

	var containerList []c.Summary
	for _, container := range containers {
		if container.State == "running" {
			containerList = append(containerList, container)
		}
	}

	return containerList, nil
}
func (d *DockerService) GetVolumesList() ([]*v.Volume, error) {
	volumes, err := d.Cli.VolumeList(context.Background(), v.ListOptions{})
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error listing volumes: %v", err))
	}

	var volumeList []*v.Volume
	for _, volume := range volumes.Volumes {
		if volume.Name == "gdbase-pg-data" /* || volume.Name != "gdbase-redis-data" */ {
			volumeList = append(volumeList, volume)
		}
	}

	return volumeList, nil
}
func (d *DockerService) StartContainerByName(containerName string) error {
	ctx := context.Background()
	err := d.Cli.ContainerStart(ctx, containerName, c.StartOptions{})
	if err != nil {
		return fmt.Errorf("error starting container %s: %w", containerName, err)
	}
	fmt.Printf("✅ Container %s started successfully!\n", containerName)
	return nil
}
func (d *DockerService) StopContainerByName(containerName string, stopOptions c.StopOptions) error {
	ctx := context.Background()
	err := d.Cli.ContainerStop(ctx, containerName, stopOptions)
	if err != nil {
		return fmt.Errorf("error stopping container %s: %w", containerName, err)
	}
	fmt.Printf("✅ Container %s stopped successfully!\n", containerName)
	return nil
}
func (d *DockerService) GetProperty(name string) any {
	if prop, ok := d.properties[name]; ok {
		return prop
	}
	return nil
}
func (d *DockerService) On(name string, event string, callback func(...any)) {
	if d.mutexes == nil {
		d.mutexes = it.NewMutexesType()
	}
	if d.pool == nil {
		d.pool = &sync.Pool{}
	}
	// d.mutexes.MuRLock()
	// defer d.mutexes.MuRUnlock()
	if callback != nil {
		d.pool.Put(callback)
	}
}
func (d *DockerService) Off(name string, event string) {
	if d.mutexes == nil {
		d.mutexes = it.NewMutexesType()
	}
	if d.pool == nil {
		d.pool = &sync.Pool{}
	}
	// d.mutexes.MuRLock()
	// defer d.mutexes.MuRUnlock()
	d.pool.Put(nil)
}
func (d *DockerService) GetContainersCache() map[string]*Services {
	if containersCache == nil {
		containersCache = make(map[string]*Services)
	}
	return containersCache
}
func (d *DockerService) GetEventBus() *evs.EventBus {
	if d.eventBus == nil {
		d.eventBus = evs.NewEventBus()
	}
	return d.eventBus
}
func (d *DockerService) AddService(name string, image string, env []string, ports []nat.PortMap, volumes map[string]struct{}) *Services {
	if containersCache == nil {
		containersCache = make(map[string]*Services)
	}
	service := &Services{
		Name:     name,
		Image:    image,
		Env:      env,
		Ports:    ports,
		Volumes:  volumes,
		StateMap: make(map[string]any),
	}
	if d.services == nil {
		d.services = make(map[string]any)
	}

	d.services[name] = service

	if _, ok := containersCache[name]; !ok {
		containersCache[name] = service
	} else {
		containersCache[name].Name = name
		containersCache[name].Image = image
		containersCache[name].Env = env
		containersCache[name].Ports = ports
		containersCache[name].Volumes = volumes
	}
	return service
}
