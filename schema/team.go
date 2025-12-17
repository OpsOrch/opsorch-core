package schema

// TeamQuery defines filters passed to the active team provider.
// Providers decide how to apply these hints (server-side or client-side).
type TeamQuery struct {
	// Name is a substring / fuzzy match filter for team names.
	// Used for discovery (autocomplete, UIs, MCP listing, etc.).
	Name string `json:"name,omitempty"`

	// Tags filters teams by key/value tag pairs.
	// Providers may map these to labels, attributes, or custom metadata.
	Tags map[string]string `json:"tags,omitempty"`

	// Scope provides shared service/team/environment hints.
	// Providers can ignore fields they do not support.
	Scope QueryScope `json:"scope,omitempty"`

	// Metadata carries provider-specific query hints.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Team represents a normalized team or group from the active team provider.
// OpsOrch does not store or mutate teams; providers own the source of truth.
type Team struct {
	// ID is the canonical OpsOrch handle for this team.
	// This is what QueryScope.Team should carry.
	ID string `json:"id"`

	// Name is the human-readable display name of the team.
	Name string `json:"name"`

	// Parent is the canonical ID of this team's parent, if any.
	// Root-level teams leave this empty. Call the team provider's Get with this
	// value if you need the parent details.
	Parent string `json:"parent,omitempty"`

	// Tags contains normalized key/value tags for filtering and correlation.
	Tags map[string]string `json:"tags,omitempty"`

	// Metadata stores provider-specific fields not covered by the normalized schema.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// TeamMember is a lightweight, provider-backed representation of a team member.
// OpsOrch does not manage identities; this is a best-effort snapshot.
type TeamMember struct {
	// ID is the canonical handle for this person within OpsOrch.
	// Providers can map this to a user ID, email, or any stable upstream identifier.
	ID string `json:"id"`

	// Name is the display name for this member.
	Name string `json:"name,omitempty"`

	// Email is often the most stable handle for routing notifications, etc.
	Email string `json:"email,omitempty"`

	// Handle can be a chat/username (Slack handle, GitHub login, etc.).
	Handle string `json:"handle,omitempty"`

	// Role is a free-form normalized label like "owner", "manager", "member", "oncall".
	Role string `json:"role,omitempty"`

	// Metadata is provider-specific: raw user object, links, extra attributes, etc.
	Metadata map[string]any `json:"metadata,omitempty"`
}
