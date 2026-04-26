package commands

import (
	"fmt"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(instructCmd)
}

var instructCmd = &cobra.Command{
	Use:   "instruct [task_id]",
	Short: "Generate the sovereign instruction prompt for an AI agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := bridge.NewFactory(DBInstance)
		prompt, err := f.GenerateInstruction(args[0])
		if err != nil {
			return fmt.Errorf("instruct: failed to generate instruction for %s: %w", args[0], err)
		}
		fmt.Println(prompt)
		return nil
	},
}
