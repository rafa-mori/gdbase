package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/go-connections/nat"
	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"
	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	logz "github.com/kubex-ecosystem/logz"

	krg "github.com/kubex-ecosystem/gdbase/internal/security/external"
)

func SetupDatabaseServices(ctx context.Context, d ci.IDockerService, rootConfig *kbx.RootConfig) error {
	if rootConfig == nil {
		return fmt.Errorf("❌ Configuração do banco de dados não encontrada")
	}
	if !kbx.DefaultTrue(rootConfig.Enabled) {
		logz.Log("debug", "Database services are disabled in config, skipping DB setup")
		return nil
	}
	if len(rootConfig.Databases) == 0 {
		logz.Log("debug", "Not found connections in config, skipping DB setup")
		return nil
	}
	var services = make([]*ci.Services, 0)

	// START GENERIC DATABASE CONFIGS
	if len(rootConfig.Databases) > 0 {
		for _, dbConfig := range rootConfig.Databases {
			if dbConfig != nil {
				if (dbConfig.Type == "postgres" || dbConfig.Type == "postgresql") && kbx.DefaultTrue(dbConfig.Enabled) {
					// Check if the database is already running
					if IsServiceRunning("kubexdb-pg") {
						logz.Log("debug", fmt.Sprintf("✅ %s já está rodando!", "kubexdb-pg"))
						continue
					} else {
						if err := d.StartContainerByName("kubexdb-pg"); err == nil {
							logz.Log("debug", fmt.Sprintf("✅ %s já está rodando!", "kubexdb-pg"))
							continue
						} else {
							// Check if Password is empty, if so, try to retrieve it from keyring
							// if not found, generate a new one
							if dbConfig.Pass == "" {
								pgPassKey, pgPassErr := krg.GetOrGenerateFromKeyring("pgpass")
								if pgPassErr != nil {
									logz.Log("error", fmt.Sprintf("Error generating key: %v", pgPassErr))
									continue
								}
								dbConfig.Pass = string(pgPassKey)
							} else {
								logz.Log("debug", fmt.Sprintf("Password found in config: %s", dbConfig.Pass[0:2]))
							}

							var vol string
							if volume, ok := dbConfig.Options["volume"]; ok {
								vol = volume.(string)
							}

							if len(vol) == 0 {
								vol = os.ExpandEnv(kbx.DefaultPostgresVolume)
							}

							pgVolRootDir := os.ExpandEnv(vol)
							pgVolInitDir := filepath.Join(pgVolRootDir, "init")
							vols := map[string]struct{}{
								strings.Join([]string{pgVolInitDir, "/docker-entrypoint-initdb.d"}, ":"): {},
							}
							// initDBSQLs, initDBSQLErr := embed.FS.ReadDir(initDBSQLFiles, "embedded")
							// if initDBSQLErr != nil {
							// 	logz.Log("error", fmt.Sprintf("❌ Erro ao ler diretório de scripts SQL: %v", initDBSQLErr))
							// 	continue
							// } else {
							// 	for _, initDBSQL := range initDBSQLs {
							// 		initDBSQLData, initDBSQLErr := embed.FS.ReadFile(initDBSQLFiles, filepath.Join("embedded", initDBSQL.Name()))
							// 		if initDBSQLErr != nil {
							// 			logz.Log("error", fmt.Sprintf("❌ Erro ao ler script SQL %s: %v", initDBSQL.Name(), initDBSQLErr))
							// 			continue
							// 		}
							// 		if _, err := WriteInitDBSQL(pgVolInitDir, initDBSQL.Name(), string(initDBSQLData)); err != nil {
							// 			logz.Log("error", fmt.Sprintf("❌ Erro ao criar diretório do PostgreSQL: %v", err))
							// 			continue
							// 		}
							// 	}
							// 	if err := d.CreateVolume("kubexdb-pg-init", pgVolInitDir); err != nil {
							// 		logz.Log("error", fmt.Sprintf("❌ Erro ao criar volume do PostgreSQL: %v", err))
							// 		continue
							// 	}
							// 	pgVolDataDir := filepath.Join(pgVolRootDir, "pgdata")
							// 	if err := d.CreateVolume("kubexdb-pg-data", pgVolDataDir); err != nil {
							// 		logz.Log("error", fmt.Sprintf("❌ Erro ao criar volume do PostgreSQL: %v", err))
							// 		continue
							// 	}
							// 	vols[strings.Join([]string{pgVolDataDir, "/var/lib/postgresql/data"}, ":")] = struct{}{}
							// }
							// Check if the port is already in use and find an available one if necessary
							if (dbConfig.Port == "") || len(dbConfig.Port) == 0 {
								dbConfig.Port = "5432"
							}
							port, err := svc.FindAvailablePort(5432, 10)
							if err != nil {
								logz.Log("error", fmt.Sprintf("❌ Erro ao encontrar porta disponível: %v", err))
								continue
							}
							dbConfig.Port = port
							// Map the port to the container
							portMap := d.MapPorts(dbConfig.Port, "5432/tcp")

							// Check if the database name is empty, if so, generate a random one
							if dbConfig.Name == "" {
								dbConfig.Name = "godo-" + svc.RandStringBytes(5)
							}
							enableTLS := func(dbConfig *kbx.DBConfig) string {
								if dbConfig.TLSEnabled {
									return "require"
								}
								return "disable"
							}

							// Insert the PostgreSQL service into the services slice
							dbConnObj := NewServices(
								"kubexdb-pg",
								"postgres:17-alpine",
								map[string]string{
									// "POSTGRES_HOST_AUTH_METHOD=trust", // Use only for development, not recommended for production
									"POSTGRES_HOST_AUTH_METHOD": "trust",
									// Necessary for Postgres 12+
									"POSTGRES_INITDB_ARGS":    "--encoding=UTF8 --locale=pt_BR.UTF-8 --data-checksums",
									"POSTGRES_USER":           dbConfig.User,
									"POSTGRES_PASSWORD":       dbConfig.Pass,
									"POSTGRES_DB":             dbConfig.Name,
									"POSTGRES_PORT":           dbConfig.Port,
									"POSTGRES_DB_NAME":        dbConfig.Name,
									"POSTGRES_DB_VOLUME":      vol,
									"POSTGRES_DB_SSLMODE":     enableTLS(dbConfig),
									"POSTGRES_DB_INITDB_ARGS": "--data-checksums",
									// Necessary for some clients
									"PGUSER":     dbConfig.User,
									"PGPASSWORD": dbConfig.Pass,
									"PGDATABASE": dbConfig.Name,
									"PGPORT":     dbConfig.Port,
									"PGHOST":     dbConfig.Host,
									"PGDATA":     "/var/lib/postgresql/data/pgdata",
									"PGSSLMODE":  enableTLS(dbConfig),
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
		logz.Log("debug", "Not found databases in config, skipping DB setup")
	}

	logz.Log("debug", fmt.Sprintf("Iniciando %d serviços...", len(services)))
	for _, srv := range services {
		mapPorts := map[nat.Port]struct{}{}
		for _, port := range srv.Ports {
			pt := svc.ExtractPort(port)
			if pt == "" {
				logz.Log("error", fmt.Sprintf("❌ Erro ao mapear porta %s", pt))
				continue
			}
			if _, ok := pt.(map[string]string); !ok {
				logz.Log("error", fmt.Sprintf("❌ Erro ao mapear porta %s", pt))
				continue
			}
			// Verifica se a porta já está mapeada
			ptStr, ok := pt.(map[string]string)
			if !ok || ptStr["port"] == "" || ptStr["protocol"] == "" {
				logz.Log("error", fmt.Sprintf("❌ Erro ao mapear porta: tipo inválido ou campos ausentes: %v", pt))
				continue
			}
			portKey := nat.Port(fmt.Sprintf("%s/%s", ptStr["port"], ptStr["protocol"]))
			if _, exists := mapPorts[portKey]; exists {
				logz.Log("error", fmt.Sprintf("❌ Erro ao mapear porta %s", portKey))
				continue
			}
			// Adiciona a porta ao mapa
			portMap, ok := pt.(map[string]string)
			if !ok {
				logz.Log("error", fmt.Sprintf("❌ Erro ao converter porta: %v", pt))
				continue
			}
			portKey = nat.Port(fmt.Sprintf("%s/%s", portMap["port"], portMap["protocol"]))
			mapPorts[portKey] = struct{}{}
		}
		// Verifica se o serviço já está rodando
		// Isso já está dentro do StartContainer
		// if IsServiceRunning(srv.Name) {
		// 	logz.Log("info", fmt.Sprintf("✅ %s já está rodando!", srv.Name))
		// 	continue
		// }
		if err := d.StartContainer(srv.Name, srv.Image, srv.Env, mapPorts, nil); err != nil {
			return err
		}
	}
	return nil
}
