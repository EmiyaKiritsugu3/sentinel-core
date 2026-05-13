package commands

import (
	"fmt"
	"log"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/liveview"
	"github.com/EmiyaKiritsugu3/sentinel-core/internal/registry"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	registry.Register(NewLiveCmd)
}

// NewLiveCmd creates a cobra command that starts the Sentinel Live View
// WebSocket server and bootstraps a background project scan.
func NewLiveCmd(db *sqlite.DB) *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "live",
		Short: "Start the Sentinel Live View WebSocket server",
	}

	if err := sqlite.ValidateDB(db, "live-cmd"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return err }
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// 1. Instantiates LiveView Server
		server := liveview.NewServer()

		errChan := make(chan error, 1)
		go func() {
			errChan <- server.Run(cmd.Context())
		}()

		// 2. Registers Server as an Observer to the Engine
		if err := graph.Migrate(cmd.Context(), db); err != nil {
			return fmt.Errorf("live: migration failed: %w", err)
		}

		engine, err := graph.NewEngine(db)
		if err != nil {
			return fmt.Errorf("live: failed to create engine: %w", err)
		}
		engine.RegisterObserver(server)

		// Start a background scan (for demonstration/bootstrapping)
		fmt.Println("🚀 Sentinel: Bootstrapping initial background scan for Live View...")
		go func() {
			err := engine.ScanProject(cmd.Context(), ".")
			if err != nil {
				log.Printf("liveview: background scan error: %v\n", err)
			}
		}()

		// 3. Starts the HTTP server
		go func() {
			errChan <- server.StartHTTP(port, db)
		}()

		select {
		case err := <-errChan:
			if err != nil {
				return err
			}
		case <-cmd.Context().Done():
			return cmd.Context().Err()
		}
		return nil
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for the Live View server")
	return cmd
}
