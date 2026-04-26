package commands

import (
	"fmt"
	"log"

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
	Run: func(cmd *cobra.Command, args []string) {
		f := bridge.NewFactory(DBInstance)
		prompt, err := f.GenerateInstruction(args[0])
		if err != nil {
			log.Fatalf("❌ Failed to generate instruction: %v", err)
		}
		fmt.Println(prompt)
	},
}
