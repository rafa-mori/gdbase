package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/kubex-ecosystem/gdbase/internal/services/docker"
	logz "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"
)

func DatabaseCmd() *cobra.Command {
	var initArgs = &kbx.InitArgs{}

	shortDesc := "Database management commands for KubexDB"
	longDesc := "Database management commands for KubexDB"
	cmd := &cobra.Command{
		Use:         "database",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logz.GetLoggerZ("KubexDB")
			if initArgs.Debug {
				logger.SetDebugMode(true)
			}
			logger.Log("info", "KubexDB", "Database management commands for KubexDB")
			if err := cmd.Help(); err != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error displaying help: %v", err))
				return
			}
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")
	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVar(&initArgs.ConfigFile, "config-file", "config.yaml", "Path to configuration file")

	cmd.AddCommand(startDatabaseCmd())
	cmd.AddCommand(stopDatabaseCmd())
	cmd.AddCommand(statusDatabaseCmd())
	cmd.AddCommand(migrateDatabaseCmd())

	return cmd
}

func startDatabaseCmd() *cobra.Command {
	var initArgs = &kbx.InitArgs{}

	shortDesc := "Start Database services"
	longDesc := "Start Database services"

	cmd := &cobra.Command{
		Use:         "start",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logz.GetLoggerZ("KubexDB")
			if initArgs.Debug {
				logger.SetDebugMode(true)
			}

			if len(initArgs.ConfigFile) == 0 {
				// initArgs = config.NewInitArgs()
			}
			if initArgs.ConfigFile == "" {
				initArgs.ConfigFile = kbx.DefaultConfigFile
			}
			// cfg, err := config.NewDatabaseConfigMapWithFile(initArgs.ConfigFile)
			// if err != nil {
			// 	logz.Log("error", "KubexDB", fmt.Sprintf("Error loading configuration: %v", err))
			// 	return
			// }

			// D.1. Start Docker Service and DB containers
			logger.Log("info", "KubexDB", "Creating docker service...")
			dockerService, err := docker.NewDockerService(logger)
			if err != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error creating docker service: %v", err))
				return
			}

			// D.2. Setup Database Services
			logger.Log("info", "KubexDB", "Setting up database services...")
			err = docker.SetupDatabaseServices(context.Background(), dockerService, nil)
			if err != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error setting up database services: %v", err))
				return
			}

			// D.3. Create DB Service
			logger.Log("info", "KubexDB", "Creating database service...")
			// dbService, err := docker.NewDatabaseServiceImpl(context.Background(), nil, logger)
			// if err != nil {
			// 	logger.Log("error", "KubexDB", fmt.Sprintf("Error creating db service: %v", err))
			// 	return
			// }

			// D.4. Initialize DB Service
			logger.Log("info", "KubexDB", "Initializing database service...")
			// if err := dbService.Initialize(context.Background()); err != nil {
			// 	logger.Log("error", "KubexDB", fmt.Sprintf("Error initializing db service: %v", err))
			// 	return
			// }

			logger.Log("success", "KubexDB", "Database services started successfully.")
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")

	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVar(&initArgs.ConfigFile, "config-file", "config.yaml", "Path to configuration file")

	return cmd
}

func stopDatabaseCmd() *cobra.Command {
	var initArgs = &kbx.InitArgs{}

	shortDesc := "Stop Docker"
	longDesc := "Stop Docker service"

	cmd := &cobra.Command{
		Use:         "stop",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")

	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVar(&initArgs.ConfigFile, "config-file", "config.yaml", "Path to configuration file")

	return cmd
}

func statusDatabaseCmd() *cobra.Command {
	var initArgs = &kbx.InitArgs{}

	shortDesc := "Status Docker"
	longDesc := "Status Docker service"

	cmd := &cobra.Command{
		Use:         "status",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")

	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVar(&initArgs.ConfigFile, "config-file", "config.yaml", "Path to configuration file")

	return cmd
}
