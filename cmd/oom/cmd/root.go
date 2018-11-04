package cmd

import (
	"fmt"
	"jcheng/bexp/app"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var useScorch bool
var dataDir string
var idxDir string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "oom",
	Short: "short desc",
	Long: `long desc`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if dataDir == "" {
			fmt.Println("dataDir is required")
			os.Exit(1)
		}
		if idxDir == "" {
			fmt.Println("idxDir is required")
			os.Exit(1)
		}

		err := app.OOMIndex(useScorch, dataDir, idxDir)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bexp.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVarP(&useScorch, "useScorch", "s", false, "use scorch or not")
	rootCmd.Flags().StringVarP(&dataDir, "dataDir", "d", "", "dataDir")
	rootCmd.Flags().StringVarP(&idxDir, "idxDir", "i", "", "idxDir")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".bexp" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".bexp")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
