package schema

// ServiceQuery defines filters passed to the single active service provider.
// Providers decide how to apply these filters (server-side or client-side).
type ServiceQuery struct {
	// IDs filters services by their normalized OpsOrch ID.
	// Providers can map this to their upstream "id", "uid", or similar fields.
	IDs []string `json:"ids,omitempty"`

	// Name is a substring / fuzzy match filter for service names.
	// Providers typically match this against their display name or slug.
	Name string `json:"name,omitempty"`

	// Tags filters services by key/value tag pairs.
	// Providers may match these to labels, tags, attributes, or custom metadata.
	Tags map[string]string `json:"tags,omitempty"`

	// Limit restricts the maximum number of services returned.
	// Providers may apply this server-side or let OpsOrch slice results.
	Limit int `json:"limit,omitempty"`

	// Scope provides a shared set of filtering hints applied across providers.
	// Providers can ignore fields they do not support.
	Scope QueryScope `json:"scope,omitempty"`

	// Metadata carries provider-specific query hints.
	// This is an escape hatch: adapters may interpret these fields as needed.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Service represents a normalized service definition from the active service provider.
// OpsOrch does not create, modify, or delete services; providers own the source of truth.
type Service struct {
	// ID is the stable, normalized OpsOrch identifier for this service.
	// Typically derived from the provider's unique field (e.g. UID, ID, slug).
	ID string `json:"id"`

	// Name is the human-readable display name of the service.
	// Providers may map this from title, name, alias, or other upstream fields.
	Name string `json:"name"`

	// Tags contains normalized key/value tags for filtering and correlation.
	// Providers flatten their native tag/label formats into string pairs.
	Tags map[string]string `json:"tags,omitempty"`

	// Metadata stores provider-specific fields not covered by the normalized schema.
	// Providers may include raw upstream objects, attributes, or extended data here.
	Metadata map[string]any `json:"metadata,omitempty"`
}
