package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
)

type udpRouterRepresentation struct {
	*runtime.UDPRouterInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
}

func newUDPRouterRepresentation(name string, rt *runtime.UDPRouterInfo) udpRouterRepresentation {
	return udpRouterRepresentation{
		UDPRouterInfo: rt,
		Name:          name,
		Provider:      getProviderName(name),
	}
}

type udpServiceRepresentation struct {
	*runtime.UDPServiceInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
	Type     string `json:"type,omitempty"`
}

func newUDPServiceRepresentation(name string, si *runtime.UDPServiceInfo) udpServiceRepresentation {
	return udpServiceRepresentation{
		UDPServiceInfo: si,
		Name:           name,
		Provider:       getProviderName(name),
		Type:           strings.ToLower(extractType(si.UDPService)),
	}
}

type udpMiddlewareRepresentation struct {
	*runtime.UDPMiddlewareInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
	Type     string `json:"type,omitempty"`
}

func newUDPMiddlewareRepresentation(name string, mi *runtime.UDPMiddlewareInfo) udpMiddlewareRepresentation {
	return udpMiddlewareRepresentation{
		UDPMiddlewareInfo: mi,
		Name:              name,
		Provider:          getProviderName(name),
		Type:              strings.ToLower(extractType(mi.UDPMiddleware)),
	}
}

func (h Handler) getUDPRouters(rw http.ResponseWriter, request *http.Request) {
	results := make([]udpRouterRepresentation, 0, len(h.runtimeConfiguration.UDPRouters))

	query := request.URL.Query()
	criterion := newSearchCriterion(query)

	for name, rt := range h.runtimeConfiguration.UDPRouters {
		if keepUDPRouter(name, rt, criterion) {
			results = append(results, newUDPRouterRepresentation(name, rt))
		}
	}

	sortRouters(query, results)

	rw.Header().Set("Content-Type", "application/json")

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		writeError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getUDPRouter(rw http.ResponseWriter, request *http.Request) {
	scapedRouterID := mux.Vars(request)["routerID"]

	routerID, err := url.PathUnescape(scapedRouterID)
	if err != nil {
		writeError(rw, fmt.Sprintf("unable to decode routerID %q: %s", scapedRouterID, err), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	router, ok := h.runtimeConfiguration.UDPRouters[routerID]
	if !ok {
		writeError(rw, fmt.Sprintf("router not found: %s", routerID), http.StatusNotFound)
		return
	}

	result := newUDPRouterRepresentation(routerID, router)

	err = json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getUDPServices(rw http.ResponseWriter, request *http.Request) {
	results := make([]udpServiceRepresentation, 0, len(h.runtimeConfiguration.UDPServices))

	query := request.URL.Query()
	criterion := newSearchCriterion(query)

	for name, si := range h.runtimeConfiguration.UDPServices {
		if keepUDPService(name, si, criterion) {
			results = append(results, newUDPServiceRepresentation(name, si))
		}
	}

	sortServices(query, results)

	rw.Header().Set("Content-Type", "application/json")

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		writeError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getUDPService(rw http.ResponseWriter, request *http.Request) {
	scapedServiceID := mux.Vars(request)["serviceID"]

	serviceID, err := url.PathUnescape(scapedServiceID)
	if err != nil {
		writeError(rw, fmt.Sprintf("unable to decode serviceID %q: %s", scapedServiceID, err), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	service, ok := h.runtimeConfiguration.UDPServices[serviceID]
	if !ok {
		writeError(rw, fmt.Sprintf("service not found: %s", serviceID), http.StatusNotFound)
		return
	}

	result := newUDPServiceRepresentation(serviceID, service)

	err = json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getUDPMiddlewares(rw http.ResponseWriter, request *http.Request) {
	results := make([]udpMiddlewareRepresentation, 0, len(h.runtimeConfiguration.Middlewares))

	query := request.URL.Query()
	criterion := newSearchCriterion(query)

	for name, mi := range h.runtimeConfiguration.UDPMiddlewares {
		if keepUDPMiddleware(name, mi, criterion) {
			results = append(results, newUDPMiddlewareRepresentation(name, mi))
		}
	}

	sortMiddlewares(query, results)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	rw.Header().Set("Content-Type", "application/json")

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		writeError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getUDPMiddleware(rw http.ResponseWriter, request *http.Request) {
	scapedMiddlewareID := mux.Vars(request)["middlewareID"]

	middlewareID, err := url.PathUnescape(scapedMiddlewareID)
	if err != nil {
		writeError(rw, fmt.Sprintf("unable to decode middlewareID %q: %s", scapedMiddlewareID, err), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	middleware, ok := h.runtimeConfiguration.UDPMiddlewares[middlewareID]
	if !ok {
		writeError(rw, fmt.Sprintf("middleware not found: %s", middlewareID), http.StatusNotFound)
		return
	}

	result := newUDPMiddlewareRepresentation(middlewareID, middleware)

	err = json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func keepUDPRouter(name string, item *runtime.UDPRouterInfo, criterion *searchCriterion) bool {
	if criterion == nil {
		return true
	}

	return criterion.withStatus(item.Status) &&
		criterion.searchIn(name) &&
		criterion.filterService(item.Service) &&
		criterion.filterMiddleware(item.Middlewares)
}

func keepUDPService(name string, item *runtime.UDPServiceInfo, criterion *searchCriterion) bool {
	if criterion == nil {
		return true
	}

	return criterion.withStatus(item.Status) && criterion.searchIn(name)
}

func keepUDPMiddleware(name string, item *runtime.UDPMiddlewareInfo, criterion *searchCriterion) bool {
	if criterion == nil {
		return true
	}

	return criterion.withStatus(item.Status) && criterion.searchIn(name)
}
