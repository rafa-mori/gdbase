// Package module provides internal types and functions for the GoBE application.
package module

import (
	"github.com/kubex-ecosystem/gdbase/cmd/cli"
	"github.com/kubex-ecosystem/gdbase/internal/module/version"
	gl "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"

	"os"
	"strings"
)

type GDBase struct {
	parentCmdName string
	hideBanner    bool
	certPath      string
	keyPath       string
	configPath    string
}

func (m *GDBase) Alias() string {
	return ""
}
func (m *GDBase) ShortDescription() string {
	return "GDBase: GKBX Database and Docker manager/service. "
}
func (m *GDBase) LongDescription() string {
	return `GDBase: Is a tool to manage GKBX database and Docker services. It provides many DB flavors like MySQL, PostgreSQL, MongoDB, Redis, etc. It also provides Docker services like Docker Swarm, Docker Compose, etc. It is a command line tool that can be used to manage GKBX database and Docker services.`
}
func (m *GDBase) Usage() string {
	return "gdbase [command] [args]"
}
func (m *GDBase) Examples() []string {
	return []string{"gdbase [command] [args]", "gdbase database user auth'", "gdbase db roles list"}
}
func (m *GDBase) Active() bool {
	return true
}
func (m *GDBase) Module() string {
	return "gdbase"
}
func (m *GDBase) Execute() error {
	dbChanData := make(chan interface{})
	defer close(dbChanData)

	if spyderErr := m.Command().Execute(); spyderErr != nil {
		gl.Log("error", spyderErr.Error())
		return spyderErr
	} else {
		return nil
	}
}
func (m *GDBase) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: m.Module(),
		//Aliases:     []string{m.Alias(), "w", "wb", "webServer", "http"},
		Example: m.concatenateExamples(),
		Annotations: m.GetDescriptions(
			[]string{
				m.LongDescription(),
				m.ShortDescription(),
			}, m.hideBanner,
		),
		Version: version.GetVersion(),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(version.CliCommand())
	cmd.AddCommand(cli.DockerCmd())
	cmd.AddCommand(cli.DatabaseCmd())
	cmd.AddCommand(cli.UtilsCmds())
	cmd.AddCommand(cli.SSHCmds())

	setUsageDefinition(cmd)
	for _, c := range cmd.Commands() {
		setUsageDefinition(c)
		if !strings.Contains(strings.Join(os.Args, " "), c.Use) {
			if c.Short == "" {
				c.Short = c.Annotations["description"]
			}
		}
	}

	return cmd
}

func (m *GDBase) GetDescriptions(descriptionArg []string, hideBanner bool) map[string]string {
	return cli.GetDescriptions(descriptionArg, (m.hideBanner || hideBanner))
}
func (m *GDBase) SetParentCmdName(rtCmd string) {
	m.parentCmdName = rtCmd
}
func (m *GDBase) concatenateExamples() string {
	examples := ""
	rtCmd := m.parentCmdName
	if rtCmd != "" {
		rtCmd = rtCmd + " "
	}
	for _, example := range m.Examples() {
		examples += rtCmd + example + "\n  "
	}
	return examples
}
