package services

// import (
// 	"fmt"
// 	"os"

// 	"github.com/docker/go-connections/nat"
// 	krs "github.com/kubex-ecosystem/gdbase/internal/security/external"
// 	t "github.com/kubex-ecosystem/gdbase/internal/types"
// 	gl "github.com/kubex-ecosystem/logz"
// )

// func SetupRabbitMQ(config *t.RabbitMQ, dockerService IDockerService) error {
// 	if config == nil || !config.Enabled {
// 		gl.Log("debug", "RabbitMQ está desabilitado na configuração. Ignorando inicialização.")
// 		return nil
// 	}

// 	// Verifica se o serviço já está rodando
// 	if IsServiceRunning(config.Reference.Name) {
// 		gl.Log("info", fmt.Sprintf("✅ RabbitMQ (%s) já está rodando!", config.Reference.Name))
// 		return nil
// 	}

// 	if config.Username == "" {
// 		config.Username = "gobe"
// 	}
// 	if config.Password == "" {
// 		rabbitPassKey, rabbitPassErr := krs.GetOrGenerateFromKeyring("rabbitmq")
// 		if rabbitPassErr != nil {
// 			gl.Log("error", "Skipping RabbitMQ setup due to error generating password")
// 			gl.Log("debug", fmt.Sprintf("Error generating key: %v", rabbitPassErr))
// 			return rabbitPassErr
// 		} else {
// 			config.Password = string(rabbitPassKey)
// 		}
// 	}
// 	if config.Vhost == "" {
// 		config.Vhost = "gobe"
// 	}
// 	if config.Port == nil || config.Port == "" {
// 		config.Port = "5672"
// 	}
// 	if config.ManagementPort == "" {
// 		config.ManagementPort = "15672"
// 	}
// 	if config.ErlangCookie == "" {
// 		rabbitCookieKey, rabbitCookieErr := krs.GetOrGenerateFromKeyring("rabbitmq-cookie")
// 		if rabbitCookieErr != nil {
// 			gl.Log("error", "Skipping RabbitMQ setup due to error generating password")
// 			gl.Log("debug", fmt.Sprintf("Error generating key: %v", rabbitCookieErr))
// 			return rabbitCookieErr
// 		} else {
// 			config.ErlangCookie = string(rabbitCookieKey)
// 		}
// 	}
// 	if config.Volume == "" {
// 		config.Volume = os.ExpandEnv(DefaultRabbitMQVolume)
// 	}

// 	// Cria o volume, se necessário
// 	if err := dockerService.CreateVolume(config.Reference.Name, config.Volume); err != nil {
// 		gl.Log("error", fmt.Sprintf("❌ Erro ao criar volume do RabbitMQ: %v", err))
// 		return err
// 	}

// 	// Mapeia as portas

// 	portBindings := []nat.PortMap{
// 		{
// 			nat.Port(fmt.Sprintf("%s/tcp", config.Port)):           []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: fmt.Sprintf("%v", config.Port)}},           // publica AMQP
// 			nat.Port(fmt.Sprintf("%s/tcp", config.ManagementPort)): []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: fmt.Sprintf("%v", config.ManagementPort)}}, // publica console
// 		},
// 	}

// 	// Configura as variáveis de ambiente
// 	envVars := []string{
// 		"RABBITMQ_DEFAULT_USER=" + config.Username,
// 		"RABBITMQ_DEFAULT_PASS=" + config.Password,
// 		"RABBITMQ_DEFAULT_VHOST=" + config.Vhost,
// 		"RABBITMQ_ERLANG_COOKIE=" + config.ErlangCookie,
// 	}

// 	// Inicializa o container do RabbitMQ
// 	service := dockerService.AddService(
// 		config.Reference.Name,
// 		"rabbitmq:latest",
// 		envVars,
// 		portBindings,
// 		map[string]struct{}{
// 			fmt.Sprintf("%s:/var/lib/rabbitmq", config.Volume): {},
// 		},
// 	)
// 	if service == nil {
// 		err := fmt.Errorf("serviço não encontrado: %s", config.Reference.Name)
// 		gl.Log("error", fmt.Sprintf("❌ Erro ao iniciar o RabbitMQ: %v", err))
// 		return err
// 	}

// 	gl.Log("success", fmt.Sprintf("✅ RabbitMQ (%s) iniciado com sucesso!", config.Reference.Name))
// 	return nil
// }
