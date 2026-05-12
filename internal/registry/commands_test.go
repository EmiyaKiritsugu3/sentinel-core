package registry

import (
	"sync"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_ConcurrentSafety(t *testing.T) {
	ResetForTesting()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Register(func(_ *sqlite.DB) *cobra.Command {
				return &cobra.Command{Use: "test"}
			})
		}()
	}
	wg.Wait()

	cmds := GetCommands()
	assert.Equal(t, 100, len(cmds))
}

func TestGetCommands_ReturnsDefensiveCopy(t *testing.T) {
	ResetForTesting()

	Register(func(_ *sqlite.DB) *cobra.Command {
		return &cobra.Command{Use: "original"}
	})
	require.Len(t, GetCommands(), 1)

	returned := GetCommands()
	returned[0] = nil

	original := GetCommands()
	assert.NotNil(t, original[0], "modificar cópia não deve afetar original")
}

func TestGetCommands_EmptyRegistry(t *testing.T) {
	ResetForTesting()

	cmds := GetCommands()
	assert.Empty(t, cmds)
}
