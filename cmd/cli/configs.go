package cli

import (
	"fmt"
	"os"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ConfigCmd() *cobra.Command {
	var initArgs = &kbx.InitArgs{}
	// var configFile string
	shortDesc := "Edit configuration"
	longDesc := "Edit configuration file interactively"

	cmd := &cobra.Command{
		Use:         "config",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: GetDescriptions([]string{shortDesc, longDesc}, (os.Getenv("KUBEXDS_HIDEBANNER") == "true")),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := ReadConfig(initArgs.ConfigFile)
			if err != nil {
				return err
			}

			updatedConfig, err := EditConfig(config)
			if err != nil {
				return err
			}

			if err := SaveConfig(initArgs.ConfigFile, updatedConfig); err != nil {
				return err
			}

			fmt.Println("Configuration updated successfully")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")

	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVar(&initArgs.ConfigFile, "config-file", "config.yaml", "Path to configuration file")

	return cmd
}

func ReadConfig(configFile string) (map[string]interface{}, error) {
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config map[string]interface{}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}

	return config, nil
}

func EditConfig(config map[string]interface{}) (map[string]interface{}, error) {
	// Definimos o Título do formulário
	//var title = "Editando Configuração"
	//
	//// Aqui vamos criar um novo mapa de configuração, que será o mapa que será retornado
	//var updatedConfig = make(map[string]string)
	//
	//// Aqui vamos criar as instâncias vazias das estruturas que compõem o formulário
	//// FormFields e FormField
	//var formFields = x.FormFields{}
	//var formField = x.FormField{}
	//
	//// O FormConfig é a estrutura que serve de fundação para o formulário.
	//
	//// Aqui vamos iterar sobre as chaves do mapa de configuração, criando nas iterações as estruturas
	//// FormConfig, FormFields e FormField, que são as estruturas que representam todos os formulários e campos
	//// que podem ser criados pelo xtui. Após a criação das instâncias, vamos chamar a função ShowForm, que é
	//// responsável por exibir o formulário na tela e retornar os valores preenchidos pelo usuário.
	//// As instâncias serão criadas TODAS de uma vêz. As estruturas permitem campos aninhados, então a criação
	//// de formulários e campos é feita de forma recursiva e uma só vez.

	return nil, fmt.Errorf("error reading config file: %v", "not implemented")

}

func SaveConfig(configFile string, config map[string]interface{}) error {
	for key, value := range config {
		viper.Set(key, value)
	}

	return viper.WriteConfigAs(configFile)
}
