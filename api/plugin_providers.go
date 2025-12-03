package api

import (
	"context"

	"github.com/opsorch/opsorch-core/schema"
)

// Alert plugin provider -------------------------------------------------------

type alertPluginProvider struct {
	runner *pluginRunner
}

func newAlertPluginProvider(path string, cfg map[string]any) alertPluginProvider {
	return alertPluginProvider{runner: newPluginRunner(path, cfg)}
}

func (p alertPluginProvider) Query(ctx context.Context, query schema.AlertQuery) ([]schema.Alert, error) {
	var res []schema.Alert
	return res, p.runner.call(ctx, "alert.query", query, &res)
}

func (p alertPluginProvider) Get(ctx context.Context, id string) (schema.Alert, error) {
	var res schema.Alert
	return res, p.runner.call(ctx, "alert.get", map[string]any{"id": id}, &res)
}

// Incident plugin provider ----------------------------------------------------

type incidentPluginProvider struct {
	runner *pluginRunner
}

func newIncidentPluginProvider(path string, cfg map[string]any) incidentPluginProvider {
	return incidentPluginProvider{runner: newPluginRunner(path, cfg)}
}

func (p incidentPluginProvider) Query(ctx context.Context, query schema.IncidentQuery) ([]schema.Incident, error) {
	var res []schema.Incident
	return res, p.runner.call(ctx, "incident.query", query, &res)
}

func (p incidentPluginProvider) List(ctx context.Context) ([]schema.Incident, error) {
	var res []schema.Incident
	return res, p.runner.call(ctx, "incident.list", nil, &res)
}

func (p incidentPluginProvider) Get(ctx context.Context, id string) (schema.Incident, error) {
	var res schema.Incident
	return res, p.runner.call(ctx, "incident.get", map[string]any{"id": id}, &res)
}

func (p incidentPluginProvider) Create(ctx context.Context, in schema.CreateIncidentInput) (schema.Incident, error) {
	var res schema.Incident
	return res, p.runner.call(ctx, "incident.create", in, &res)
}

func (p incidentPluginProvider) Update(ctx context.Context, id string, in schema.UpdateIncidentInput) (schema.Incident, error) {
	payload := map[string]any{"id": id, "input": in}
	var res schema.Incident
	return res, p.runner.call(ctx, "incident.update", payload, &res)
}

func (p incidentPluginProvider) GetTimeline(ctx context.Context, id string) ([]schema.TimelineEntry, error) {
	var res []schema.TimelineEntry
	return res, p.runner.call(ctx, "incident.timeline.get", map[string]any{"id": id}, &res)
}

func (p incidentPluginProvider) AppendTimeline(ctx context.Context, id string, entry schema.TimelineAppendInput) error {
	payload := map[string]any{"id": id, "entry": entry}
	return p.runner.call(ctx, "incident.timeline.append", payload, nil)
}

// Log plugin provider ---------------------------------------------------------

type logPluginProvider struct {
	runner *pluginRunner
}

func newLogPluginProvider(path string, cfg map[string]any) logPluginProvider {
	return logPluginProvider{runner: newPluginRunner(path, cfg)}
}

func (p logPluginProvider) Query(ctx context.Context, query schema.LogQuery) ([]schema.LogEntry, error) {
	var res []schema.LogEntry
	return res, p.runner.call(ctx, "log.query", query, &res)
}

// Metric plugin provider ------------------------------------------------------

type metricPluginProvider struct {
	runner *pluginRunner
}

func newMetricPluginProvider(path string, cfg map[string]any) metricPluginProvider {
	return metricPluginProvider{runner: newPluginRunner(path, cfg)}
}

func (p metricPluginProvider) Query(ctx context.Context, query schema.MetricQuery) ([]schema.MetricSeries, error) {
	var res []schema.MetricSeries
	return res, p.runner.call(ctx, "metric.query", query, &res)
}

func (p metricPluginProvider) Describe(ctx context.Context, scope schema.QueryScope) ([]schema.MetricDescriptor, error) {
	var res []schema.MetricDescriptor
	return res, p.runner.call(ctx, "metric.describe", scope, &res)
}

// Ticket plugin provider ------------------------------------------------------

type ticketPluginProvider struct {
	runner *pluginRunner
}

func newTicketPluginProvider(path string, cfg map[string]any) ticketPluginProvider {
	return ticketPluginProvider{runner: newPluginRunner(path, cfg)}
}

func (p ticketPluginProvider) Query(ctx context.Context, query schema.TicketQuery) ([]schema.Ticket, error) {
	var res []schema.Ticket
	return res, p.runner.call(ctx, "ticket.query", query, &res)
}

func (p ticketPluginProvider) Get(ctx context.Context, id string) (schema.Ticket, error) {
	var res schema.Ticket
	return res, p.runner.call(ctx, "ticket.get", map[string]any{"id": id}, &res)
}

func (p ticketPluginProvider) Create(ctx context.Context, in schema.CreateTicketInput) (schema.Ticket, error) {
	var res schema.Ticket
	return res, p.runner.call(ctx, "ticket.create", in, &res)
}

func (p ticketPluginProvider) Update(ctx context.Context, id string, in schema.UpdateTicketInput) (schema.Ticket, error) {
	payload := map[string]any{"id": id, "input": in}
	var res schema.Ticket
	return res, p.runner.call(ctx, "ticket.update", payload, &res)
}

// Messaging plugin provider ---------------------------------------------------

type messagingPluginProvider struct {
	runner *pluginRunner
}

func newMessagingPluginProvider(path string, cfg map[string]any) messagingPluginProvider {
	return messagingPluginProvider{runner: newPluginRunner(path, cfg)}
}

func (p messagingPluginProvider) Send(ctx context.Context, msg schema.Message) (schema.MessageResult, error) {
	var res schema.MessageResult
	return res, p.runner.call(ctx, "messaging.send", msg, &res)
}

// Service plugin provider ----------------------------------------------------

type servicePluginProvider struct {
	runner *pluginRunner
}

func newServicePluginProvider(path string, cfg map[string]any) servicePluginProvider {
	return servicePluginProvider{runner: newPluginRunner(path, cfg)}
}

func (p servicePluginProvider) Query(ctx context.Context, query schema.ServiceQuery) ([]schema.Service, error) {
	var res []schema.Service
	return res, p.runner.call(ctx, "service.query", query, &res)
}

// Secret plugin provider -----------------------------------------------------

type secretPluginProvider struct {
	runner *pluginRunner
}

func newSecretPluginProvider(path string, cfg map[string]any) secretPluginProvider {
	return secretPluginProvider{runner: newPluginRunner(path, cfg)}
}

func (p secretPluginProvider) Get(ctx context.Context, key string) (string, error) {
	var res string
	return res, p.runner.call(ctx, "secret.get", map[string]any{"key": key}, &res)
}

func (p secretPluginProvider) Put(ctx context.Context, key, value string) error {
	payload := map[string]any{"key": key, "value": value}
	return p.runner.call(ctx, "secret.put", payload, nil)
}
