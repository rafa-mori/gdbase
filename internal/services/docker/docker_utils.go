package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"time"

	ds "github.com/docker/docker/api/types/container"
	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"
	"github.com/kubex-ecosystem/gdbase/internal/types"
	logz "github.com/kubex-ecosystem/logz"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// initDBSQL is an embedded SQL file that initializes the database
// with a complete schema and some initial data for testing purposes.
// The default database is implemented in PostgreSQL and can provide a simple, but complete
// database for working with almost any comercial scenario for products selling.

var (
	containersCache map[string]*ci.Services
	// initDBSQLFiles  embed.FS = cf.MigrationFiles
)

// Config represents a generic configuration type for different DATABASE services.
type Config[T types.DBConfig] = *T
type Configs[T types.DBConfig] = map[reflect.Type]Config[T]

type ContainerVolumeReport struct{}

func NewContainerVolumeReport() *ContainerVolumeReport { return &ContainerVolumeReport{} }

func (cvm *ContainerVolumeReport) GetStructuredVolume(volumeName, pathsForBind string) (ci.StructuredVolume, error) {
	var volStructuredList = ci.StructuredVolume{Name: volumeName}
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
func (cvm *ContainerVolumeReport) GetStructuredVolumes(volumes []string) ([]ci.StructuredVolume, error) {
	var volStructuredList []ci.StructuredVolume
	for _, volume := range volumes {
		vol, err := cvm.GetStructuredVolume(volume, volume)
		if err != nil {
			return nil, err
		}
		volStructuredList = append(volStructuredList, vol)
	}
	return volStructuredList, nil
}
func (cvm *ContainerVolumeReport) GetStructuredVolumesMap(volumes []string) (map[string]ci.StructuredVolume, error) {
	volStructuredList, err := cvm.GetStructuredVolumes(volumes)
	if err != nil {
		return nil, err
	}
	volMap := make(map[string]ci.StructuredVolume)
	for _, vol := range volStructuredList {
		volMap[vol.Name] = vol
	}
	return volMap, nil
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
		HostIP:   ResolveHostIP(),
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
	if strings.HasPrefix(serviceName, "kubexdb-") {
		return serviceName
	}
	return "kubexdb-" + serviceName
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

func IsServiceRunning(serviceName string) bool {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", serviceName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		logz.Log("error", fmt.Sprintf("❌ Error checking containers: %v\n", err))
	}
	return string(output) != ""
}
func IsDockerRunning() bool {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("systemctl", "is-active", "--quiet", "docker")
	case "darwin": // macOS
		cmd = exec.Command("brew", "services", "list") // Apenas para verificar se o serviço está ativo
	case "windows":
		cmd = exec.Command("powershell", "Get-Service", "docker", "|", "Select-Object", "Status")
	default:
		fmt.Println("Sistema operacional não suportado para ativação automática do Docker.")
		return false
	}

	err := cmd.Run()
	return err == nil
}
func AskToStartDocker() bool {

	// Exibe uma mensagem de aviso ao usuário
	fmt.Printf("\033[1;36m%s\033[0m", `
┌──────────────────────────────────────────────────────────────┐
│                                                              │
│  O serviço do Docker não está ativo.                         │
│                                                              │
│  Para continuar, precisamos ativá-lo.                        │
│  Podemos iniciar o Docker automaticamente para você ou       │
│  você pode ativá-lo manualmente.                             │
│                                                              │
│  Se você deseja iniciar o Docker automaticamente,            │
│  pressione 'Y' e 'Enter'.                                    │
│  Caso contrário, pressione 'N' e 'Enter' para sair.          │
│                                                              │
└──────────────────────────────────────────────────────────────┘`)
	fmt.Printf("\033[1;36m%s\033[0m", `
┌──────────────────────────────────────────────────────────────┐
│                                                              │
│  Se você não tem certeza de como ativá-lo manualmente,       │
│  pode siguir as instruções abaixo em seu terminal:           │
│                                                              │
│  Linux: sudo systemctl start docker                          │
│  macOS: brew services start docker                           │
│  Windows: powershell Start-Service docker                    │
│                                                              │
└──────────────────────────────────────────────────────────────┘`)
	fmt.Printf("\033[1;33m%s\033[0m", `
┌──────────────────────────────────────────────────────────────┐
│                                                              │
│  Pressione 'Y' para iniciar o Docker automaticamente         │
│  ou 'N' para sair do programa. (15 segundos para resposta)   │
│                                                              │
└──────────────────────────────────────────────────────────────┘
`)

	fmt.Print("Resposta: ")

	// Canal para receber a resposta do usuário
	// O canal é usado para evitar o bloqueio do terminal enquanto espera a entrada do usuário
	// e permite que o programa continue executando.
	responseCh := make(chan string, 1)
	defer close(responseCh)

	// Lê a resposta do usuário em uma goroutine
	// Isso permite que o programa continue executando enquanto espera a entrada do usuário
	// e evita o bloqueio do terminal. O usuário pode pressionar 'Y' ou 'N', caso contrário
	// o programa irá aguardar 15 segundos antes de encerrar.
	go func() {
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToUpper(response))
		responseCh <- response
	}()

	// Aguarda a resposta do usuário ou timeout de 15 segundos
	select {
	case response := <-responseCh:
		return response == "Y"
	case <-time.After(15 * time.Second):
		logz.Log("warn", "\nTempo esgotado. Docker não será ativado.")
		return false
	}
}
func ResolveHostIP() string {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return "host.docker.internal"
	}
	return "127.0.0.1"
}

// Ativa o serviço do Docker dinamicamente conforme o sistema operacional

func StartDockerService() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("sudo", "systemctl", "start", "docker")
	case "darwin": // macOS
		cmd = exec.Command("brew", "services", "start", "docker")
	case "windows":
		cmd = exec.Command("powershell", "Start-Service", "docker")
	default:
		return fmt.Errorf("sistema operacional não suportado para ativação automática do Docker")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func EnsureDockerIsRunning() {
	if !IsDockerRunning() {
		isInteractive := os.Getenv("IS_INTERACTIVE") == "true" || os.Getenv("IS_INTERACTIVE") == "1"
		if isInteractive {
			if AskToStartDocker() {
				logz.Log("info", "Starting Docker service...")
				if err := StartDockerService(); err != nil {
					logz.Log("fatal", fmt.Sprintf("Error starting Docker: %v", err))
				}
				logz.Log("success", "Docker service started successfully!")
			} else {
				logz.Log("warn", "Docker service is not running and user chose not to start it.")
				logz.Log("warn", "Please start Docker manually to continue.")
				logz.Log("warn", "Exiting...")
				os.Exit(1)
			}
		} else {
			if err := StartDockerService(); err != nil {
				logz.Log("fatal", fmt.Sprintf("Error starting Docker: %v", err))
			}
		}
	}
}
