package cli

import (
	"os"

	"github.com/kubex-ecosystem/gdbase/utils/helpers"
	"github.com/spf13/cobra"
)

// UtilsCmds retorna uma lista de comandos Cobra relacionados a utilitários do sistema.
// Retorna um slice de ponteiros para comandos Cobra.
func UtilsCmds() *cobra.Command {

	shortDesc := "Configura os utilitários do sistema"
	longDesc := "Configura os utilitários do sistema"

	uCmd := &cobra.Command{
		Use:         "utils",
		Aliases:     []string{"u", "util"},
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
	}
	uCmd.AddCommand(installUtilsCmd())
	uCmd.AddCommand(uninstallUtilsCmd())
	return uCmd
}

// sshTunnelServiceCmd cria um comando Cobra para configurar um serviço de túnel SSH em segundo plano.
// Retorna um ponteiro para o comando Cobra configurado.
func installUtilsCmd() *cobra.Command {

	shortDesc := "Install the bash helpers."
	longDesc := "Install the bash helpers."

	rootCmd := &cobra.Command{
		Use:         "install",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			helpers.InstallBashHelpers()
		},
	}
	return rootCmd
}

// sshTunnelCmd cria um comando Cobra para configurar um túnel SSH.
// Retorna um ponteiro para o comando Cobra configurado.
func uninstallUtilsCmd() *cobra.Command {

	shortDesc := "Uninstall the bash helpers."
	longDesc := "Uninstall the bash helpers."

	rootCmd := &cobra.Command{
		Use:         "uninstall",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			helpers.UninstallBashHelpers()
		},
	}

	return rootCmd
}
