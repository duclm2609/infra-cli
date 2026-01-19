package tagpolicy

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ToAWSJSON serializes a TagPolicy to the AWS tag policy JSON format.
// This produces JSON with the @@assign operators used by AWS Organizations.
//
// The output JSON structure is:
//
//	{
//	    "tags": {
//	        "TagKeyName": {
//	            "tag_key": { "@@assign": "TagKeyName" },
//	            "tag_value": { "@@assign": ["Value1", "Value2"] },
//	            "enforced_for": { "@@assign": ["ec2:instance", "s3:bucket"] }
//	        }
//	    }
//	}
func (p *TagPolicy) ToAWSJSON() (string, error) {
	if p == nil {
		return "", fmt.Errorf("cannot serialize nil TagPolicy")
	}

	// Build the tags map structure
	tagsMap := make(map[string]interface{})

	for _, tagKey := range p.Tags {
		keyData := make(map[string]interface{})

		// Always include tag_key with @@assign
		keyData["tag_key"] = map[string]interface{}{
			"@@assign": tagKey.Name,
		}

		// Include tag_value if there are values
		if len(tagKey.Values) > 0 {
			keyData["tag_value"] = map[string]interface{}{
				"@@assign": tagKey.Values,
			}
		}

		// Include enforced_for if there are constraints
		if len(tagKey.EnforcedFor) > 0 {
			keyData["enforced_for"] = map[string]interface{}{
				"@@assign": tagKey.EnforcedFor,
			}
		}

		tagsMap[tagKey.Name] = keyData
	}

	// Build the root structure
	root := map[string]interface{}{
		"tags": tagsMap,
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(root)
	if err != nil {
		return "", fmt.Errorf("failed to serialize TagPolicy to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// PolicyParser handles parsing of AWS Organizations tag policy JSON.
// It converts the raw JSON policy content into structured TagPolicy data.
type PolicyParser struct{}

// NewPolicyParser creates a new PolicyParser instance.
func NewPolicyParser() *PolicyParser {
	return &PolicyParser{}
}

// Parse converts raw policy JSON into a structured TagPolicy.
// It handles the AWS tag policy JSON structure with @@assign operators.
//
// The expected JSON structure is:
//
//	{
//	    "tags": {
//	        "TagKeyName": {
//	            "tag_key": { "@@assign": "TagKeyName" },
//	            "tag_value": { "@@assign": ["Value1", "Value2"] },
//	            "enforced_for": { "@@assign": ["ec2:instance", "s3:bucket"] }
//	        }
//	    }
//	}
//
// Returns a descriptive error if the JSON is malformed or doesn't conform
// to the expected AWS tag policy schema.
func (p *PolicyParser) Parse(policyContent string) (*TagPolicy, error) {
	if policyContent == "" {
		return nil, fmt.Errorf("policy content is empty")
	}

	// Parse the raw JSON into a generic structure
	var rawPolicy map[string]interface{}
	if err := json.Unmarshal([]byte(policyContent), &rawPolicy); err != nil {
		return nil, fmt.Errorf("failed to parse policy JSON: %w", err)
	}

	// Extract the "tags" field
	tagsRaw, ok := rawPolicy["tags"]
	if !ok {
		return nil, fmt.Errorf("policy JSON missing required 'tags' field")
	}

	tagsMap, ok := tagsRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("'tags' field must be an object, got %T", tagsRaw)
	}

	// Parse each tag key
	tagPolicy := &TagPolicy{
		Tags: make([]TagKey, 0, len(tagsMap)),
	}

	// Iterate over keys - preserve original order from JSON by iterating the map directly
	// Note: Go maps don't guarantee order, but AWS tag policies typically have a small
	// number of keys, and we preserve the tag_key name from the @@assign value
	for keyName, keyData := range tagsMap {
		keyDataMap, ok := keyData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("tag key '%s' must be an object, got %T", keyName, keyData)
		}

		tagKey, err := p.ParseTagKey(keyName, keyDataMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tag key '%s': %w", keyName, err)
		}

		tagPolicy.Tags = append(tagPolicy.Tags, *tagKey)
	}

	// Sort tag keys alphabetically by name (ascending)
	sort.Slice(tagPolicy.Tags, func(i, j int) bool {
		return tagPolicy.Tags[i].Name < tagPolicy.Tags[j].Name
	})

	return tagPolicy, nil
}

// ParseTagKey extracts tag key information from a policy JSON node.
// It handles the @@assign operator used in AWS tag policies.
//
// The keyData map is expected to have the following structure:
//
//	{
//	    "tag_key": { "@@assign": "KeyName" },
//	    "tag_value": { "@@assign": ["Value1", "Value2"] },
//	    "enforced_for": { "@@assign": ["ec2:instance"] }
//	}
//
// The tag_value and enforced_for fields are optional.
func (p *PolicyParser) ParseTagKey(keyName string, keyData map[string]interface{}) (*TagKey, error) {
	tagKey := &TagKey{}

	// Parse tag_key field (required)
	tagKeyName, err := p.extractAssignString(keyData, "tag_key")
	if err != nil {
		// If tag_key is not present, use the keyName from the parent object
		tagKey.Name = keyName
	} else {
		tagKey.Name = tagKeyName
	}

	// Parse tag_value field (optional)
	tagValues, err := p.extractAssignStringSlice(keyData, "tag_value")
	if err == nil {
		tagKey.Values = tagValues
	}

	// Parse enforced_for field (optional)
	enforcedFor, err := p.extractAssignStringSlice(keyData, "enforced_for")
	if err == nil {
		tagKey.EnforcedFor = enforcedFor
	}

	return tagKey, nil
}

// extractAssignString extracts a string value from a field.
// It handles two structures:
// 1. Direct string: { "field": "value" }
// 2. With @@assign operator: { "field": { "@@assign": "value" } }
func (p *PolicyParser) extractAssignString(data map[string]interface{}, field string) (string, error) {
	fieldData, ok := data[field]
	if !ok {
		return "", fmt.Errorf("field '%s' not found", field)
	}

	// Handle direct string value (AWS effective policy format)
	if strValue, ok := fieldData.(string); ok {
		return strValue, nil
	}

	// Handle @@assign operator format (AWS policy definition format)
	fieldMap, ok := fieldData.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("field '%s' must be a string or object with @@assign, got %T", field, fieldData)
	}

	assignValue, ok := fieldMap["@@assign"]
	if !ok {
		return "", fmt.Errorf("field '%s' missing @@assign operator", field)
	}

	strValue, ok := assignValue.(string)
	if !ok {
		return "", fmt.Errorf("@@assign value for '%s' must be a string, got %T", field, assignValue)
	}

	return strValue, nil
}

// extractAssignStringSlice extracts a string slice from a field.
// It handles multiple structures:
// 1. Direct array: { "field": ["value1", "value2"] }
// 2. Direct string: { "field": "value" }
// 3. With @@assign array: { "field": { "@@assign": ["value1", "value2"] } }
// 4. With @@assign string: { "field": { "@@assign": "value" } }
func (p *PolicyParser) extractAssignStringSlice(data map[string]interface{}, field string) ([]string, error) {
	fieldData, ok := data[field]
	if !ok {
		return nil, fmt.Errorf("field '%s' not found", field)
	}

	// Handle direct string value
	if strValue, ok := fieldData.(string); ok {
		return []string{strValue}, nil
	}

	// Handle direct array value (AWS effective policy format)
	if sliceValue, ok := fieldData.([]interface{}); ok {
		return p.convertToStringSlice(sliceValue, field)
	}

	// Handle @@assign operator format (AWS policy definition format)
	fieldMap, ok := fieldData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("field '%s' must be a string, array, or object with @@assign, got %T", field, fieldData)
	}

	assignValue, ok := fieldMap["@@assign"]
	if !ok {
		return nil, fmt.Errorf("field '%s' missing @@assign operator", field)
	}

	// Handle single string value in @@assign
	if strValue, ok := assignValue.(string); ok {
		return []string{strValue}, nil
	}

	// Handle array of values in @@assign
	sliceValue, ok := assignValue.([]interface{})
	if !ok {
		return nil, fmt.Errorf("@@assign value for '%s' must be a string or array, got %T", field, assignValue)
	}

	return p.convertToStringSlice(sliceValue, field)
}

// convertToStringSlice converts []interface{} to []string
func (p *PolicyParser) convertToStringSlice(slice []interface{}, field string) ([]string, error) {
	result := make([]string, 0, len(slice))
	for i, v := range slice {
		strVal, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("array element %d for '%s' must be a string, got %T", i, field, v)
		}
		result = append(result, strVal)
	}
	return result, nil
}
