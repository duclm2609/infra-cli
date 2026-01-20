// Package tagpolicy provides functionality for retrieving and displaying
// AWS Organizations tag policies through an interactive TUI.
package tagpolicy

// TagPolicy represents the complete effective tag policy retrieved from
// AWS Organizations. It contains all tag keys defined in the policy.
//
// This struct maps to the parsed representation of the AWS tag policy JSON
// structure, which uses the "tags" field to contain all tag key definitions.
type TagPolicy struct {
	// Tags contains all tag keys defined in the effective tag policy.
	// Each TagKey includes the key name, allowed values, and enforcement rules.
	Tags []TagKey `json:"tags"`
}

// TagKey represents a single tag key with its configuration from the
// AWS Organizations tag policy.
//
// The JSON tags map to the AWS tag policy JSON structure where:
// - "tag_key" contains the tag key name
// - "tag_value" contains the list of allowed values (optional)
// - "enforced_for" contains the list of resource types where the tag is enforced (optional)
type TagKey struct {
	// Name is the tag key name (e.g., "Environment", "CostCenter").
	// This corresponds to the "tag_key" field in the AWS policy JSON.
	Name string `json:"tag_key"`

	// Values contains the list of allowed values for this tag key.
	// If empty, any value is allowed for this tag key.
	// This corresponds to the "tag_value" field in the AWS policy JSON.
	Values []string `json:"tag_value,omitempty"`

	// EnforcedFor contains the list of AWS resource types where this tag
	// is enforced (e.g., "ec2:instance", "s3:bucket").
	// If empty, the tag is not enforced for any specific resource types.
	// This corresponds to the "enforced_for" field in the AWS policy JSON.
	EnforcedFor []string `json:"enforced_for,omitempty"`
}
