package execute

import (
	"fmt"
	"hfrd/modules/gosdk/common"
	"hfrd/modules/gosdk/utilities"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"bytes"
	"syscall"
	"time"

	"strings"

	"github.com/pkg/errors"
)

var executeCommandCmd = &cobra.Command{
	Use:              "command",
	Short:            "execute bash script or binary",
	TraverseChildren: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeCommand()
	},
}

const EXECUTE_CMD = "execute.command"

func cmCmd() *cobra.Command {
	flags := executeCommandCmd.Flags()
	flags.StringVar(&commandName, COMMAND_NAME, "", "The command that will be executed.Can be a path of script or binary")
	flags.StringVar(&commandParams, COMMAND_PARAMS, "", "The parameters that script or binary will use")
	executeCommandCmd.MarkFlagRequired(COMMAND_NAME)
	return executeCommandCmd
}

func executeCommand() error {
	common.Delay(viper.GetString(common.DELAY_TIME))
	base := common.NewBase()
	base.IterationCount = viper.GetString(common.ITERATION_COUNT)
	base.SetIterationInterval(viper.GetString(common.ITERATION_INTERVAL))
	base.RetryCount = viper.GetInt(common.RETRY_COUNT)
	defer base.PrintMetrics(EXECUTE_CMD)
	// Split parameters by #
	commandParamArray := strings.Split(commandParams, "#")
	common.Logger.Debug(fmt.Sprintf("commandParams: %s", commandParamArray))
	commandFunc := func(iterationIndex int) error {
		// Generate required command parameters
		commandArgs, err := utilities.GetComplexArgs(commandParamArray, iterationIndex)
		if err != nil {
			return err
		}
		return ExecuteCommand(commandArgs)
	}

	_, _, err := common.IterateFunc(base, commandFunc, true)
	if err != nil {
		return err
	}
	return nil
}

func ExecuteCommand(commandArgs []string) error {
	defer common.TrackTime(time.Now(), EXECUTE_CMD)
	cmd := exec.Command(commandName, commandArgs...)
	var outbuf, errbuf bytes.Buffer
	var exitCode int
	const defaultFailedCode = 1
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout := outbuf.String()
	stderr := errbuf.String()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			common.Logger.Error(fmt.Sprintf("Could not get exit code for failed program: %v, %v \n", commandName, commandArgs))
			exitCode = defaultFailedCode
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}
	common.Logger.Info(fmt.Sprintf("Execute command result, stdout: %v, stderr: %v, exitCode: %v \n", stdout, stderr, exitCode))
	if exitCode != 0 {
		return errors.Errorf("Execute command failed with exitCode %v \n", exitCode)
	}
	return nil
}
