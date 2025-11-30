package services

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	ds "github.com/docker/docker/api/types/container"
	cf "github.com/kubex-ecosystem/gdbase/internal/bootstrap"
	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"

	"github.com/docker/docker/client"
	nl "github.com/docker/docker/libnetwork/netlabel"
	"github.com/docker/go-connections/nat"
	t "github.com/kubex-ecosystem/gdbase/internal/types"
)

// initDBSQL is an embedded SQL file that initializes the database
// with a complete schema and some initial data for testing purposes.
// The default database is implemented in PostgreSQL and can provide a simple, but complete
// database for working with almost any comercial scenario for products selling.

var (
	containersCache map[string]*ci.Services
	initDBSQLFiles  embed.FS = cf.MigrationFiles
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

// Config represents a generic configuration type for different DATABASE services.
type Config[T t.Database | t.Redis | t.RabbitMQ | t.MongoDB] = *T
type Configs[T t.Database | t.Redis | t.RabbitMQ | t.MongoDB] = map[reflect.Type]Config[T]

type IContainerVolumeReport interface {
	GetStructuredVolume(volumeName, pathsForBind string) (StructuredVolume, error)
	GetStructuredVolumes(volumes []string) ([]StructuredVolume, error)
	GetStructuredVolumesMap(volumes []string) (map[string]StructuredVolume, error)
}
type ContainerVolumeReport struct{}

func NewContainerVolumeReport() *ContainerVolumeReport { return &ContainerVolumeReport{} }

func (cvm *ContainerVolumeReport) GetStructuredVolume(volumeName, pathsForBind string) (StructuredVolume, error) {
	var volStructuredList = StructuredVolume{Name: volumeName}
	pathsArr := strings.Split(pathsForBind, ":")
	if len(pathsArr) > 1 {
		volStructuredList.HostPath = pathsArr[0]
		volStructuredList.ContainerPath = pathsArr[1]
	} else {
		volStructuredList.HostPath = pathsForBind
		volStructuredList.ContainerPath = pathsForBind
	}
	return volStructuredList, nil
}
func (cvm *ContainerVolumeReport) GetStructuredVolumes(volumes []string) ([]StructuredVolume, error) {
	var volStructuredList []StructuredVolume
	for _, volume := range volumes {
		vol, err := cvm.GetStructuredVolume(volume, volume)
		if err != nil {
			return nil, err
		}
		volStructuredList = append(volStructuredList, vol)
	}
	return volStructuredList, nil
}
func (cvm *ContainerVolumeReport) GetStructuredVolumesMap(volumes []string) (map[string]StructuredVolume, error) {
	volStructuredList, err := cvm.GetStructuredVolumes(volumes)
	if err != nil {
		return nil, err
	}
	volMap := make(map[string]StructuredVolume)
	for _, vol := range volStructuredList {
		volMap[vol.Name] = vol
	}
	return volMap, nil
}

type IContainerImageReport interface {
	// GetImageMap returns a map of container images with their names as keys.
	GetImageMap(images []string) map[string]struct{}
}
type ContainerImageReport struct{}

func NewContainerImageReport() *ContainerImageReport { return &ContainerImageReport{} }

func (cim *ContainerImageReport) GetImageMap(images []string) map[string]struct{} {
	imageMap := make(map[string]struct{})
	for _, image := range images {
		imageMap[image] = struct{}{}
	}
	return imageMap
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
type DockerUtils struct{}

func NewDockerUtils() *DockerUtils { return &DockerUtils{} }

func (d *DockerUtils) GetStateMap(stateMap map[string]any) map[string]any {
	if stateMap == nil {
		return make(map[string]any)
	}
	newStateMap := make(map[string]any, len(stateMap))
	for key, value := range stateMap {
		newStateMap[key] = value
	}
	return newStateMap
}
func (d *DockerUtils) MapPorts(hostPort, containerPort string) nat.PortMap {
	hostPort = strings.TrimSuffix(hostPort, "/tcp")
	hostPortBinding := nat.PortBinding{
		HostIP:   nl.HostIPv4,
		HostPort: hostPort,
	}
	containerPort = strings.TrimSuffix(containerPort, "/tcp")
	prtPort := nat.Port(containerPort + "/tcp")
	portBindings := nat.PortMap{
		prtPort: []nat.PortBinding{hostPortBinding},
	}
	return portBindings
}
func (d *DockerUtils) MapPortsSlice(hostPorts, containerPorts []string) []nat.PortMap {
	if len(hostPorts) != len(containerPorts) {
		return nil
	}
	portMaps := make([]nat.PortMap, len(hostPorts))
	for i := range hostPorts {
		portMaps[i] = d.MapPorts(hostPorts[i], containerPorts[i])
	}
	return portMaps
}
func (d *DockerUtils) ExtractPorts(portBindings map[nat.Port]struct{}) nat.PortSet {
	portSet := nat.PortSet{}
	for port := range portBindings {
		portSet[port] = struct{}{}
	}
	return portSet
}
func (d *DockerUtils) GetPortsMap(ports []nat.PortMap) map[string][]nat.PortBinding {
	portsMap := make(map[string][]nat.PortBinding)
	for _, port := range ports {
		for portKey, portValue := range port {
			portsMap[portKey.Port()] = portValue
		}
	}
	return portsMap
}
func (d *DockerUtils) GetEnvSlice(envMap map[string]struct{}) []string {
	envSlice := make([]string, 0, len(envMap))
	for env := range envMap {
		envSlice = append(envSlice, env)
	}
	return envSlice
}
func (d *DockerUtils) GetEnvMap(envs []string) map[string]struct{} {
	envMap := make(map[string]struct{})
	for _, env := range envs {
		if strings.Contains(env, "=") {
			envParts := strings.SplitN(env, "=", 2)
			if len(envParts) == 2 {
				envMap[envParts[0]] = struct{}{}
			}
		} else {
			envMap[env] = struct{}{}
		}
	}
	return envMap
}

type IContainerNameReport interface {
	// GetName returns the container name based on the service name.
	GetName(args ...any) string
	// GetNames returns a slice of container names based on the provided service names.
	GetNames(args ...any) []string
	// GetNamesMap returns a map of container names based on the provided service names.
	GetNamesMap(services []string) map[string]struct{}
}
type ContainerNameReport struct{}

func NewContainerNameReport() *ContainerNameReport { return &ContainerNameReport{} }

func (cnr *ContainerNameReport) GetName(args ...any) string {
	if len(args) < 1 {
		return ""
	}
	serviceName, ok := args[0].(string)
	if !ok {
		return ""
	}
	if strings.HasPrefix(serviceName, "gdbase-") {
		return serviceName
	}
	return "gdbase-" + serviceName
}
func (cnr *ContainerNameReport) GetNames(args ...any) []string {
	containerNames := make([]string, len(args))
	if len(args) == 0 {
		return containerNames
	}
	services, ok := args[0].([]string)
	if !ok || len(services) == 0 {
		return containerNames
	}
	if len(services) == 1 {
		containerNames[0] = cnr.GetName(services[0])
		return containerNames
	}
	containerNames = make([]string, len(services))
	if len(services) == 0 {
		return containerNames
	}
	if len(services) == 1 {
		containerNames[0] = cnr.GetName(services[0])
		return containerNames
	}
	return containerNames
}
func (cnr *ContainerNameReport) GetNamesMap(services []string) map[string]struct{} {
	containerNames := cnr.GetNames(services)
	containerNamesMap := make(map[string]struct{}, len(containerNames))
	for _, name := range containerNames {
		containerNamesMap[name] = struct{}{}
	}
	return containerNamesMap
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

// DockerServiceReport agora possui um cliente Docker embutido.
type DockerServiceReport struct {
	Cli *client.Client
}

// NewDockerServiceReport cria e retorna um novo DockerServiceReport, inicializando o client.
func NewDockerServiceReport() *DockerServiceReport {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		// Em um cenário real, trate o erro apropriadamente
		panic(fmt.Sprintf("Erro ao criar Docker client: %v", err))
	}
	return &DockerServiceReport{Cli: cli}
}

// GeneralStats retorna informações gerais do sistema Docker.
// Usa o método Info do SDK e converte o resultado para map[string]any.
func (dsr *DockerServiceReport) GeneralStats() (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, err := dsr.Cli.Info(ctx)
	if err != nil {
		return nil, err
	}
	// Convertendo o objeto types.Info para mapa por meio de JSON.
	data, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ServiceStats retorna as estatísticas de um container específico (não stream).
func (dsr *DockerServiceReport) ServiceStats(serviceName string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	containerID, err := dsr.findContainerIDByName(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	// Solicita as estatísticas em modo "não streaming".
	stats, err := dsr.Cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, err
	}
	defer stats.Body.Close()

	var statsJSON ds.StatsResponse
	decoder := json.NewDecoder(stats.Body)
	if err := decoder.Decode(&statsJSON); err != nil {
		return nil, err
	}

	// Converte o struct StatsJSON para map[string]any.
	data, err := json.Marshal(statsJSON)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ServiceStatsStream retorna as estatísticas de um container específico em modo stream,
// mas lê e retorna apenas o primeiro conjunto de dados.
func (dsr *DockerServiceReport) ServiceStatsStream(serviceName string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	containerID, err := dsr.findContainerIDByName(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	// Solicita estatísticas em modo streaming.
	stats, err := dsr.Cli.ContainerStats(ctx, containerID, true)
	if err != nil {
		return nil, err
	}
	defer stats.Body.Close()

	var statsJSON ds.StatsResponse
	decoder := json.NewDecoder(stats.Body)
	// Lê o primeiro objeto de estatísticas da stream.
	if err := decoder.Decode(&statsJSON); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("stream de stats terminou prematuramente")
		}
		return nil, err
	}

	data, err := json.Marshal(statsJSON)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ServiceStatsMap obtém estatísticas de cada container na lista de serviços e retorna um mapa.
func (dsr *DockerServiceReport) ServiceStatsMap(services []string) (map[string]map[string]any, error) {
	statsMap := make(map[string]map[string]any, len(services))
	for _, service := range services {
		stats, err := dsr.ServiceStats(service)
		if err != nil {
			return nil, err
		}
		statsMap[service] = stats
	}
	return statsMap, nil
}

// ServiceStatsStreamMap obtém as estatísticas em modo stream para cada container e retorna um mapa.
func (dsr *DockerServiceReport) ServiceStatsStreamMap(services []string) (map[string]map[string]any, error) {
	statsMap := make(map[string]map[string]any, len(services))
	for _, service := range services {
		stats, err := dsr.ServiceStatsStream(service)
		if err != nil {
			return nil, err
		}
		statsMap[service] = stats
	}
	return statsMap, nil
}

// findContainerIDByName auxilia na busca do container ID dado um nome de serviço.
func (dsr *DockerServiceReport) findContainerIDByName(ctx context.Context, serviceName string) (string, error) {
	containers, err := dsr.Cli.ContainerList(ctx, ds.ListOptions{All: true})
	if err != nil {
		return "", err
	}
	for _, cnt := range containers {
		for _, name := range cnt.Names {
			// Considera que o nome pode vir com prefixo "/"
			if name == "/"+serviceName || name == serviceName {
				return cnt.ID, nil
			}
		}
	}
	return "", fmt.Errorf("container com o nome %s não foi encontrado", serviceName)
}
