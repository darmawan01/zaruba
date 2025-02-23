package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/state-alchemists/zaruba/config"
	"github.com/state-alchemists/zaruba/explainer"
	"github.com/state-alchemists/zaruba/inputer"
	"github.com/state-alchemists/zaruba/logger"
	"github.com/state-alchemists/zaruba/runner"
)

var pleaseEnv []string
var pleaseValues []string
var pleaseFile string
var pleaseInteractive *bool

// pleaseCmd represents the please command
var pleaseCmd = &cobra.Command{
	Use:     "please",
	Short:   "Ask Zaruba to do something for you",
	Long:    "💀 Ask Zaruba to do something for you",
	Aliases: []string{"run", "do", "invoke", "perform"},
	Run: func(cmd *cobra.Command, args []string) {
		project, taskNames, err := getProjectAndTaskNames(args)
		if err != nil {
			showErrorAndExit(err)
		}
		// no task provided
		if len(taskNames) == 0 {
			showDefaultResponse()
			return
		}
		// handle "please explain [taskNames...]"
		if taskNames[0] == "explain" {
			if err = project.Init(); err != nil {
				showErrorAndExit(err)
			}
			if err = explain(project, taskNames); err != nil {
				showErrorAndExit(err)
			}
			return
		}
		// handle "--interactive" flag
		if *pleaseInteractive {
			if err = inputer.Ask(project, taskNames); err != nil {
				showErrorAndExit(err)
			}
		}
		if err = project.Init(); err != nil {
			showErrorAndExit(err)
		}
		// handle "please explain [taskNames...]"
		r, err := runner.NewRunner(project, taskNames, time.Minute*5)
		if err != nil {
			showErrorAndExit(err)
		}
		if err := r.Run(); err != nil {
			showErrorAndExit(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(pleaseCmd)
	// get current working directory
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	// define defaultPleaseFile
	defaultPleaseFile := filepath.Join(dir, "main.zaruba.yaml")
	if _, err := os.Stat(defaultPleaseFile); os.IsNotExist(err) {
		defaultPleaseFile = "${ZARUBA_HOME}/scripts/core.zaruba.yaml"
	}
	// define defaultPleaseValues
	defaultPleaseValues := []string{}
	defaultValuesFile := filepath.Join(dir, "default.values.yaml")
	if _, err := os.Stat(defaultValuesFile); !os.IsNotExist(err) {
		defaultPleaseValues = append(defaultPleaseValues, defaultValuesFile)
	}
	// define defaultEnvFile
	defaultEnv := []string{}
	defaultEnvFile := filepath.Join(dir, ".env")
	if _, err := os.Stat(defaultEnvFile); !os.IsNotExist(err) {
		defaultEnv = append(defaultEnv, defaultEnvFile)
	}
	// register flags
	pleaseCmd.Flags().StringVarP(&pleaseFile, "file", "f", defaultPleaseFile, "task file")
	pleaseCmd.Flags().StringArrayVarP(&pleaseEnv, "environment", "e", defaultEnv, "environment file or pairs (e.g: '-e environment.env' or '-e key=val')")
	pleaseCmd.Flags().StringArrayVarP(&pleaseValues, "value", "v", defaultPleaseValues, "yaml file or pairs (e.g: '-v value.yaml' or '-v key=val')")
	pleaseInteractive = pleaseCmd.Flags().BoolP("interactive", "i", false, "if set, zaruba will ask you to fill inputs (e.g: -i)")
}

func showDefaultResponse() {
	d := logger.NewDecoration()
	logger.Printf("%sPlease what?%s\n", d.Bold, d.Normal)
	logger.Printf("Here are several things you can try:\n")
	logger.Printf("* %szaruba please explain task %s%s[task-keyword]%s\n", d.Yellow, d.Normal, d.Blue, d.Normal)
	logger.Printf("* %szaruba please explain input %s%s[input-keyword]%s\n", d.Yellow, d.Normal, d.Blue, d.Normal)
	logger.Printf("* %szaruba please explain %s%s[task-or-input-keyword]%s\n", d.Yellow, d.Normal, d.Blue, d.Normal)
}

func explain(project *config.Project, taskNames []string) (err error) {
	if len(taskNames) >= 2 {
		if taskNames[1] == "input" || taskNames[1] == "task" {
			keyword := strings.Join(taskNames[2:], " ")
			// handle "please explain input"
			if taskNames[1] == "input" {
				explainer.ExplainInputs(project, keyword)
				return nil
			}
			// handle "please explain task"
			explainer.ExplainTasks(project, keyword)
			return nil
		}
	}
	// handle "please explain"
	keyword := strings.Join(taskNames[1:], " ")
	explainer.ExplainTasks(project, keyword)
	explainer.ExplainInputs(project, keyword)
	return nil
}

func getProjectAndTaskNames(args []string) (project *config.Project, taskNames []string, err error) {
	taskNames = []string{}
	project, err = config.NewProject(pleaseFile)
	if err != nil {
		return project, taskNames, err
	}
	// process globalEnv
	for _, env := range pleaseEnv {
		if err = project.AddGlobalEnv(env); err != nil {
			return project, taskNames, err
		}
	}
	// process values from flag
	for _, value := range pleaseValues {
		if err = project.AddValue(value); err != nil {
			return project, taskNames, err
		}
	}
	//  distinguish taskNames and additional values
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			if err = project.AddValue(arg); err != nil {
				return project, taskNames, err
			}
			continue
		}
		_, argIsTask := project.Tasks[arg]
		if !argIsTask {
			if arg == "autostop" {
				if err = project.AddValue("autostop=true"); err != nil {
					return project, taskNames, err
				}
				continue
			}
		}
		taskNames = append(taskNames, arg)
	}
	return project, taskNames, err
}

func showErrorAndExit(err error) {
	d := logger.NewDecoration()
	if err != nil {
		logger.PrintfError("%s%s%s%s\n", d.Bold, d.Red, err.Error(), d.Normal)
		os.Exit(1)
	}
}
