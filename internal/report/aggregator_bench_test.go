package report

import (
	"context"
	"testing"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

func BenchmarkFetchStats(b *testing.B) {
	db, err := sqlite.InitAtPath(":memory:")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	// create dummy tables
	_, err = db.Conn.Exec(`
		CREATE TABLE nodes (id TEXT, type TEXT);
		CREATE TABLE tasks (id TEXT, status TEXT, math_delta REAL, tier TEXT, description TEXT, created_at DATETIME, verification_command TEXT);
	`)
	if err != nil {
		b.Fatal(err)
	}

	agg, err := NewAggregator(db)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := agg.FetchStats(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
