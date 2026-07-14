package domain

import "context"

type IndexRun struct {
	Status         string
	DocumentsCount int
	ChunksCount    int
	FinishedAt     string
}

type Stats struct {
	SourcesCount    int
	DocumentsCount  int
	ChunksCount     int
	SearchLogsCount int
	LastIndexRun    IndexRun
}

type StatsRepository interface {
	GetGeneralStats(ctx context.Context) (*Stats, error)
}

type ChunkRepository interface {
	GetChunksCount(ctx context.Context) (int, error)
}

type StatsUseCase interface {
	GetStats(ctx context.Context) (*Stats, error)
}