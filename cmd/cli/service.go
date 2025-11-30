package cli

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kubex-ecosystem/gdbase/internal/bootstrap"
	info "github.com/kubex-ecosystem/gdbase/internal/module/info"
	kbx "github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/spf13/cobra"
)

var (
	// configLoader holds the embedded service configuration files.
	// configLoader, configErr = bootstrap.MigrationFiles.ReadDir("services")
	configLoader, configErr []os.DirEntry

	// infoModule holds the manifest information of the current module.
	infoModule, infoErr = info.GetManifest()

	// systemdTemplate is the embedded template for systemd service unit files.
	systemdTemplate string

	// wrapperTemplate is the embedded template for the service wrapper script.
	wrapperTemplate string

	// ps1Template is the embedded template for Windows service installation script.
	ps1Template string
)

// -------------------------------------------------------------------
// UNIVERSAL SERVICE MANAGER FOR ANY MODULE
// -------------------------------------------------------------------

func NewServiceCommand() *cobra.Command {
	var moduleArgs *kbx.InitArgs

	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage OS service for this Kubex module",
	}

	cmd.AddCommand(cmdInstall(moduleArgs))
	cmd.AddCommand(cmdUninstall(moduleArgs))
	cmd.AddCommand(cmdStart(moduleArgs))
	cmd.AddCommand(cmdStop(moduleArgs))
	cmd.AddCommand(cmdStatus(moduleArgs))

	return cmd
}

func systemctlCmd() string {
	if os.Geteuid() == 0 {
		return "systemctl"
	}
	return "systemctl --user"
}

// -------------------------------------------------------------------
// INSTALL
// -------------------------------------------------------------------

func cmdInstall(initArgs *kbx.InitArgs) *cobra.Command {
	if initArgs == nil {
		initArgs = &kbx.InitArgs{}
	}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install module as OS service",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch runtime.GOOS {
			case "linux":
				return installLinux(initArgs)
			case "windows":
				return installWindows(initArgs)
			default:
				return fmt.Errorf("OS não suportado")
			}
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Habilitar modo debug para o serviço")
	cmd.Flags().BoolVarP(&initArgs.FailFast, "fail-fast", "f", false, "Habilitar modo fail-fast para o serviço")
	cmd.Flags().BoolVarP(&initArgs.BatchMode, "batch-mode", "b", false, "Habilitar modo batch para o serviço")
	cmd.Flags().BoolVarP(&initArgs.NoColor, "no-color", "", false, "Desabilitar cores no log")
	cmd.Flags().BoolVarP(&initArgs.RootMode, "root-mode", "", false, "Habilitar modo root para o serviço")

	cmd.Flags().StringVarP(&initArgs.Name, "name", "", "localhost", "Nome do serviço")
	cmd.Flags().StringVarP(&initArgs.Host, "host", "", "localhost", "Host para o serviço escutar")
	cmd.Flags().StringVarP(&initArgs.Command, "command", "", "localhost", "Comando para o serviço executar")
	cmd.Flags().StringVarP(&initArgs.Subcommand, "subcommand", "", "localhost", "Subcomando para o serviço executar")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config", "c", os.ExpandEnv(kbx.DefaultKubexDSConfigPath), "Caminho para o arquivo de configuração")
	cmd.Flags().StringVarP(&initArgs.EnvFile, "env", "e", "", "Caminho para o arquivo .env")
	cmd.Flags().StringVarP(&initArgs.LogFile, "log", "l", "", "Caminho para o arquivo de log")

	cmd.Flags().IntVarP(&initArgs.MaxProcs, "max-procs", "", 0, "Número máximo de processos")
	cmd.Flags().IntVarP(&initArgs.TimeoutMS, "timeout-ms", "", 0, "Timeout em milissegundos para operações")

	cmd.Flags().StringToStringVarP(&initArgs.EnvVars, "option", "o", nil, "Variável de ambiente adicional no formato CHAVE=VALOR (aceita múltiplas)")

	return cmd
}

func installLinux(moduleArgs *kbx.InitArgs) error {
	name := moduleArgs.Name
	defaultCmd := moduleArgs.Command
	binPath := infoModule.GetBin()
	defaultConfig := moduleArgs.ConfigFile

	isRoot := os.Geteuid() == 0 || moduleArgs.RootMode
	home, _ := os.UserHomeDir()

	var unitPath, wrapperPath, serviceUser string

	if isRoot {
		unitPath = fmt.Sprintf("/etc/systemd/system/%s.service", name)
		wrapperPath = "/usr/local/bin/kubex-svc"
		serviceUser = "appuser"
	} else {
		unitPath = fmt.Sprintf("%s/.config/systemd/user/%s.service", home, name)
		wrapperPath = fmt.Sprintf("%s/.local/bin/kubex-svc", home)
		serviceUser = os.Getenv("USER")
		os.MkdirAll(filepath.Dir(unitPath), 0755)
		os.MkdirAll(filepath.Dir(wrapperPath), 0755)
	}

	// prepare wrapper content
	wrapperContent := wrapperTemplate
	wrapperContent = strings.ReplaceAll(wrapperContent, "{{MODULE_BIN}}", binPath)
	wrapperContent = strings.ReplaceAll(wrapperContent, "{{MODULE_CONFIG}}", defaultConfig)

	var flags []string
	if moduleArgs.Debug {
		flags = append(flags, "--debug")
	}
	if moduleArgs.FailFast {
		flags = append(flags, "--fail-fast")
	}
	if moduleArgs.BatchMode {
		flags = append(flags, "--batch-mode")
	}
	if moduleArgs.NoColor {
		flags = append(flags, "--no-color")
	}
	if moduleArgs.RootMode {
		flags = append(flags, "--root-mode")
	}
	if moduleArgs.LogFile != "" {
		flags = append(flags, fmt.Sprintf("--log=%s", moduleArgs.LogFile))
	}
	if moduleArgs.EnvFile != "" {
		flags = append(flags, fmt.Sprintf("--env=%s", moduleArgs.EnvFile))
	}
	if moduleArgs.MaxProcs > 0 {
		flags = append(flags, fmt.Sprintf("--max-procs=%d", moduleArgs.MaxProcs))
	}
	if moduleArgs.TimeoutMS > 0 {
		flags = append(flags, fmt.Sprintf("--timeout-ms=%d", moduleArgs.TimeoutMS))
	}
	for k, v := range moduleArgs.EnvVars {
		flags = append(flags, fmt.Sprintf("--option=%s=%s", k, v))
	}

	if moduleArgs.Subcommand != "" {
		flags = append(flags, moduleArgs.Subcommand)
	}

	if defaultCmd != "" {
		flags = append(flags, defaultCmd)
	}

	wrapperContent = strings.ReplaceAll(wrapperContent, "{{MODULE_FLAGS}}", strings.Join(flags, " "))
	// write wrapper
	os.WriteFile(wrapperPath, []byte(wrapperContent), 0755)

	// compile systemd template
	unit := systemdTemplate

	unit = strings.ReplaceAll(unit, "{{MODULE_NAME}}", name)
	unit = strings.ReplaceAll(unit, "{{MODULE_BIN}}", binPath)
	unit = strings.ReplaceAll(unit, "{{MODULE_CONFIG}}", defaultConfig)
	unit = strings.ReplaceAll(unit, "{{MODULE_DEFAULT_CMD}}", defaultCmd)
	unit = strings.ReplaceAll(unit, "{{WRAPPER_PATH}}", wrapperPath)
	unit = strings.ReplaceAll(unit, "{{SERVICE_USER}}", serviceUser)

	// write unit
	os.WriteFile(unitPath, []byte(unit), 0644)

	// reload + enable
	exec.Command("sh", "-c", systemctlCmd()+" daemon-reload").Run()
	exec.Command("sh", "-c", systemctlCmd()+" enable "+name).Run()

	fmt.Println("✓ Serviço Kubex instalado:", name)
	return nil
}

// -------------------------------------------------------------------
// WINDOWS
// -------------------------------------------------------------------

func installWindows(args *kbx.InitArgs) error {
	name := args.Name
	binPath := infoModule.GetBin()
	config := args.ConfigFile

	os.MkdirAll("C:\\kubex", 0755)

	path := fmt.Sprintf("C:\\kubex\\install-%s.ps1", name)
	content := strings.ReplaceAll(ps1Template, "$ModuleName", name)
	content = strings.ReplaceAll(content, "$BinaryPath", binPath)
	content = strings.ReplaceAll(content, "$ConfigPath", config)

	os.WriteFile(path, []byte(content), 0644)

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// -------------------------------------------------------------------
// START/STOP/STATUS
// -------------------------------------------------------------------

func cmdStart(moduleArgs *kbx.InitArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return exec.Command("sh", "-c", systemctlCmd()+" start "+moduleArgs.Name).Run()
		},
	}
}

func cmdStop(moduleArgs *kbx.InitArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return exec.Command("sh", "-c", systemctlCmd()+" stop "+moduleArgs.Name).Run()
		},
	}
}

func cmdStatus(moduleArgs *kbx.InitArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show service status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return exec.Command("sh", "-c", systemctlCmd()+" status "+moduleArgs.Name).Run()
		},
	}
}

func cmdUninstall(moduleArgs *kbx.InitArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove service",
		RunE: func(cmd *cobra.Command, args []string) error {

			exec.Command("sh", "-c", systemctlCmd()+" stop "+moduleArgs.Name).Run()
			exec.Command("sh", "-c", systemctlCmd()+" disable "+moduleArgs.Name).Run()

			os.Remove("/etc/systemd/system/" + moduleArgs.Name + ".service")
			os.Remove(os.Getenv("HOME") + "/.config/systemd/user/" + moduleArgs.Name + ".service")

			fmt.Println("✓ Serviço removido:", moduleArgs.Name)
			return nil
		},
	}
}

func init() {
	// if configErr != nil {
	// 	panic(configErr)
	// }
	if infoErr != nil {
		panic(infoErr)
	}

	if configLoader != nil {
		return
	}

	// load templates
	for _, file := range configLoader {
		if file.IsDir() {
			continue
		}
		data, err := bootstrap.MigrationFiles.ReadFile("services/" + file.Name())
		if err != nil {
			panic(err)
		}
		switch file.Name() {
		case "systemd.service.tpl":
			systemdTemplate = string(data)
		case "wrapper.sh.tpl":
			wrapperTemplate = string(data)
		case "install.ps1.tpl":
			ps1Template = string(data)
		}
	}
}
