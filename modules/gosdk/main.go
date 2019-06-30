package main

import (
	"fmt"
	"hfrd/modules/gosdk/chaincode"
	"hfrd/modules/gosdk/channel"
	"hfrd/modules/gosdk/common"
	"hfrd/modules/gosdk/execute"
	"hfrd/modules/gosdk/metadata"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:               "gosdk",
	PersistentPreRunE: preRunE,
}

func preRunE(cmd *cobra.Command, args []string) error {
	// check connection profile and certs root directory
	cmdFlags := cmd.Flags()
	flagNames := []string{common.CONN_PROFILE, common.NAME, common.ITERATION_COUNT, common.ITERATION_INTERVAL}
	for _, flagName := range flagNames {
		if f := cmdFlags.Lookup(flagName); f == nil || f.Value.String() == "" {
			return fmt.Errorf("%s flag required", flagName)
		} else if flagName == common.CONN_PROFILE {
			if !strings.Contains(f.Value.String(), "/") {
				f.Value.Set("/fabric/keyfiles/" + f.Value.String() + "/connection.yml")
				fmt.Printf("connectionProfile: %s \n", f.Value.String())
			}
		}
	}
	common.InitLog(viper.GetString(common.LOG_LEVEL))
	common.SetRoutineLimit(viper.GetInt(common.CONCURRENCY_LIMIT))
	return nil
}

func main() {
	rootCmd.Version = metadata.GetVersion()
	rootFlags := rootCmd.PersistentFlags()
	rootFlags.StringP(common.CONN_PROFILE, "c", "./fixtures/ConnectionProfile_org1.yaml", "connection profile file")
	rootFlags.String(common.NAME, "default test name", "test name")
	rootFlags.String(common.ITERATION_COUNT, "1",
		"test iteration count/time. e.g. \"5\" for 5 iterations; \"1h5m2s\" for 1 hour + 5 minutes + 2 seconds loop")
	rootFlags.String(common.ITERATION_INTERVAL, "1s", "wait time for next iteration")
	rootFlags.Int(common.RETRY_COUNT, 1, "test retry count")
	rootFlags.String(common.LOG_LEVEL, "ERROR", "logging level")
	rootFlags.Int(common.CONCURRENCY_LIMIT, 500, "Max number of goroutines to send request")
	rootFlags.String(common.IGNORE_ERRORS, "false", "Ignore errors to make tests continue")
	rootFlags.String(common.DELAY_TIME, "",
		"test delay time. e.g. \"1h5m2s\" for 1 hour + 5 minutes + 2 seconds")

	viper.BindPFlag(common.CONN_PROFILE, rootFlags.Lookup(common.CONN_PROFILE))
	viper.BindPFlag(common.NAME, rootFlags.Lookup(common.NAME))
	viper.BindPFlag(common.ITERATION_COUNT, rootFlags.Lookup(common.ITERATION_COUNT))
	viper.BindPFlag(common.ITERATION_INTERVAL, rootFlags.Lookup(common.ITERATION_INTERVAL))
	viper.BindPFlag(common.RETRY_COUNT, rootFlags.Lookup(common.RETRY_COUNT))
	viper.BindPFlag(common.LOG_LEVEL, rootFlags.Lookup(common.LOG_LEVEL))
	viper.BindPFlag(common.CONCURRENCY_LIMIT, rootFlags.Lookup(common.CONCURRENCY_LIMIT))
	viper.BindPFlag(common.IGNORE_ERRORS, rootFlags.Lookup(common.IGNORE_ERRORS))
	viper.BindPFlag(common.DELAY_TIME, rootFlags.Lookup(common.DELAY_TIME))

	rootCmd.AddCommand(channel.Cmd())
	rootCmd.AddCommand(chaincode.Cmd())
	rootCmd.AddCommand(execute.Cmd())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
