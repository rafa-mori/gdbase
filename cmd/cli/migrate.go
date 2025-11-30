package cli

import (
	"context"
	"os"

	"github.com/kubex-ecosystem/gdbase/internal/engine"
	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/kubex-ecosystem/gdbase/internal/services/docker"

	dockerStack "github.com/kubex-ecosystem/gdbase/internal/backends/dockerstack"
	logz "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"
)

func migrateDatabaseCmd() *cobra.Command {
	var initArgs = &kbx.InitArgs{}
	var keepAlive bool

	shortDesc := "Run database migrations"
	longDesc := "Run database migrations for all registered models."

	cmd := &cobra.Command{
		Use:         "migrate",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize context and logger
			ctx := context.Background()
			logger := logz.GetLoggerZ("Migration")
			if initArgs.Debug {
				logger.SetDebugMode(true)
			}

			// ========== STEP 1: LOAD CONFIG ==========
			logz.Log("info", "📋 Loading configuration...")
			rootConfig, err := engine.LoadRootConfig(
				kbx.GetValueOrDefaultSimple(initArgs.ConfigFile, os.ExpandEnv(kbx.DefaultConfigFile)),
			)
			if err != nil {
				logz.Log("error", "Failed to load config:", err.Error())
				return
			}

			// ========== STEP 2: CREATE SERVICE MANAGER ==========
			logz.Log("info", "🔧 Initializing Docker service...")
			dockerService, err := docker.NewDockerService(logger)
			if err != nil {
				logz.Log("error", "Failed to create Docker service:", err.Error())
				return
			}

			// ========== STEP 3: CREATE PROVIDER WITH INJECTION ==========
			logz.Log("info", "🏗️  Initializing DockerStack provider...")
			dsp := dockerStack.NewDockerStackProvider(dockerService)

			// ========== STEP 4-6: PROVIDER ORCHESTRATES EVERYTHING ==========
			logz.Log("info", "🚀 Starting migration pipeline...")
			if err := dsp.StartServices(ctx, rootConfig); err != nil {
				logz.Log("error", "Migration pipeline failed:", err.Error())
				return
			}

			// ========== STEP 7 (OPTIONAL): ENGINE CONNECTIONS ==========
			if keepAlive {
				logz.Log("info", "🔌 Establishing engine connections (keep-alive mode)...")
				mgr := engine.NewDatabaseManager(logger)
				if err := mgr.InitFromRootConfig(ctx, rootConfig); err != nil {
					logz.Log("error", "Failed to initialize engine:", err.Error())
					return
				}
				logz.Log("success", "Engine ready for runtime operations")
				// Note: In keep-alive mode, connections remain open.
				// Add graceful shutdown handling if needed.
			}

			logz.Log("success", "✅ Migration pipeline completed successfully!")
		},
	}

	// Flags
	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")
	cmd.Flags().BoolVarP(&keepAlive, "keep-alive", "k", false, "Keep engine connections alive after migration (default: false)")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config-file", "C", "config.yaml", "Path to configuration file")

	// Future flags (not yet implemented)
	cmd.Flags().BoolVarP(&initArgs.Force, "force", "f", false, "Force apply all migrations (not yet implemented)")
	cmd.Flags().BoolVarP(&initArgs.Reset, "reset", "r", false, "Reset database before migrations (not yet implemented)")
	cmd.Flags().BoolVarP(&initArgs.DryRun, "dry-run", "", false, "Simulate migrations without applying (not yet implemented)")

	return cmd
}
