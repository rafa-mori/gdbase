// Package cli contém comandos relacionados à linha de comando.
package cli

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kubex-ecosystem/gdbase/utils"
	"github.com/spf13/cobra"
)

// SSHCmds retorna uma lista de comandos Cobra relacionados a SSH.
// Retorna um slice de ponteiros para comandos Cobra.
func SSHCmds() *cobra.Command {
	shortDesc := "Configura os utilitários SSH do sistema"
	longDesc := "Configura os utilitários SSH do sistema"

	rootCmd := &cobra.Command{
		Use:         "ssh",
		Aliases:     []string{"s", "ss"},
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
	}

	rootCmd.AddCommand(sshTunnelCmd())
	rootCmd.AddCommand(sshTunnelServiceCmd())

	return rootCmd
}

// sshTunnelCmd cria um comando Cobra para configurar um túnel SSH.
// Retorna um ponteiro para o comando Cobra configurado.
func sshTunnelCmd() *cobra.Command {
	var sshUser, sshCert, sshPassword, sshAddress, sshPort string
	var tunnels []string
	var background bool

	shortDesc := "Configura um túnel SSH"
	longDesc := "Configura um túnel SSH"

	rootCmd := &cobra.Command{
		Use:         "tunnel",
		Aliases:     []string{"tun", "t"},
		Short:       shortDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		RunE: func(cmd *cobra.Command, args []string) error {
			if background {
				sshCmdRun := exec.Command("kbx", "u", "s", "tunnel-service-background", "--sshUser", sshUser, "--sshCert", sshCert, "--sshPassword", sshPassword, "--sshAddress", sshAddress, "--sshPort", sshPort, "--tunnels", strings.Join(tunnels, ","))
				sshCmdRunErr := sshCmdRun.Start()
				if sshCmdRunErr != nil {
					log.Println("Erro ao iniciar o serviço de túnel SSH:", sshCmdRunErr)
					return nil
				}
				//processReleaseErr := sshCmdRun.Process.Release()
				//if processReleaseErr != nil {
				//	log.Println("Erro ao liberar o processo do serviço de túnel SSH:", processReleaseErr)
				//	return nil
				//}
				log.Println("Serviço de túnel SSH iniciado em segundo plano")
				return nil
			}

			var sshCred utils.SSHCred
			sshCred.User = sshUser
			sshCred.PrivateKey = []byte(sshCert)
			sshCred.Password = sshPassword
			timeout := 10 * time.Second

			tnl, err := utils.SSHConnect(sshAddress, sshCred, timeout)
			if err != nil {
				log.Println("Erro ao conectar via SSH:", err)
				return nil
			}
			defer tnl.Close()

			return nil
		},
	}

	rootCmd.Flags().BoolVarP(&background, "background", "b", false, "Executar em segundo plano")
	rootCmd.Flags().StringVarP(&sshUser, "login", "l", "", "Usuário SSH")
	rootCmd.Flags().StringVarP(&sshCert, "cert", "i", "", "Certificado SSH")
	rootCmd.Flags().StringVarP(&sshPassword, "secret", "s", "", "Senha SSH")
	rootCmd.Flags().StringVarP(&sshAddress, "host", "t", "", "Endereço SSH")
	rootCmd.Flags().StringVarP(&sshPort, "port", "p", "", "Porta SSH")
	rootCmd.Flags().StringSliceVarP(&tunnels, "tunnels", "L", []string{}, "Túneis")

	return rootCmd
}

// sshTunnelServiceCmd cria um comando Cobra para configurar um serviço de túnel SSH em segundo plano.
// Retorna um ponteiro para o comando Cobra configurado.
func sshTunnelServiceCmd() *cobra.Command {
	var sshUser, sshCert, sshPassword, sshAddress, sshPort string
	var tunnels []string
	rootCmd := &cobra.Command{
		Use:    "tunnel-service-background",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			var sshCred utils.SSHCred
			sshCred.User = sshUser
			sshCred.PrivateKey = []byte(sshCert)
			sshCred.Password = sshPassword
			timeout := 10 * time.Second

			tnl, err := utils.SSHConnect(sshAddress, sshCred, timeout)
			if err != nil {
				log.Println("Erro ao conectar via SSH:", err)
				return
			}
			defer tnl.Close()

		},
	}
	rootCmd.Flags().StringVarP(&sshUser, "sshUser", "l", "", "Usuário SSH")
	rootCmd.Flags().StringVarP(&sshCert, "sshCert", "i", "", "Certificado SSH")
	rootCmd.Flags().StringVarP(&sshPassword, "sshPassword", "s", "", "Senha SSH")
	rootCmd.Flags().StringVarP(&sshAddress, "sshAddress", "t", "", "Endereço SSH")
	rootCmd.Flags().StringVarP(&sshPort, "sshPort", "p", "", "Porta SSH")
	rootCmd.Flags().StringSliceVarP(&tunnels, "tunnels", "L", []string{}, "Túneis")
	return rootCmd
}

// TODO: FIX USAGE

// cred := utils.SSHCred{
//   User: "ubuntu",
//   // Password: "xxx",                      // ou
//   PrivateKey: os.ReadFile("~/.ssh/id_rsa"), // o que preferir
// }

// t, err := utils.SshConnect("meu-vps:22", cred, 5*time.Second)
// if err != nil { panic(err) }
// defer t.Close()

// // 1) Local forward: expõe localmente 15432 → conecta no remoto 127.0.0.1:5432 (via SSH)
// lf := "L:127.0.0.1:15432->127.0.0.1:5432"

// // 2) Remote forward: expõe remotamente 0.0.0.0:8080 → conecta no local 127.0.0.1:8080
// rf := "R:0.0.0.0:8080->127.0.0.1:8080"

// sp1, _ := utils.ParseForwardSpec(lf)
// sp2, _ := utils.ParseForwardSpec(rf)

// stopAll, err := t.Start(sp1, sp2)
// if err != nil { panic(err) }
// defer stopAll()

// select {} // fica de pé
