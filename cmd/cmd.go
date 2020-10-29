/*
Copyright © 2020 Jody Scott <jody@thescottsweb.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/jodydadescott/tokenmachine/config"
	"github.com/jodydadescott/tokenmachine/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "tokenmachine",
	Short: "get kerberos keytabs with oauth tokens",
	Long: `Provides expiring kerberos keytabs to holders of bearer tokens by validating token is
 permitted keytab by policy. Policy is in the form of Open Policy Agent (OPA). Keytabs
 may be used to generate kerberos tickets and then discarded.
 
 See https://github.com/jodydadescott/keytab-token-broker for more details.
	`,
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "manage service",
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "install service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return installService()
	},
}

var serviceRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeService()
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "start service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startService()
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return stopService()
	},
}

var servicePauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "pause service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pauseService()
	},
}

var serviceContinueCmd = &cobra.Command{
	Use:   "continue",
	Short: "continue service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return continueService()
	},
}

var serviceConfigSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set configuration",

	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) < 1 {
			return errors.New("config string required")
		}

		runtimeConfigString := args[0]
		// Need to verify string

		err := SetRuntimeConfigString(runtimeConfigString)
		if err != nil {
			return err
		}
		return nil
	},
}

var serviceConfigShowCmd = &cobra.Command{
	Use:   "show",
	Short: "show config",

	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := GetRuntimeConfigString()
		if err != nil {
			return err
		}
		fmt.Println(config)
		return nil
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "make config or config example",
}

var configExampleCmd = &cobra.Command{
	Use:   "example",
	Short: "example configuration",
	RunE: func(cmd *cobra.Command, args []string) error {

		newConfig := config.NewV1ExampleConfig()

		configString := ""
		switch strings.ToLower(viper.GetString("format")) {

		case "", "yaml":
			configString = newConfig.YAML()
			break

		case "json":
			configString = newConfig.JSON()
			break

		default:
			return fmt.Errorf(fmt.Sprintf("Output format %s is unknown. Must be yaml or json", viper.GetString("format")))
		}

		fmt.Print(configString)
		return nil

	},
}

var configMakeCmd = &cobra.Command{
	Use:   "make",
	Short: "make configuration",

	RunE: func(cmd *cobra.Command, args []string) error {

		configLoader := internal.NewLoader()

		// Input config can be zero, one or many
		if viper.GetString("config") != "" {
			for _, s := range strings.Split(viper.GetString("config"), ",") {
				err := configLoader.LoadFrom(s)
				if err != nil {
					return err
				}
			}
		}

		configString := ""
		switch strings.ToLower(viper.GetString("format")) {

		case "", "yaml":
			configString = configLoader.Config.YAML()
			break

		case "json":
			configString = configLoader.Config.JSON()
			break

		default:
			return fmt.Errorf(fmt.Sprintf("Output format %s is unknown. Must be yaml or json", viper.GetString("format")))
		}

		fmt.Print(configString)
		return nil

	},
}

var windowsRunDebugCmd = &cobra.Command{
	Use:   "run-debug",
	Short: "run debug (non service)",

	RunE: func(cmd *cobra.Command, args []string) error {

		var err error
		configLoader := internal.NewLoader()

		if viper.GetString("config") == "" {

			config, err := GetRuntimeConfigString()
			if err != nil {
				return err
			}

			err = configLoader.LoadFrom(config)

		} else {
			err = configLoader.LoadFrom(viper.GetString("config"))
		}

		if err != nil {
			return err
		}

		// Override debug level
		configLoader.Config.Logging.LogLevel = "debug"

		serverConfig, err := configLoader.ServerConfig()
		if err != nil {
			return err
		}

		zapConfig, err := configLoader.ZapConfig()
		if err != nil {
			return err
		}

		logger, err := zapConfig.Build()
		if err != nil {
			return err
		}

		zap.ReplaceGlobals(logger)
		//defer logger.Sync()

		sig := make(chan os.Signal, 2)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		server, err := serverConfig.Build()
		if err != nil {
			return err
		}

		zap.L().Debug("Started successfully")
		<-sig

		zap.L().Debug("Shutting down on signal")
		server.Shutdown()

		return nil
	},
}

var serverCmd = &cobra.Command{
	Use:   "start",
	Short: "start server",

	RunE: func(cmd *cobra.Command, args []string) error {

		var err error
		configLoader := internal.NewLoader()

		if viper.GetString("config") == "" {

			config, err := GetRuntimeConfigString()
			if err != nil {
				return err
			}

			err = configLoader.LoadFrom(config)

		} else {
			err = configLoader.LoadFrom(viper.GetString("config"))
		}

		if err != nil {
			return err
		}

		serverConfig, err := configLoader.ServerConfig()
		if err != nil {
			return err
		}

		zapConfig, err := configLoader.ZapConfig()
		if err != nil {
			return err
		}

		logger, err := zapConfig.Build()
		if err != nil {
			return err
		}

		zap.ReplaceGlobals(logger)
		//defer logger.Sync()

		sig := make(chan os.Signal, 2)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		server, err := serverConfig.Build()
		if err != nil {
			return err
		}

		zap.L().Debug("Started successfully")
		<-sig

		zap.L().Debug("Shutting down on signal")
		server.Shutdown()

		return nil
	},
}

// Execute ...
func Execute() {

	if runtime.GOOS == "windows" {
		isIntSess, err := isAnInteractiveSession()
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("failed to determine if we are running in an interactive session: %v", err))
			os.Exit(1)
		}
		if !isIntSess {
			runService()
			return
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// serviceCmd

func init() {

	if runtime.GOOS == "windows" {

		serviceCmd.AddCommand(serviceInstallCmd, serviceRemoveCmd, serviceStartCmd, serviceStopCmd, servicePauseCmd, serviceContinueCmd, serviceConfigSetCmd, serviceConfigShowCmd)
		configCmd.AddCommand(configExampleCmd, configMakeCmd)
		rootCmd.AddCommand(serviceCmd, configCmd, windowsRunDebugCmd)

	} else {

		configCmd.AddCommand(configMakeCmd, configExampleCmd)
		rootCmd.AddCommand(configCmd, serverCmd)

	}

	// Server

	rootCmd.PersistentFlags().StringP("config", "", "", "configuration file")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	// Config
	rootCmd.PersistentFlags().StringP("format", "", "", "output format in yaml or json; default is yaml")
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))

}
