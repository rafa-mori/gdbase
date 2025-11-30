package services

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	u "github.com/kubex-ecosystem/gdbase/utils"
	gl "github.com/kubex-ecosystem/logz"
)

func SlitMessage(recPayload []string) (id, msg []string) {
	if recPayload[1] == "" {
		id = recPayload[:2]
		msg = recPayload[2:]
	} else {
		id = recPayload[:1]
		msg = recPayload[1:]
	}
	return
}
func GetBrokersPath() (string, error) {
	brkDir, homeErr := os.UserHomeDir()
	if homeErr != nil || brkDir == "" {
		brkDir, homeErr = os.UserConfigDir()
		if homeErr != nil || brkDir == "" {
			brkDir, homeErr = os.UserCacheDir()
			if homeErr != nil || brkDir == "" {
				brkDir = "/tmp"
			}
		}
	}

	brkDir = filepath.Join(brkDir, ".kubex", "gkbxsrv", "brokers")

	if _, statErr := os.Stat(brkDir); statErr != nil {
		if mkDirErr := os.MkdirAll(brkDir, 0755); mkDirErr != nil {
			gl.Log("error", "Error creating brokers")
			return "", mkDirErr
		}
	}

	gl.Log("info", fmt.Sprintf("PID's folder: %s", brkDir))

	return brkDir, nil
}
func RndomName() string {
	return "broker-" + randStringBytes(5)
}
func randStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func FindAvailablePort(basePort int, maxAttempts int) (string, error) {
	for z := 0; z < maxAttempts; z++ {
		if basePort+z < 1024 || basePort+z > 49151 {
			continue
		}
		port := fmt.Sprintf("%d", basePort+z)
		isOpen, err := u.CheckPortOpen(port)
		if err != nil {
			return "", fmt.Errorf("error checking port %s: %w", port, err)
		}
		if !isOpen {
			gl.Log("warn", fmt.Sprintf("⚠️ Port %s is occupied, trying the next one...\n", port))
			continue
		}
		gl.Log("info", fmt.Sprintf("✅ Available port found: %s\n", port))
		return port, nil
	}
	return "", fmt.Errorf("no available port in range %d-%d", basePort, basePort+maxAttempts-1)
}
func IsServiceRunning(serviceName string) bool {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", serviceName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Error checking containers: %v\n", err))
	}
	return string(output) != ""
}
func isDockerRunning() bool {
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
func WriteInitDBSQL(initVolumePath, initDBSQL, initDBSQLData string) (string, error) {
	if err := os.MkdirAll(initVolumePath, 0755); err != nil {
		gl.Log("error", fmt.Sprintf("Error creating directory: %v", err))
		return "", err
	}
	filePath := filepath.Join(initVolumePath, initDBSQL)
	if _, err := os.Stat(filePath); err == nil {
		gl.Log("debug", fmt.Sprintf("File %s already exists, skipping creation", filePath))
		return filePath, nil
	}
	if err := os.WriteFile(filePath, []byte(initDBSQLData), 0644); err != nil {
		gl.Log("error", fmt.Sprintf("Error writing file: %v", err))
		return "", err
	}
	gl.Log("info", fmt.Sprintf("✅ File %s created successfully!\n", filePath))
	return filePath, nil
}
func SetupDatabaseServices(ctx context.Context, d IDockerService, config *DBConfig) error {
	if config == nil {
		return fmt.Errorf("❌ Configuração do banco de dados não encontrada")
	}
	if config.Databases == nil {
		gl.Log("debug", "Not found databases in config, skipping DB setup")
	}
	var services = make([]*Services, 0)

	if len(config.Databases) > 0 {
		for _, dbConfig := range config.Databases {
			if dbConfig != nil {
				if dbConfig.Type == "postgresql" && dbConfig.Enabled {
					// Check if the database is already running
					if IsServiceRunning("gdbase-pg") {
						gl.Log("debug", fmt.Sprintf("✅ %s já está rodando!", "gdbase-pg"))
						continue
					} else {
						if err := d.StartContainerByName("gdbase-pg"); err == nil {
							gl.Log("debug", fmt.Sprintf("✅ %s já está rodando!", "gdbase-pg"))
							continue
						} else {
							// Check if Password is empty, if so, try to retrieve it from keyring
							// if not found, generate a new one
							if dbConfig.Password == "" {
								pgPassKey, pgPassErr := kbx.GetOrGenPasswordKeyringPass("pgpass")
								if pgPassErr != nil {
									gl.Log("error", fmt.Sprintf("Error generating key: %v", pgPassErr))
									continue
								}
								dbConfig.Password = string(pgPassKey)
							} else {
								gl.Log("debug", fmt.Sprintf("Password found in config: %s", dbConfig.Password[0:2]))
							}
							if dbConfig.Volume == "" {
								dbConfig.Volume = os.ExpandEnv(DefaultPostgresVolume)
							}
							pgVolRootDir := os.ExpandEnv(dbConfig.Volume)
							pgVolInitDir := filepath.Join(pgVolRootDir, "init")
							vols := map[string]struct{}{
								strings.Join([]string{pgVolInitDir, "/docker-entrypoint-initdb.d"}, ":"): {},
							}
							initDBSQLs, initDBSQLErr := embed.FS.ReadDir(initDBSQLFiles, "embedded")
							if initDBSQLErr != nil {
								gl.Log("error", fmt.Sprintf("❌ Erro ao ler diretório de scripts SQL: %v", initDBSQLErr))
								continue
							} else {
								for _, initDBSQL := range initDBSQLs {
									initDBSQLData, initDBSQLErr := embed.FS.ReadFile(initDBSQLFiles, filepath.Join("embedded", initDBSQL.Name()))
									if initDBSQLErr != nil {
										gl.Log("error", fmt.Sprintf("❌ Erro ao ler script SQL %s: %v", initDBSQL.Name(), initDBSQLErr))
										continue
									}
									if _, err := WriteInitDBSQL(pgVolInitDir, initDBSQL.Name(), string(initDBSQLData)); err != nil {
										gl.Log("error", fmt.Sprintf("❌ Erro ao criar diretório do PostgreSQL: %v", err))
										continue
									}
								}
								if err := d.CreateVolume("gdbase-pg-init", pgVolInitDir); err != nil {
									gl.Log("error", fmt.Sprintf("❌ Erro ao criar volume do PostgreSQL: %v", err))
									continue
								}
								pgVolDataDir := filepath.Join(pgVolRootDir, "pgdata")
								if err := d.CreateVolume("gdbase-pg-data", pgVolDataDir); err != nil {
									gl.Log("error", fmt.Sprintf("❌ Erro ao criar volume do PostgreSQL: %v", err))
									continue
								}
								vols[strings.Join([]string{pgVolDataDir, "/var/lib/postgresql/data"}, ":")] = struct{}{}
							}
							// Check if the port is already in use and find an available one if necessary
							if dbConfig.Port == nil || dbConfig.Port == "" {
								dbConfig.Port = "5432"
							}
							port, err := FindAvailablePort(5432, 10)
							if err != nil {
								gl.Log("error", fmt.Sprintf("❌ Erro ao encontrar porta disponível: %v", err))
								continue
							}
							dbConfig.Port = port
							// Map the port to the container
							portMap := d.MapPorts(fmt.Sprintf("%s", dbConfig.Port), "5432/tcp")

							// Check if the database name is empty, if so, generate a random one
							if dbConfig.Name == "" {
								dbConfig.Name = "godo-" + randStringBytes(5)
							}
							// Insert the PostgreSQL service into the services slice
							dbConnObj := NewServices(
								"gdbase-pg",
								"postgres:17-alpine",
								[]string{
									// "POSTGRES_HOST_AUTH_METHOD=trust", // Use only for development, not recommended for production
									"POSTGRES_HOST_AUTH_METHOD=trust",
									// Necessary for Postgres 12+
									"POSTGRES_INITDB_ARGS=--data-checksums",
									"POSTGRES_INITDB_ARGS=--encoding=UTF8",
									"POSTGRES_INITDB_ARGS=--locale=pt_BR.UTF-8",
									"POSTGRES_USER=" + dbConfig.Username,
									"POSTGRES_PASSWORD=" + dbConfig.Password,
									"POSTGRES_DB=" + dbConfig.Name,
									"POSTGRES_PORT=" + dbConfig.Port.(string),
									"POSTGRES_DB_NAME=" + dbConfig.Name,
									"POSTGRES_DB_VOLUME=" + dbConfig.Volume,
									"POSTGRES_DB_SSLMODE=disable",
									"POSTGRES_DB_INITDB_ARGS=--data-checksums",
									// Necessary for some clients
									"PGUSER=" + dbConfig.Username,
									"PGPASSWORD=" + dbConfig.Password,
									"PGDATABASE=" + dbConfig.Name,
									"PGPORT=" + dbConfig.Port.(string),
									"PGHOST=localhost",
									"PGDATA=/var/lib/postgresql/data/pgdata",
									"PGSSLMODE=disable",
								},
								[]nat.PortMap{portMap},
								vols,
							)
							services = append(services, dbConnObj)
						}
					}
				}
			}
		}
	} else {
		gl.Log("debug", "Not found databases in config, skipping DB setup")
	}
	if config.Messagery != nil {
		if config.Messagery.RabbitMQ != nil && config.Messagery.RabbitMQ.Enabled {
			// Check if the RabbitMQ service is already running
			if IsServiceRunning("gdbase-rabbitmq") {
				gl.Log("info", fmt.Sprintf("✅ %s já está rodando!\n", "gdbase-rabbitmq"))
			} else {
				rabbitCfg := config.Messagery.RabbitMQ
				rabbitUser := rabbitCfg.Username
				rabbitPass := rabbitCfg.Password
				if rabbitUser == "" {
					rabbitUser = "gobe"
				}
				if rabbitCfg.Password == "" {
					rabbitPassKey, rabbitPassErr := kbx.GetOrGenPasswordKeyringPass(rabbitCfg.Reference.Name)
					if rabbitPassErr != nil {
						gl.Log("error", "Skipping RabbitMQ setup due to error generating password")
						gl.Log("debug", fmt.Sprintf("Error generating key: %v", rabbitPassErr))
						goto postRabbit
					}
					rabbitPass = string(rabbitPassKey)
				} else {
					gl.Log("debug", fmt.Sprintf("Password found in config: %s...", rabbitCfg.Password[0:2]))
				}
				if rabbitCfg.Reference.Name == "" {
					rabbitCfg.Reference.Name = "gdbase-rabbitmq"
				}
				// if rabbitCfg.Volume == "" {
				// 	rabbitCfg.Volume = os.ExpandEnv(glb.DefaultRabbitMQVolume)
				// }
				if rabbitCfg.Host == "" {
					rabbitCfg.Host = "localhost"
				}
				if rabbitCfg.Port == nil || rabbitCfg.Port == "" {
					rabbitCfg.Port = "5672"
				}
				if rabbitCfg.ManagementPort == "" {
					rabbitCfg.ManagementPort = "15672"
				}
				// Check if the port is already in use and find an available one if necessary
				port, err := FindAvailablePort(5672, 10)
				if err != nil {
					gl.Log("error", fmt.Sprintf("❌ Erro ao encontrar porta disponível: %v", err))
					goto postRabbit
				}
				rabbitCfg.Port = port
				managementPort, err := FindAvailablePort(15672, 10)
				if err != nil {
					gl.Log("error", fmt.Sprintf("❌ Erro ao encontrar porta disponível: %v", err))
					goto postRabbit
				}
				rabbitCfg.ManagementPort = managementPort
				// Create the volume for RabbitMQ, if exists definitions on the config
				if err := d.CreateVolume(rabbitCfg.Reference.Name, rabbitCfg.Volume); err != nil {
					gl.Log("error", fmt.Sprintf("❌ Erro ao criar volume do RabbitMQ: %v", err))
					goto postRabbit
				}
				// Check if ErlangCookie is empty, if so, generate a new one
				// RabbitMQ nodes use the Erlang cookie to authenticate with each other.
				// If you are running a single node, it is not strictly necessary to set this value,
				// but it is a good practice to do so.
				// The cookie must be the same for all nodes in the cluster.
				// The default value is "defaultcookie", but it is recommended to change it to a random value.
				// You can generate a random value using the command: openssl rand -base64 32
				// Then, set the value in the RABBITMQ_ERLANG_COOKIE environment variable.
				// More info: https://www.rabbitmq.com/clustering.html#erlang-cookie
				if rabbitCfg.ErlangCookie == "" {
					rabbitCookieKey, rabbitCookieErr := kbx.GetOrGenPasswordKeyringPass("rabbitmq-cookie")
					if rabbitCookieErr != nil {
						gl.Log("error", "Skipping RabbitMQ setup due to error generating password")
						gl.Log("debug", fmt.Sprintf("Error generating key: %v", rabbitCookieErr))
						goto postRabbit
					}
					rabbitCfg.ErlangCookie = string(rabbitCookieKey)
				}
				portBindings := []nat.PortMap{
					{
						nat.Port(fmt.Sprintf("%s/tcp", rabbitCfg.Port)):           []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: fmt.Sprintf("%v", rabbitCfg.Port)}},           // publica AMQP
						nat.Port(fmt.Sprintf("%s/tcp", rabbitCfg.ManagementPort)): []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: fmt.Sprintf("%v", rabbitCfg.ManagementPort)}}, // publica console
					},
				}

				if rabbitCfg.Vhost == "" {
					rabbitCfg.Vhost = "gobe"
				}

				dbConnObj := NewServices(
					"gdbase-rabbitmq",
					"rabbitmq:latest",
					[]string{
						"RABBITMQ_DEFAULT_USER=" + rabbitUser,
						"RABBITMQ_DEFAULT_PASS=" + rabbitPass,
						"RABBITMQ_DEFAULT_VHOST=" + rabbitCfg.Vhost,
						"RABBITMQ_PORT=" + rabbitCfg.Port.(string),
						"RABBITMQ_DB_NAME=" + rabbitCfg.Reference.Name,
						// "RABBITMQ_DB_VOLUME=" + rabbitCfg.Volume,
						"RABBITMQ_ERLANG_COOKIE=" + rabbitCfg.ErlangCookie,
						"RABBITMQ_PORT_5672_TCP_ADDR=" + rabbitCfg.Host,
						"RABBITMQ_PORT_5672_TCP_PORT=" + rabbitCfg.Port.(string),
						"RABBITMQ_PORT_15672_TCP_ADDR=" + rabbitCfg.Host,
						"RABBITMQ_PORT_15672_TCP_PORT=" + rabbitCfg.Port.(string),
					}, portBindings,
					map[string]struct{}{}, /* map[string]struct{}{
						fmt.Sprintf("%s:/var/lib/rabbitmq", rabbitCfg.Volume): {},
					}, */
				)
				services = append(services, dbConnObj)
			}
		}
	postRabbit:
		if config.Messagery.Redis != nil && config.Messagery.Redis.Enabled {
			if IsServiceRunning("gdbase-redis") {
				gl.Log("info", fmt.Sprintf("✅ %s já está rodando!\n", "gdbase-redis"))
			} else {
				if err := d.StartContainerByName("gdbase-redis"); err == nil {
					gl.Log("info", fmt.Sprintf("✅ %s já está rodando!\n", "gdbase-redis"))
				} else {
					rdsCfg := config.Messagery.Redis
					redisPass := rdsCfg.Password
					if redisPass == "" {
						redisPass = "guest"
					}
					// Create the volume for Redis, if exists definitions on the config
					// if rdsCfg.Volume == "" {
					// 	rdsCfg.Volume = os.ExpandEnv(glb.DefaultRedisVolume)
					// 	if err := d.CreateVolume("gdbase-redis-data", rdsCfg.Volume); err != nil {
					// 		return fmt.Errorf("❌ Erro ao criar volume do Redis: %v", err)
					// 	}
					// }
					// Create the Redis service
					servicesR := []*Services{
						NewServices("gdbase-redis", "redis:latest", []string{"REDIS_PASSWORD=" + redisPass}, []nat.PortMap{d.MapPorts("6379", "6379/tcp")}, nil),
					}
					// append the Redis service to the services slice
					services = append(services, servicesR...)
				}
			}
		}
	} else {
		gl.Log("debug", "Not found messagery in config, skipping RabbitMQ setup")
	}
	gl.Log("debug", fmt.Sprintf("Iniciando %d serviços...", len(services)))
	for _, srv := range services {
		mapPorts := map[nat.Port]struct{}{}
		for _, port := range srv.Ports {
			pt := ExtractPort(port)
			if pt == "" {
				gl.Log("error", fmt.Sprintf("❌ Erro ao mapear porta %s", pt))
				continue
			}
			if _, ok := pt.(map[string]string); !ok {
				gl.Log("error", fmt.Sprintf("❌ Erro ao mapear porta %s", pt))
				continue
			}
			// Verifica se a porta já está mapeada
			ptStr, ok := pt.(map[string]string)
			if !ok || ptStr["port"] == "" || ptStr["protocol"] == "" {
				gl.Log("error", fmt.Sprintf("❌ Erro ao mapear porta: tipo inválido ou campos ausentes: %v", pt))
				continue
			}
			portKey := nat.Port(fmt.Sprintf("%s/%s", ptStr["port"], ptStr["protocol"]))
			if _, exists := mapPorts[portKey]; exists {
				gl.Log("error", fmt.Sprintf("❌ Erro ao mapear porta %s", portKey))
				continue
			}
			// Adiciona a porta ao mapa
			portMap, ok := pt.(map[string]string)
			if !ok {
				gl.Log("error", fmt.Sprintf("❌ Erro ao converter porta: %v", pt))
				continue
			}
			portKey = nat.Port(fmt.Sprintf("%s/%s", portMap["port"], portMap["protocol"]))
			mapPorts[portKey] = struct{}{}
		}
		// Verifica se o serviço já está rodando
		// Isso já está dentro do StartContainer
		// if IsServiceRunning(srv.Name) {
		// 	gl.Log("info", fmt.Sprintf("✅ %s já está rodando!", srv.Name))
		// 	continue
		// }
		if err := d.StartContainer(srv.Name, srv.Image, srv.Env, mapPorts, nil); err != nil {
			return err
		}
	}
	return nil
}

func ExtractPort(port nat.PortMap) any {
	// Verifica se a porta é válida
	if port == nil {
		return nil
	}
	// Extrai a porta e o protocolo do primeiro elemento do map
	for k := range port {
		portStr := strings.Split(string(k), "/")
		if len(portStr) != 2 {
			return nil
		}
		portNum := portStr[0]
		protocol := portStr[1]
		return map[string]string{
			"port":     portNum,
			"protocol": protocol,
		}
	}
	return nil
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
		gl.Log("warn", "\nTempo esgotado. Docker não será ativado.")
		return false
	}
}

// Ativa o serviço do Docker dinamicamente conforme o sistema operacional
func startDockerService() error {
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
	if !isDockerRunning() {
		isInteractive := os.Getenv("IS_INTERACTIVE") == "true" || os.Getenv("IS_INTERACTIVE") == "1"
		if isInteractive {
			if AskToStartDocker() {
				gl.Log("info", "Starting Docker service...")
				if err := startDockerService(); err != nil {
					gl.Log("fatal", fmt.Sprintf("Error starting Docker: %v", err))
				}
				gl.Log("success", "Docker service started successfully!")
			} else {
				gl.Log("warn", "Docker service is not running and user chose not to start it.")
				gl.Log("warn", "Please start Docker manually to continue.")
				gl.Log("warn", "Exiting...")
				os.Exit(1)
			}
		} else {
			if err := startDockerService(); err != nil {
				gl.Log("fatal", fmt.Sprintf("Error starting Docker: %v", err))
			}
		}
	}
}
