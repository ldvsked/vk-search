package di

import (
	"context"
	"net/http"

	"go.uber.org/fx"
	"vk-search/internal/api"
	"vk-search/internal/api/handlers"
	"vk-search/internal/app/auth"
	"vk-search/internal/app/search"
	"vk-search/internal/infrastructure/config"
	"vk-search/internal/infrastructure/mocks"
	"vk-search/internal/infrastructure/postgres"
)

func BuildApp() *fx.App {
	return fx.New(
		fx.Provide(
			// 1. Конфигурация
		config.Load,
		fx.Annotate(
			func(cfg *config.Config) auth.TokenConfig { return cfg },
		),

		// 2. Репозитории (Инфраструктурный слой)
		postgres.NewPgxPool,
		postgres.NewUserRepository,
		mocks.NewSearchMockRepository, 

		// 3. Юзкейсы (Бизнес-логика / Слой приложения)
		auth.NewAuthUseCase,
		search.NewSearchUseCase, 

		// 4. Хендлеры и Маршрутизация (Транспортный слой)
		handlers.NewAuthHandler,
		handlers.NewSearchHandler, 
		api.NewRouter,
		),
		fx.Invoke(func(lc fx.Lifecycle, handler http.Handler, cfg *config.Config) {
			srv := &http.Server{
				Addr:    ":" + cfg.GetHTTPPort(),
				Handler: handler,
			}
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go srv.ListenAndServe()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return srv.Shutdown(ctx)
				},
			})
		}),
	)
}