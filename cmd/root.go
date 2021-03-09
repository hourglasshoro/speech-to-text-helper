/*
Copyright © 2021 hourglasshoro

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
	"fmt"
	"github.com/hourglasshoro/speech-to-text-helper/pkg"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "speech-to-text-helper",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		currentDir, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "cannot get current dir")
		}

		source := cmd.Flag("source").Value.String()
		fileDir := Solve(source, currentDir)

		output := cmd.Flag("output").Value.String()
		outputDir := Solve(output, currentDir)

		overwrite, err := strconv.ParseBool(cmd.Flag("overwrite").Value.String())
		if err != nil {
			return err
		}

		parallel, err := strconv.Atoi(cmd.Flag("parallel").Value.String())
		if err != nil {
			return err
		}

		apiKey, serviceUrl, err := pkg.LoadEnv()
		if err != nil {
			return err
		}

		err = pkg.Send(apiKey, serviceUrl, fileDir, outputDir, overwrite, parallel)
		if err != nil {
			return err
		}

		return nil
	},
}

// Solve resolves the path of the root directory to be searched from the input source and currentDir
func Solve(source string, currentDir string) (outputDir string) {
	if source != "" && !filepath.IsAbs(source) {
		outputDir = path.Join(currentDir, source)
	} else if filepath.IsAbs(source) {
		outputDir = source
	} else {
		outputDir = currentDir
	}
	return
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.speech-to-text-helper.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringP("source", "s", "", "Directory to search")
	rootCmd.Flags().StringP("output", "o", "", "Directory to output")
	rootCmd.Flags().BoolP("overwrite", "w", false, "Overwrite an already existing file or")
	rootCmd.Flags().IntP("parallel", "p", 100, "Number of goroutine")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".speech-to-text-helper" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".speech-to-text-helper")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
