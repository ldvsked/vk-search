package domain

import "context"

type HealthRepository interface {
	Ping(ctx context.Context) error
}

type HealthUseCase interface {
	CheckHealth(ctx context.Context) error
}