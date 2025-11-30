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

func DockerCmd() *cobra.Command {
	var configFile string

	shortDesc := "Docker management commands for KubexDB"
	longDesc := "Docker management commands for KubexDB"

	cmd := &cobra.Command{
		Use:         "docker",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				return
			}
		},
	}
	cmd.Flags().StringVar(&configFile, "config-file", "config.yaml", "Path to configuration file")

	cmd.AddCommand(
		startDockerCmd(),
		stopDockerCmd(),
		statusDockerCmd(),
		restartDockerCmd(),
		restartDockerCmd(),
		getContainerLogs(),
		startContainerCmd(),
		createVolumeCmd(),
		getContainersListCmd(),
		getVolumesListCmd(),
		startContainerByNameCmd(),
		stopContainerByNameCmd(),
		addServiceCmd(),
	)
	return cmd
}

func startDockerCmd() *cobra.Command {

	shortDesc := "Start Docker"
	longDesc := "Start Docker service"

	cmd := &cobra.Command{
		Use:         "start",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			loggers := logz.GetLoggerZ("KubexDB")
			dkr, dkrErr := docker.NewDockerService(loggers)
			if dkrErr != nil {
				loggers.Log("error", "KubexDB", fmt.Sprintf("Error starting Docker service: %v", dkrErr))
				return
			}
			dkrErr = dkr.Initialize()
			if dkrErr != nil {
				loggers.Log("error", "KubexDB", fmt.Sprintf("Error initializing Docker service: %v", dkrErr))
				return
			}
			dkrErr = docker.SetupDatabaseServices(context.Background(), dkr, nil)
			if dkrErr != nil {
				loggers.Log("error", "KubexDB", fmt.Sprintf("Error setting up database services: %v", dkrErr))
				return
			}
		},
	}
	return cmd
}

func stopDockerCmd() *cobra.Command {

	shortDesc := "Stop Docker"
	longDesc := "Stop Docker service"

	cmd := &cobra.Command{
		Use:         "stop",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logz.GetLoggerZ("KubexDB")
			dkr, dkrErr := docker.NewDockerService(logger)
			if dkrErr != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error stopping Docker service: %v", dkrErr))
				return
			}
			dkrErr = dkr
			if dkrErr != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error stopping Docker service: %v", dkrErr))
				return
			}
			return
		},
	}
	return cmd
}

func statusDockerCmd() *cobra.Command {

	shortDesc := "Status Docker"
	longDesc := "Status Docker service"

	cmd := &cobra.Command{
		Use:         "status",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, logz.GetLogger("KubexDB"))
			//if dkrErr != nil {
			//	fmt.Printf("Error getting Docker status: %v\n", dkrErr)
			//	return
			//}
			//dkrErr = dkr.Status()
			//if dkrErr != nil {
			//	fmt.Printf("Error getting Docker status: %v\n", dkrErr)
			//	return
			//}
			_ = cmd.Help()
		},
	}
	return cmd
}

func restartDockerCmd() *cobra.Command {

	shortDesc := "Restart Docker"
	longDesc := "Restart Docker service"

	cmd := &cobra.Command{
		Use:         "restart",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logz.GetLoggerZ("KubexDB")
			dkr, dkrErr := docker.NewDockerService(logger)
			if dkrErr != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error restarting Docker service: %v", dkrErr))
				return
			}
			dkrErr = dkr.Initialize()
			if dkrErr != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error initializing Docker service: %v", dkrErr))
				return
			}
			dkrErr = docker.SetupDatabaseServices(context.Background(), dkr, nil)
			if dkrErr != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error setting up database services: %v", dkrErr))
				return
			}
		},
	}
	return cmd
}

// getContainerLogsCmd
func getContainerLogs() *cobra.Command {
	shortDesc := "Get Container Logs"
	longDesc := "Get logs from a specific Docker container"
	cmd := &cobra.Command{
		Use:         "logs",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logz.GetLoggerZ("KubexDB")
			dkr, dkrErr := docker.NewDockerService(logger)
			if dkrErr != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error getting container logs: %v\n", dkrErr))
				return
			}
			dkrErr = dkr.GetContainerLogs(context.Background(), "container_name", true)
			if dkrErr != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error getting container logs: %v\n", dkrErr))
				return
			}
			_ = cmd.Help()
		},
	}
	return cmd
}

// startContainerCmd
func startContainerCmd() *cobra.Command {
	shortDesc := "Start Container"
	longDesc := "Start a specific Docker container"
	var initArgs = &kbx.InitArgs{}

	cmd := &cobra.Command{
		Use:         "start",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logz.GetLoggerZ("KubexDB")
			dkr, dkrErr := docker.NewDockerService(logger)
			if dkrErr != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error starting container: %v\n", dkrErr))
				return
			}
			dkrErr = dkr.StartContainer(
				initArgs.Name,
				initArgs.Image,
				initArgs.EnvVars,
				nil, // initArgs.Ports,
				nil, //initArgs.Volumes,
			)
			if dkrErr != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error starting container: %v\n", dkrErr))
				return
			}
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "D", false, "Enable debug mode")

	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config", "c", "", "Path to config file")
	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to env file")
	cmd.Flags().StringVarP(&initArgs.Name, "name", "n", "", "Name of the container")
	cmd.Flags().StringVarP(&initArgs.Image, "image", "i", "", "Image of the container")

	cmd.Flags().StringToStringVarP(&initArgs.EnvVars, "env-vars", "E", nil, "Environment variables for the container")
	cmd.Flags().StringToStringVar(&initArgs.Ports, "ports", nil, "Port mappings for the container")
	cmd.Flags().StringToStringVar(&initArgs.Volumes, "volumes", nil, "Volume mappings for the container")

	return cmd
}

// createVolumeCmd
func createVolumeCmd() *cobra.Command {
	shortDesc := "Create Volume"
	longDesc := "Create a new Docker volume"

	cmd := &cobra.Command{
		Use:         "create",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, logz.GetLogger("KubexDB"))
			//if dkrErr != nil {
			//	fmt.Printf("Error creating volume: %v\n", dkrErr)
			//	return
			//}
			//dkrErr = dkr.CreateVolume("volume_name")
			//if dkrErr != nil {
			//	fmt.Printf("Error creating volume: %v\n", dkrErr)
			//	return
			//}
			_ = cmd.Help()
		},
	}
	return cmd
}

// getContainersListCmd
func getContainersListCmd() *cobra.Command {
	shortDesc := "Get Containers List"
	longDesc := "Get a list of all Docker containers"

	cmd := &cobra.Command{
		Use:         "list",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, logz.GetLogger("KubexDB"))
			//if dkrErr != nil {
			//	fmt.Printf("Error getting containers list: %v\n", dkrErr)
			//	return
			//}
			//dkrErr = dkr.GetContainersList()
			//if dkrErr != nil {
			//	fmt.Printf("Error getting containers list: %v\n", dkrErr)
			//	return
			//}
			_ = cmd.Help()
		},
	}
	return cmd
}

// getVolumesListCmd
func getVolumesListCmd() *cobra.Command {
	shortDesc := "Get Volumes List"
	longDesc := "Get a list of all Docker volumes"
	var showPath bool

	cmd := &cobra.Command{
		Use:         "list-volumes",
		Aliases:     []string{"list-vol", "volumes", "vol"},
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logz.GetLoggerZ("KubexDB")
			dkr, dkrErr := docker.NewDockerService(logger)
			if dkrErr != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error getting volumes list: %v", dkrErr))
				return
			}
			volList, dkrErr := dkr.GetVolumesList()
			if dkrErr != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error getting volumes list: %v", dkrErr))
				return
			}
			logger.Log("info", "Volumes list:")
			for _, volume := range volList {
				if showPath {
					logger.Log("info", fmt.Sprintf("    %s (Path: %s)", volume.Name, volume.Mountpoint))
					continue
				}
				logger.Log("info", fmt.Sprintf("    %s", volume.Name))
			}
		},
	}

	cmd.Flags().BoolVarP(&showPath, "show-path", "p", false, "Show the path of each volume")

	return cmd
}

// startContainerByNameCmd
func startContainerByNameCmd() *cobra.Command {
	shortDesc := "Start Container By Name"
	longDesc := "Start a specific Docker container by name"

	cmd := &cobra.Command{
		Use:         "start",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			logger := logz.GetLoggerZ("KubexDB")
			//dkr, dkrErr := factory.NewDockerService(nil, logger)
			//if dkrErr != nil {
			//	logger.Log("error", "KubexDB", fmt.Sprintf("Error starting container by name: %v", dkrErr))
			//	return
			//}
			//dkrErr = dkr.StartContainerByName("container_name")
			//if dkrErr != nil {
			//	logger.Log("error", "KubexDB", fmt.Sprintf("Error starting container by name: %v", dkrErr))
			//	return
			//}

			if err := cmd.Help(); err != nil {
				logger.Log("error", "KubexDB", fmt.Sprintf("Error displaying help: %v", err))
				return
			}
		},
	}
	return cmd
}

// stopContainerByNameCmd
func stopContainerByNameCmd() *cobra.Command {
	shortDesc := "Stop Container By Name"
	longDesc := "Stop a specific Docker container by name"

	cmd := &cobra.Command{
		Use:         "stop",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, logz.GetLogger("KubexDB"))
			//if dkrErr != nil {
			//	fmt.Printf("Error stopping container by name: %v\n", dkrErr)
			//	return
			//}
			//dkrErr = dkr.StopContainerByName("container_name")
			//if dkrErr != nil {
			//	fmt.Printf("Error stopping container by name: %v\n", dkrErr)
			//	return
			//}
			_ = cmd.Help()
		},
	}
	return cmd
}

// addServiceCmd
func addServiceCmd() *cobra.Command {
	shortDesc := "Add Service"
	longDesc := "Add a new service to Docker"

	cmd := &cobra.Command{
		Use:         "add",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, logz.GetLogger("KubexDB"))
			//if dkrErr != nil {
			//	fmt.Printf("Error adding service: %v\n", dkrErr)
			//	return
			//}
			//dkrErr = dkr.AddService("service_name")
			//if dkrErr != nil {
			//	fmt.Printf("Error adding service: %v\n", dkrErr)
			//	return
			//}
			_ = cmd.Help()
		},
	}
	return cmd
}
