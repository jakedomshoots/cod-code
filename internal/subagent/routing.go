package subagent

import (
	"context"
	"fmt"
	"strings"

	"ceoharness/internal/model"
)

type RoutingRunner struct {
	defaultClient    model.Client
	defaultMetadata  RouteMetadata
	clients          map[string]model.Client
	metadata         map[string]RouteMetadata
	providerClients  map[string]model.Client
	providerMetadata map[string]RouteMetadata
	fallbackClient   model.Client
	fallbackMetadata RouteMetadata
	confidenceFloor  float64
}

type RouteMetadata struct {
	Source       string
	ProviderName string
}

type RouteError struct {
	Metadata RouteMetadata
	Err      error
}

func (e *RouteError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *RouteError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type RoutingConfig struct {
	DefaultClient    model.Client
	DefaultMetadata  RouteMetadata
	Clients          map[string]model.Client
	Metadata         map[string]RouteMetadata
	ProviderClients  map[string]model.Client
	ProviderMetadata map[string]RouteMetadata
	FallbackClient   model.Client
	FallbackMetadata RouteMetadata
	MinConfidence    float64
}

func NewRoutingRunner(defaultClient model.Client, clients map[string]model.Client) RoutingRunner {
	return NewRoutingRunnerWithConfig(RoutingConfig{
		DefaultClient:   defaultClient,
		DefaultMetadata: RouteMetadata{Source: "local"},
		Clients:         clients,
	})
}

func NewRoutingRunnerWithConfig(cfg RoutingConfig) RoutingRunner {
	copied := make(map[string]model.Client, len(cfg.Clients))
	for agentName, client := range cfg.Clients {
		copied[agentName] = client
	}
	copiedMetadata := make(map[string]RouteMetadata, len(cfg.Metadata))
	for agentName, metadata := range cfg.Metadata {
		copiedMetadata[agentName] = metadata
	}
	copiedProviderClients := make(map[string]model.Client, len(cfg.ProviderClients))
	for providerName, client := range cfg.ProviderClients {
		copiedProviderClients[providerName] = client
	}
	copiedProviderMetadata := make(map[string]RouteMetadata, len(cfg.ProviderMetadata))
	for providerName, metadata := range cfg.ProviderMetadata {
		copiedProviderMetadata[providerName] = metadata
	}
	return RoutingRunner{
		defaultClient:    cfg.DefaultClient,
		defaultMetadata:  cfg.DefaultMetadata,
		clients:          copied,
		metadata:         copiedMetadata,
		providerClients:  copiedProviderClients,
		providerMetadata: copiedProviderMetadata,
		fallbackClient:   cfg.FallbackClient,
		fallbackMetadata: cfg.FallbackMetadata,
		confidenceFloor:  cfg.MinConfidence,
	}
}

func (r RoutingRunner) Run(ctx context.Context, packet TaskPacket) (Result, error) {
	client, metadata, err := r.routeForPacket(packet)
	if err != nil {
		return Result{}, err
	}
	result, err := NewRunnerWithModel(client).Run(ctx, packet)
	if err != nil {
		if r.canUseFallback(metadata) {
			return r.runFallback(ctx, packet, metadata, fallbackReasonForError(err), err)
		}
		return Result{}, &RouteError{Metadata: metadata, Err: err}
	}
	result = resultWithRouteMetadata(result, metadata)
	if err := lowConfidenceError(result, r.minConfidence()); err != nil {
		if r.canUseFallback(metadata) {
			return r.runFallback(ctx, packet, metadata, "low_confidence", err)
		}
		return failedLowConfidenceResult(result, err), nil
	}
	return result, nil
}

func (r RoutingRunner) routeForPacket(packet TaskPacket) (model.Client, RouteMetadata, error) {
	providerName := strings.TrimSpace(packet.ProviderName)
	if providerName != "" {
		client, ok := r.providerClients[providerName]
		if !ok {
			return nil, RouteMetadata{}, &RouteError{
				Metadata: RouteMetadata{ProviderName: providerName},
				Err:      fmt.Errorf("provider %q route not configured", providerName),
			}
		}
		metadata := r.providerMetadata[providerName]
		if metadata.ProviderName == "" {
			metadata.ProviderName = providerName
		}
		return client, metadata, nil
	}
	if routedClient, ok := r.clients[packet.AgentName]; ok {
		return routedClient, r.metadata[packet.AgentName], nil
	}
	return r.defaultClient, r.defaultMetadata, nil
}

func (r RoutingRunner) canUseFallback(metadata RouteMetadata) bool {
	return r.fallbackClient != nil && !sameRouteMetadata(metadata, r.fallbackMetadata)
}

func (r RoutingRunner) runFallback(ctx context.Context, packet TaskPacket, primary RouteMetadata, reason string, primaryErr error) (Result, error) {
	result, err := NewRunnerWithModel(r.fallbackClient).Run(ctx, packet)
	if err != nil {
		return Result{}, &RouteError{
			Metadata: r.fallbackMetadata,
			Err:      fmt.Errorf("fallback after %s: %w", routeLabel(primary), err),
		}
	}
	result = resultWithRouteMetadata(result, r.fallbackMetadata)
	result.ProviderFallbackFrom = routeLabel(primary)
	result.ProviderFallbackReason = reason
	result.AttemptErrors = append([]string{primaryErr.Error()}, result.AttemptErrors...)
	if err := lowConfidenceError(result, r.minConfidence()); err != nil {
		return failedLowConfidenceResult(result, err), nil
	}
	return result, nil
}

func (r RoutingRunner) minConfidence() float64 {
	if r.confidenceFloor < 0 {
		return 0
	}
	return r.confidenceFloor
}

func resultWithRouteMetadata(result Result, metadata RouteMetadata) Result {
	if metadata.Source != "" {
		result.ModelSource = metadata.Source
	}
	result.ProviderName = metadata.ProviderName
	return result
}

func sameRouteMetadata(left RouteMetadata, right RouteMetadata) bool {
	return left.Source == right.Source && left.ProviderName == right.ProviderName
}

func routeLabel(metadata RouteMetadata) string {
	if metadata.ProviderName != "" {
		return metadata.ProviderName
	}
	if metadata.Source != "" {
		return metadata.Source
	}
	return "default"
}

func lowConfidenceError(result Result, minConfidence float64) error {
	if minConfidence <= 0 || result.Status != "pass" || result.Confidence == nil {
		return nil
	}
	if *result.Confidence >= minConfidence {
		return nil
	}
	return fmt.Errorf("confidence %.2f below minimum %.2f", *result.Confidence, minConfidence)
}

func failedLowConfidenceResult(result Result, err error) Result {
	result.Status = "fail"
	result.AttemptErrors = append(result.AttemptErrors, err.Error())
	return result
}
