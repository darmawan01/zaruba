package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/state-alchemists/zaruba/action"
)

func init() {
	rootCmd.AddCommand(doCmd)
}

var doCmd = &cobra.Command{
	Use:   "do <action> [...args]",
	Short: "Do custom action",
	Long:  `Zaruba will perform custom action`,
	Run: func(cmd *cobra.Command, args []string) {
		// handle invalid parameter
		if len(args) < 1 {
			log.Fatal("[ERROR] action expected, current arguments: ", args)
		}
		// get `action`, `arguments` and `projectDir`
		actionString := args[0]
		arguments := args[1:]
		projectDir := "."
		// invoke action
		log.Printf("[INFO] Invoking %s. project-dir: %s, other arguments: %#v", actionString, projectDir, arguments)
		if err := action.Do(actionString, projectDir, arguments...); err != nil {
			log.Fatal("[ERROR] ", err)
		}
	},
}