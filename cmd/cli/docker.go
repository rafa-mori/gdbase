package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/kubex-ecosystem/gdbase/factory"
	s "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"
)

func DockerCmd() *cobra.Command {
	var configFile string

	shortDesc := "Docker management commands for GDBase"
	longDesc := "Docker management commands for GDBase"

	cmd := &cobra.Command{
		Use:         "docker",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
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
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			dkr, dkrErr := factory.NewDockerService(nil, gl.GetLogger("GDBase"))
			if dkrErr != nil {
				fmt.Printf("Error starting Docker service: %v\n", dkrErr)
				return
			}
			dkrErr = dkr.Initialize()
			if dkrErr != nil {
				fmt.Printf("Error initializing Docker service: %v\n", dkrErr)
				return
			}
			dkrErr = s.SetupDatabaseServices(context.Background(), dkr, nil)
			if dkrErr != nil {
				gl.Log("error", fmt.Sprintf("Error setting up database services: %v", dkrErr))
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
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, l.GetLogger("GDBase"))
			//if dkrErr != nil {
			//	fmt.Printf("Error stopping Docker service: %v\n", dkrErr)
			//	return
			//}
			//dkrErr = dkr
			//if dkrErr != nil {
			//	fmt.Printf("Error stopping Docker service: %v\n", dkrErr)
			//	return
			//}
			// return
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
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, l.GetLogger("GDBase"))
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
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			dkr, dkrErr := factory.NewDockerService(nil, gl.GetLogger("GDBase"))
			if dkrErr != nil {
				fmt.Printf("Error restarting Docker service: %v\n", dkrErr)
				return
			}
			dkrErr = dkr.Initialize()
			if dkrErr != nil {
				fmt.Printf("Error initializing Docker service: %v\n", dkrErr)
				return
			}
			dkrErr = s.SetupDatabaseServices(context.Background(), dkr, nil)
			if dkrErr != nil {
				gl.Log("error", fmt.Sprintf("Error setting up database services: %v", dkrErr))
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
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, l.GetLogger("GDBase"))
			//if dkrErr != nil {
			//	fmt.Printf("Error getting container logs: %v\n", dkrErr)
			//	return
			//}
			//dkrErr = dkr.GetContainerLogs("container_name")
			//if dkrErr != nil {
			//	fmt.Printf("Error getting container logs: %v\n", dkrErr)
			//	return
			//}
			_ = cmd.Help()
		},
	}
	return cmd
}

// startContainerCmd
func startContainerCmd() *cobra.Command {
	shortDesc := "Start Container"
	longDesc := "Start a specific Docker container"

	cmd := &cobra.Command{
		Use:         "start",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, l.GetLogger("GDBase"))
			//if dkrErr != nil {
			//	fmt.Printf("Error starting container: %v\n", dkrErr)
			//	return
			//}
			//dkrErr = dkr.StartContainer("container_name")
			//if dkrErr != nil {
			//	fmt.Printf("Error starting container: %v\n", dkrErr)
			//	return
			//}
			_ = cmd.Help()
		},
	}
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
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, l.GetLogger("GDBase"))
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
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, l.GetLogger("GDBase"))
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
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			dkr, dkrErr := factory.NewDockerService(nil, gl.GetLogger("GDBase"))
			if dkrErr != nil {
				fmt.Printf("Error getting volumes list: %v\n", dkrErr)
				return
			}
			volList, dkrErr := dkr.GetVolumesList()
			if dkrErr != nil {
				fmt.Printf("Error getting volumes list: %v\n", dkrErr)
				return
			}
			gl.Log("info", "Volumes list:")
			for _, volume := range volList {
				if showPath {
					gl.Log("info", fmt.Sprintf("    %s (Path: %s)", volume.Name, volume.Mountpoint))
					continue
				}
				gl.Log("info", fmt.Sprintf("    %s", volume.Name))
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
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, l.GetLogger("GDBase"))
			//if dkrErr != nil {
			//	fmt.Printf("Error starting container by name: %v\n", dkrErr)
			//	return
			//}
			//dkrErr = dkr.StartContainerByName("container_name")
			//if dkrErr != nil {
			//	fmt.Printf("Error starting container by name: %v\n", dkrErr)
			//	return
			//}
			_ = cmd.Help()
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
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, l.GetLogger("GDBase"))
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
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("GDBASE_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			//dkr, dkrErr := factory.NewDockerService(nil, l.GetLogger("GDBase"))
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
