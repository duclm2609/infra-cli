package tagpolicy

import (
	"fmt"
	"sort"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestPolicyParser_Parse_ValidPolicy(t *testing.T) {
	// Test case from the design document
	policyJSON := `{
		"tags": {
			"Environment": {
				"tag_key": {
					"@@assign": "Environment"
				},
				"tag_value": {
					"@@assign": ["Production", "Development", "Staging", "Test"]
				},
				"enforced_for": {
					"@@assign": ["ec2:instance", "s3:bucket"]
				}
			},
			"CostCenter": {
				"tag_key": {
					"@@assign": "CostCenter"
				},
				"tag_value": {
					"@@assign": ["Engineering", "Marketing", "Operations"]
				}
			}
		}
	}`

	parser := NewPolicyParser()
	policy, err := parser.Parse(policyJSON)
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	if policy == nil {
		t.Fatal("Parse() returned nil policy")
	}

	if len(policy.Tags) != 2 {
		t.Errorf("Expected 2 tag keys, got %d", len(policy.Tags))
	}

	// Find tags by name (order is not guaranteed due to Go map iteration)
	var costCenter, environment *TagKey
	for i := range policy.Tags {
		switch policy.Tags[i].Name {
		case "CostCenter":
			costCenter = &policy.Tags[i]
		case "Environment":
			environment = &policy.Tags[i]
		}
	}

	// Verify CostCenter tag
	if costCenter == nil {
		t.Fatal("Expected to find 'CostCenter' tag key")
	}
	if len(costCenter.Values) != 3 {
		t.Errorf("Expected CostCenter to have 3 values, got %d", len(costCenter.Values))
	}
	if len(costCenter.EnforcedFor) != 0 {
		t.Errorf("Expected CostCenter to have no enforced_for, got %d", len(costCenter.EnforcedFor))
	}

	// Verify Environment tag
	if environment == nil {
		t.Fatal("Expected to find 'Environment' tag key")
	}
	if len(environment.Values) != 4 {
		t.Errorf("Expected Environment to have 4 values, got %d", len(environment.Values))
	}
	expectedValues := []string{"Production", "Development", "Staging", "Test"}
	for i, expected := range expectedValues {
		if environment.Values[i] != expected {
			t.Errorf("Expected Environment value %d to be '%s', got '%s'", i, expected, environment.Values[i])
		}
	}
	if len(environment.EnforcedFor) != 2 {
		t.Errorf("Expected Environment to have 2 enforced_for, got %d", len(environment.EnforcedFor))
	}
}

func TestPolicyParser_Parse_EmptyContent(t *testing.T) {
	parser := NewPolicyParser()
	_, err := parser.Parse("")
	if err == nil {
		t.Error("Parse() should return error for empty content")
	}
}

func TestPolicyParser_Parse_InvalidJSON(t *testing.T) {
	parser := NewPolicyParser()
	_, err := parser.Parse("not valid json")
	if err == nil {
		t.Error("Parse() should return error for invalid JSON")
	}
}

func TestPolicyParser_Parse_MissingTagsField(t *testing.T) {
	parser := NewPolicyParser()
	_, err := parser.Parse(`{"other": "field"}`)
	if err == nil {
		t.Error("Parse() should return error when 'tags' field is missing")
	}
}

func TestPolicyParser_Parse_TagsNotObject(t *testing.T) {
	parser := NewPolicyParser()
	_, err := parser.Parse(`{"tags": "not an object"}`)
	if err == nil {
		t.Error("Parse() should return error when 'tags' is not an object")
	}
}

func TestPolicyParser_Parse_SingleStringValue(t *testing.T) {
	// Test that single string values in @@assign are handled correctly
	policyJSON := `{
		"tags": {
			"Owner": {
				"tag_key": {
					"@@assign": "Owner"
				},
				"tag_value": {
					"@@assign": "admin"
				}
			}
		}
	}`

	parser := NewPolicyParser()
	policy, err := parser.Parse(policyJSON)
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	if len(policy.Tags) != 1 {
		t.Fatalf("Expected 1 tag key, got %d", len(policy.Tags))
	}

	owner := policy.Tags[0]
	if owner.Name != "Owner" {
		t.Errorf("Expected tag key name 'Owner', got '%s'", owner.Name)
	}
	if len(owner.Values) != 1 || owner.Values[0] != "admin" {
		t.Errorf("Expected single value 'admin', got %v", owner.Values)
	}
}

func TestPolicyParser_Parse_NoTagKeyField(t *testing.T) {
	// Test that when tag_key field is missing, the parent key name is used
	policyJSON := `{
		"tags": {
			"Project": {
				"tag_value": {
					"@@assign": ["Alpha", "Beta"]
				}
			}
		}
	}`

	parser := NewPolicyParser()
	policy, err := parser.Parse(policyJSON)
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	if len(policy.Tags) != 1 {
		t.Fatalf("Expected 1 tag key, got %d", len(policy.Tags))
	}

	project := policy.Tags[0]
	if project.Name != "Project" {
		t.Errorf("Expected tag key name 'Project' (from parent key), got '%s'", project.Name)
	}
}

func TestPolicyParser_Parse_EmptyTags(t *testing.T) {
	policyJSON := `{"tags": {}}`

	parser := NewPolicyParser()
	policy, err := parser.Parse(policyJSON)
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	if len(policy.Tags) != 0 {
		t.Errorf("Expected 0 tag keys, got %d", len(policy.Tags))
	}
}

func TestPolicyParser_ParseTagKey_InvalidTagKeyData(t *testing.T) {
	parser := NewPolicyParser()

	// Test with invalid tag_value (not a string, array, or object with @@assign)
	keyData := map[string]interface{}{
		"tag_key":   "TestKey",
		"tag_value": 12345, // number is invalid
	}

	tagKey, err := parser.ParseTagKey("TestKey", keyData)
	// Should not error, just skip the invalid tag_value
	if err != nil {
		t.Fatalf("ParseTagKey() returned unexpected error: %v", err)
	}
	if tagKey.Name != "TestKey" {
		t.Errorf("Expected tag key name 'TestKey', got '%s'", tagKey.Name)
	}
	// Values should be nil/empty since tag_value was invalid
	if len(tagKey.Values) != 0 {
		t.Errorf("Expected no values for invalid tag_value, got %v", tagKey.Values)
	}
}


// =============================================================================
// Property-Based Tests
// =============================================================================

// genIdentifier generates valid identifier strings for tag names and values.
// Uses alphanumeric characters starting with a letter to ensure valid identifiers.
func genIdentifier() gopter.Gen {
	return gen.Identifier().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 50
	})
}

// genStringSlice generates a slice of identifier strings.
func genStringSlice() gopter.Gen {
	return gen.SliceOf(genIdentifier()).Map(func(slice []string) []string {
		// Ensure uniqueness and limit size for reasonable test data
		seen := make(map[string]bool)
		result := make([]string, 0, len(slice))
		for _, s := range slice {
			if !seen[s] && len(result) < 10 {
				seen[s] = true
				result = append(result, s)
			}
		}
		return result
	})
}

// genTagKey generates a random TagKey with valid name, values, and enforced_for.
func genTagKey() gopter.Gen {
	return gopter.CombineGens(
		genIdentifier(),   // Name
		genStringSlice(),  // Values
		genStringSlice(),  // EnforcedFor
	).Map(func(vals []interface{}) TagKey {
		return TagKey{
			Name:        vals[0].(string),
			Values:      vals[1].([]string),
			EnforcedFor: vals[2].([]string),
		}
	})
}

// genTagPolicy generates a random TagPolicy with unique tag keys.
func genTagPolicy() gopter.Gen {
	return gen.SliceOf(genTagKey()).Map(func(keys []TagKey) *TagPolicy {
		// Ensure unique tag key names
		seen := make(map[string]bool)
		uniqueKeys := make([]TagKey, 0, len(keys))
		for _, key := range keys {
			if !seen[key.Name] && len(uniqueKeys) < 10 {
				seen[key.Name] = true
				uniqueKeys = append(uniqueKeys, key)
			}
		}
		return &TagPolicy{Tags: uniqueKeys}
	})
}

// tagPoliciesEquivalent checks if two TagPolicy structures are equivalent.
// Two policies are equivalent if they have the same tag keys with the same
// values and enforced_for constraints (order-independent comparison).
func tagPoliciesEquivalent(a, b *TagPolicy) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a.Tags) != len(b.Tags) {
		return false
	}

	// Create maps for comparison (order-independent)
	aMap := make(map[string]TagKey)
	for _, tag := range a.Tags {
		aMap[tag.Name] = tag
	}

	bMap := make(map[string]TagKey)
	for _, tag := range b.Tags {
		bMap[tag.Name] = tag
	}

	// Compare each tag key
	for name, aTag := range aMap {
		bTag, exists := bMap[name]
		if !exists {
			return false
		}
		if !tagKeysEquivalent(aTag, bTag) {
			return false
		}
	}

	return true
}

// tagKeysEquivalent checks if two TagKey structures are equivalent.
func tagKeysEquivalent(a, b TagKey) bool {
	if a.Name != b.Name {
		return false
	}
	if !stringSlicesEquivalent(a.Values, b.Values) {
		return false
	}
	if !stringSlicesEquivalent(a.EnforcedFor, b.EnforcedFor) {
		return false
	}
	return true
}

// stringSlicesEquivalent checks if two string slices contain the same elements
// (order-independent comparison).
func stringSlicesEquivalent(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Sort copies for comparison
	aSorted := make([]string, len(a))
	copy(aSorted, a)
	sort.Strings(aSorted)

	bSorted := make([]string, len(b))
	copy(bSorted, b)
	sort.Strings(bSorted)

	for i := range aSorted {
		if aSorted[i] != bSorted[i] {
			return false
		}
	}
	return true
}

// TestPolicyParsingRoundTrip is a property-based test that verifies:
// For any valid TagPolicy structure, serializing it to the AWS tag policy JSON
// format and then parsing it back SHALL produce an equivalent TagPolicy with
// the same tag keys, values, and enforced_for constraints.
//
// Feature: aws-tag-policy, Property 1: Policy Parsing Round-Trip
// **Validates: Requirements 2.1, 2.2, 2.3, 2.5**
func TestPolicyParsingRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	parser := NewPolicyParser()

	properties.Property("round-trip serialization preserves TagPolicy equivalence", prop.ForAll(
		func(policy *TagPolicy) bool {
			// Step 1: Serialize the TagPolicy to AWS JSON format
			jsonStr, err := policy.ToAWSJSON()
			if err != nil {
				t.Logf("ToAWSJSON failed: %v", err)
				return false
			}

			// Step 2: Parse the JSON back into a TagPolicy
			parsedPolicy, err := parser.Parse(jsonStr)
			if err != nil {
				t.Logf("Parse failed: %v\nJSON: %s", err, jsonStr)
				return false
			}

			// Step 3: Verify the parsed policy is equivalent to the original
			if !tagPoliciesEquivalent(policy, parsedPolicy) {
				t.Logf("Policies not equivalent:\nOriginal: %+v\nParsed: %+v\nJSON: %s",
					policy, parsedPolicy, jsonStr)
				return false
			}

			return true
		},
		genTagPolicy(),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 2: Malformed JSON Produces Errors
// =============================================================================

// genInvalidJSONString generates random strings that are not valid JSON.
// These include random bytes, truncated JSON, and malformed structures.
func genInvalidJSONString() gopter.Gen {
	return gen.OneGenOf(
		// Random alphanumeric strings (not JSON)
		gen.Identifier().SuchThat(func(s string) bool {
			return len(s) > 0
		}),
		// Strings with special characters that break JSON
		gen.AnyString().Map(func(s string) string {
			if len(s) == 0 {
				return "not json"
			}
			return s + "{"
		}),
		// Truncated JSON objects
		gen.Const("{\"tags\":"),
		gen.Const("{\"tags\": {"),
		gen.Const("{\"tags\": {\"key\":"),
		// Invalid JSON with unquoted keys
		gen.Const("{tags: {}}"),
		// JSON with trailing commas
		gen.Const("{\"tags\": {},}"),
		// JSON with single quotes (invalid in JSON)
		gen.Const("{'tags': {}}"),
		// Empty or whitespace-only strings
		gen.Const(""),
		gen.Const("   "),
		gen.Const("\n\t"),
	)
}

// genValidJSONMissingTags generates valid JSON objects that are missing the required 'tags' field.
func genValidJSONMissingTags() gopter.Gen {
	return gen.OneGenOf(
		// Empty object
		gen.Const("{}"),
		// Object with other fields but no 'tags'
		gen.Identifier().Map(func(key string) string {
			return fmt.Sprintf(`{"%s": "value"}`, key)
		}),
		// Object with nested structure but no 'tags'
		gen.Identifier().Map(func(key string) string {
			return fmt.Sprintf(`{"%s": {"nested": "value"}}`, key)
		}),
		// Object with array but no 'tags'
		gen.Identifier().Map(func(key string) string {
			return fmt.Sprintf(`{"%s": ["a", "b", "c"]}`, key)
		}),
		// Object with multiple fields but no 'tags'
		gen.Const(`{"name": "test", "version": 1, "enabled": true}`),
	)
}

// genValidJSONTagsNotObject generates valid JSON where 'tags' field exists but is not an object.
func genValidJSONTagsNotObject() gopter.Gen {
	return gen.OneGenOf(
		// tags is a string
		gen.AnyString().Map(func(s string) string {
			escaped := escapeJSONString(s)
			return fmt.Sprintf(`{"tags": "%s"}`, escaped)
		}),
		// tags is a number
		gen.Int().Map(func(n int) string {
			return fmt.Sprintf(`{"tags": %d}`, n)
		}),
		// tags is a boolean
		gen.Bool().Map(func(b bool) string {
			return fmt.Sprintf(`{"tags": %t}`, b)
		}),
		// tags is an array
		gen.SliceOf(gen.Identifier()).Map(func(arr []string) string {
			if len(arr) == 0 {
				return `{"tags": []}`
			}
			items := ""
			for i, s := range arr {
				if i > 0 {
					items += ", "
				}
				items += fmt.Sprintf(`"%s"`, s)
			}
			return fmt.Sprintf(`{"tags": [%s]}`, items)
		}),
		// tags is null
		gen.Const(`{"tags": null}`),
	)
}

// genValidJSONTagEntryNotObject generates valid JSON where 'tags' is an object
// but individual tag entries are not objects.
func genValidJSONTagEntryNotObject() gopter.Gen {
	return gen.OneGenOf(
		// Tag entry is a string
		gen.Identifier().Map(func(key string) string {
			return fmt.Sprintf(`{"tags": {"%s": "not an object"}}`, key)
		}),
		// Tag entry is a number
		gen.Identifier().Map(func(key string) string {
			return fmt.Sprintf(`{"tags": {"%s": 123}}`, key)
		}),
		// Tag entry is a boolean
		gen.Identifier().Map(func(key string) string {
			return fmt.Sprintf(`{"tags": {"%s": true}}`, key)
		}),
		// Tag entry is an array
		gen.Identifier().Map(func(key string) string {
			return fmt.Sprintf(`{"tags": {"%s": ["a", "b"]}}`, key)
		}),
		// Tag entry is null
		gen.Identifier().Map(func(key string) string {
			return fmt.Sprintf(`{"tags": {"%s": null}}`, key)
		}),
		// Multiple tag entries where one is not an object
		gen.Identifier().Map(func(key string) string {
			return fmt.Sprintf(`{"tags": {"ValidKey": {"tag_key": {"@@assign": "ValidKey"}}, "%s": "invalid"}}`, key)
		}),
	)
}

// escapeJSONString escapes special characters for JSON string values.
func escapeJSONString(s string) string {
	result := ""
	for _, r := range s {
		switch r {
		case '"':
			result += `\"`
		case '\\':
			result += `\\`
		case '\n':
			result += `\n`
		case '\r':
			result += `\r`
		case '\t':
			result += `\t`
		default:
			if r < 32 {
				result += fmt.Sprintf(`\u%04x`, r)
			} else {
				result += string(r)
			}
		}
	}
	return result
}

// genMalformedInput combines all malformed input generators.
func genMalformedInput() gopter.Gen {
	return gen.OneGenOf(
		genInvalidJSONString(),
		genValidJSONMissingTags(),
		genValidJSONTagsNotObject(),
		genValidJSONTagEntryNotObject(),
	)
}

// TestMalformedJSONProducesErrors is a property-based test that verifies:
// For any string that is not valid JSON or does not conform to the AWS tag policy
// schema, the parser SHALL return an error rather than a partial or incorrect TagPolicy.
//
// Feature: aws-tag-policy, Property 2: Malformed JSON Produces Errors
// **Validates: Requirements 2.4**
func TestMalformedJSONProducesErrors(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	parser := NewPolicyParser()

	// Property: Invalid JSON strings produce errors
	properties.Property("invalid JSON strings produce errors", prop.ForAll(
		func(input string) bool {
			policy, err := parser.Parse(input)
			// Must return an error
			if err == nil {
				t.Logf("Expected error for invalid JSON, but got nil. Input: %q, Policy: %+v", input, policy)
				return false
			}
			// Must not return a partial policy
			if policy != nil {
				t.Logf("Expected nil policy for invalid JSON, but got: %+v. Input: %q", policy, input)
				return false
			}
			return true
		},
		genInvalidJSONString(),
	))

	// Property: Valid JSON missing 'tags' field produces errors
	properties.Property("valid JSON missing 'tags' field produces errors", prop.ForAll(
		func(input string) bool {
			policy, err := parser.Parse(input)
			// Must return an error
			if err == nil {
				t.Logf("Expected error for missing 'tags' field, but got nil. Input: %q, Policy: %+v", input, policy)
				return false
			}
			// Must not return a partial policy
			if policy != nil {
				t.Logf("Expected nil policy for missing 'tags' field, but got: %+v. Input: %q", policy, input)
				return false
			}
			return true
		},
		genValidJSONMissingTags(),
	))

	// Property: Valid JSON with 'tags' not an object produces errors
	properties.Property("valid JSON with 'tags' not an object produces errors", prop.ForAll(
		func(input string) bool {
			policy, err := parser.Parse(input)
			// Must return an error
			if err == nil {
				t.Logf("Expected error for 'tags' not an object, but got nil. Input: %q, Policy: %+v", input, policy)
				return false
			}
			// Must not return a partial policy
			if policy != nil {
				t.Logf("Expected nil policy for 'tags' not an object, but got: %+v. Input: %q", policy, input)
				return false
			}
			return true
		},
		genValidJSONTagsNotObject(),
	))

	// Property: Valid JSON with tag entries not objects produces errors
	properties.Property("valid JSON with tag entries not objects produces errors", prop.ForAll(
		func(input string) bool {
			policy, err := parser.Parse(input)
			// Must return an error
			if err == nil {
				t.Logf("Expected error for tag entry not an object, but got nil. Input: %q, Policy: %+v", input, policy)
				return false
			}
			// Must not return a partial policy
			if policy != nil {
				t.Logf("Expected nil policy for tag entry not an object, but got: %+v. Input: %q", policy, input)
				return false
			}
			return true
		},
		genValidJSONTagEntryNotObject(),
	))

	// Combined property: Any malformed input produces errors
	properties.Property("any malformed input produces errors, not partial results", prop.ForAll(
		func(input string) bool {
			policy, err := parser.Parse(input)
			// Must return an error
			if err == nil {
				t.Logf("Expected error for malformed input, but got nil. Input: %q, Policy: %+v", input, policy)
				return false
			}
			// Must not return a partial policy
			if policy != nil {
				t.Logf("Expected nil policy for malformed input, but got: %+v. Input: %q", policy, input)
				return false
			}
			return true
		},
		genMalformedInput(),
	))

	properties.TestingRun(t)
}

func TestPolicyParser_Parse_PreservesCaseFromTagKeyAssign(t *testing.T) {
	// Test that the tag key name comes from tag_key.@@assign, not the JSON key
	// AWS may lowercase the JSON key but preserve case in tag_key.@@assign
	policyJSON := `{
		"tags": {
			"tcbs:cost-allocation:systemcategory": {
				"tag_key": {
					"@@assign": "tcbs:cost-allocation:SystemCategory"
				},
				"tag_value": {
					"@@assign": ["Value1", "Value2"]
				}
			}
		}
	}`

	parser := NewPolicyParser()
	policy, err := parser.Parse(policyJSON)
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	if len(policy.Tags) != 1 {
		t.Fatalf("Expected 1 tag key, got %d", len(policy.Tags))
	}

	// The name should come from tag_key.@@assign, preserving the correct case
	tagKey := policy.Tags[0]
	expectedName := "tcbs:cost-allocation:SystemCategory"
	if tagKey.Name != expectedName {
		t.Errorf("Expected tag key name '%s', got '%s'", expectedName, tagKey.Name)
	}
}

func TestPolicyParser_Parse_NoTagKeyField_UsesJSONKey(t *testing.T) {
	// Test when tag_key field is missing - should fall back to JSON key
	// This might be what AWS returns in some cases
	policyJSON := `{
		"tags": {
			"tcbs:cost-allocation:systemcategory": {
				"tag_value": {
					"@@assign": ["Value1", "Value2"]
				}
			}
		}
	}`

	parser := NewPolicyParser()
	policy, err := parser.Parse(policyJSON)
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	if len(policy.Tags) != 1 {
		t.Fatalf("Expected 1 tag key, got %d", len(policy.Tags))
	}

	// Without tag_key field, it falls back to the JSON key (lowercased)
	tagKey := policy.Tags[0]
	expectedName := "tcbs:cost-allocation:systemcategory"
	if tagKey.Name != expectedName {
		t.Errorf("Expected tag key name '%s', got '%s'", expectedName, tagKey.Name)
	}
}

func TestPolicyParser_Parse_AWSEffectivePolicyFormat(t *testing.T) {
	// Test the actual AWS effective policy format where tag_key and tag_value
	// are direct values, not wrapped in @@assign
	policyJSON := `{
		"tags": {
			"tcbs:cost-allocation:environment": {
				"tag_value": ["test", "sit", "uat", "prod"],
				"tag_key": "tcbs:cost-allocation:Environment"
			}
		}
	}`

	parser := NewPolicyParser()
	policy, err := parser.Parse(policyJSON)
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	if len(policy.Tags) != 1 {
		t.Fatalf("Expected 1 tag key, got %d", len(policy.Tags))
	}

	tagKey := policy.Tags[0]
	
	// The name should come from tag_key field, preserving the correct case
	expectedName := "tcbs:cost-allocation:Environment"
	if tagKey.Name != expectedName {
		t.Errorf("Expected tag key name '%s', got '%s'", expectedName, tagKey.Name)
	}

	// Values should be parsed correctly
	if len(tagKey.Values) != 4 {
		t.Errorf("Expected 4 values, got %d", len(tagKey.Values))
	}
}
