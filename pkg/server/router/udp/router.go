package udp

import (
	"context"
	"errors"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/server/provider"
	udpservice "github.com/traefik/traefik/v3/pkg/server/service/udp"
	"github.com/traefik/traefik/v3/pkg/udp"
)

type middlewareBuilder interface {
	BuildChain(ctx context.Context, names []string) *udp.Chain
}

// NewManager Creates a new Manager.
func NewManager(conf *runtime.Configuration,
	serviceManager *udpservice.Manager,
	middlewaresBuilder middlewareBuilder,
) *Manager {
	return &Manager{
		serviceManager:     serviceManager,
		middlewaresBuilder: middlewaresBuilder,
		conf:               conf,
	}
}

// Manager is a route/router manager.
type Manager struct {
	serviceManager     *udpservice.Manager
	middlewaresBuilder middlewareBuilder
	conf               *runtime.Configuration
}

func (m *Manager) getUDPRouters(ctx context.Context, entryPoints []string) map[string]map[string]*runtime.UDPRouterInfo {
	if m.conf != nil {
		return m.conf.GetUDPRoutersByEntryPoints(ctx, entryPoints)
	}

	return make(map[string]map[string]*runtime.UDPRouterInfo)
}

// BuildHandlers builds the handlers for the given entrypoints.
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string) map[string]udp.Handler {
	entryPointsRouters := m.getUDPRouters(rootCtx, entryPoints)

	entryPointHandlers := make(map[string]udp.Handler)
	for _, entryPointName := range entryPoints {
		entryPointName := entryPointName

		routers := entryPointsRouters[entryPointName]

		logger := log.Ctx(rootCtx).With().Str(logs.EntryPointName, entryPointName).Logger()
		ctx := logger.WithContext(rootCtx)

		if len(routers) > 1 {
			logger.Warn().Msg("Config has more than one udp router for a given entrypoint.")
		}

		handlers := m.buildEntryPointHandlers(ctx, routers)

		if len(handlers) > 0 {
			// As UDP support only one router per entrypoint, we only take the first one.
			entryPointHandlers[entryPointName] = handlers[0]
		}
	}
	return entryPointHandlers
}

func (m *Manager) buildEntryPointHandlers(ctx context.Context, configs map[string]*runtime.UDPRouterInfo) []udp.Handler {
	var rtNames []string
	for routerName := range configs {
		rtNames = append(rtNames, routerName)
	}

	sort.Slice(rtNames, func(i, j int) bool {
		return rtNames[i] > rtNames[j]
	})

	var handlers []udp.Handler

	for _, routerName := range rtNames {
		routerConfig := configs[routerName]
		logger := log.Ctx(ctx).With().Str(logs.RouterName, routerName).Logger()
		ctxRouter := logger.WithContext(provider.AddInContext(ctx, routerName))

		if routerConfig.Service == "" {
			err := errors.New("the service is missing on the udp router")
			routerConfig.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}

		handler, err := m.buildUDPHandler(ctxRouter, routerConfig)
		if err != nil {
			routerConfig.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}

		handlers = append(handlers, handler)
	}

	return handlers
}

func (m *Manager) buildUDPHandler(ctx context.Context, router *runtime.UDPRouterInfo) (udp.Handler, error) {
	var qualifiedNames []string
	for _, name := range router.Middlewares {
		qualifiedNames = append(qualifiedNames, provider.GetQualifiedName(ctx, name))
	}
	router.Middlewares = qualifiedNames

	if router.Service == "" {
		return nil, errors.New("the service is missing on the router")
	}

	sHandler, err := m.serviceManager.BuildUDP(ctx, router.Service)
	if err != nil {
		return nil, err
	}

	mHandler := m.middlewaresBuilder.BuildChain(ctx, router.Middlewares)

	return udp.NewChain().Extend(*mHandler).Then(sHandler)
}
