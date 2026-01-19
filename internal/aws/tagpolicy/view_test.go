package tagpolicy

import (
	"fmt"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestNewViewModel_WithNilPolicy(t *testing.T) {
	vm := NewViewModel(nil)

	if vm == nil {
		t.Fatal("expected non-nil ViewModel")
	}
	if len(vm.TagKeys) != 0 {
		t.Errorf("expected empty TagKeys, got %d", len(vm.TagKeys))
	}
	if vm.SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex 0, got %d", vm.SelectedIndex)
	}
	if vm.ExpandedKeys == nil {
		t.Error("expected non-nil ExpandedKeys map")
	}
	if len(vm.ExpandedKeys) != 0 {
		t.Errorf("expected empty ExpandedKeys, got %d", len(vm.ExpandedKeys))
	}
}

func TestNewViewModel_WithEmptyPolicy(t *testing.T) {
	policy := &TagPolicy{Tags: []TagKey{}}
	vm := NewViewModel(policy)

	if vm == nil {
		t.Fatal("expected non-nil ViewModel")
	}
	if len(vm.TagKeys) != 0 {
		t.Errorf("expected empty TagKeys, got %d", len(vm.TagKeys))
	}
	if vm.SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex 0, got %d", vm.SelectedIndex)
	}
}

func TestNewViewModel_WithTagKeys(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production", "Development"}},
			{Name: "CostCenter", Values: []string{"Engineering", "Marketing"}},
			{Name: "Project", Values: []string{}},
		},
	}
	vm := NewViewModel(policy)

	if vm == nil {
		t.Fatal("expected non-nil ViewModel")
	}
	if len(vm.TagKeys) != 3 {
		t.Errorf("expected 3 TagKeys, got %d", len(vm.TagKeys))
	}
	if vm.SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex 0, got %d", vm.SelectedIndex)
	}
	if len(vm.ExpandedKeys) != 0 {
		t.Errorf("expected all keys collapsed (empty ExpandedKeys), got %d", len(vm.ExpandedKeys))
	}
}

func TestNewTagPolicyView_InitializesCorrectly(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production"}},
		},
	}
	view := NewTagPolicyView(policy)

	if view == nil {
		t.Fatal("expected non-nil TagPolicyView")
	}
	if view.Model() == nil {
		t.Fatal("expected non-nil model")
	}
	if len(view.Model().TagKeys) != 1 {
		t.Errorf("expected 1 TagKey, got %d", len(view.Model().TagKeys))
	}
}

func TestViewModel_IsExpanded(t *testing.T) {
	vm := &ViewModel{
		TagKeys:       []TagKey{{Name: "Test"}},
		SelectedIndex: 0,
		ExpandedKeys:  make(map[int]bool),
	}

	// Initially not expanded
	if vm.IsExpanded(0) {
		t.Error("expected index 0 to not be expanded initially")
	}

	// Set expanded
	vm.ExpandedKeys[0] = true
	if !vm.IsExpanded(0) {
		t.Error("expected index 0 to be expanded after setting")
	}
}

func TestViewModel_SetExpanded(t *testing.T) {
	vm := &ViewModel{
		TagKeys:       []TagKey{{Name: "Test1"}, {Name: "Test2"}},
		SelectedIndex: 0,
		ExpandedKeys:  make(map[int]bool),
	}

	// Set expanded to true
	vm.SetExpanded(0, true)
	if !vm.IsExpanded(0) {
		t.Error("expected index 0 to be expanded")
	}

	// Set expanded to false
	vm.SetExpanded(0, false)
	if vm.IsExpanded(0) {
		t.Error("expected index 0 to not be expanded")
	}

	// Verify the key is removed from map when set to false
	if _, exists := vm.ExpandedKeys[0]; exists {
		t.Error("expected index 0 to be removed from ExpandedKeys map")
	}
}

func TestViewModel_ToggleExpanded(t *testing.T) {
	vm := &ViewModel{
		TagKeys:       []TagKey{{Name: "Test"}},
		SelectedIndex: 0,
		ExpandedKeys:  make(map[int]bool),
	}

	// Toggle from collapsed to expanded
	vm.ToggleExpanded(0)
	if !vm.IsExpanded(0) {
		t.Error("expected index 0 to be expanded after first toggle")
	}

	// Toggle from expanded to collapsed
	vm.ToggleExpanded(0)
	if vm.IsExpanded(0) {
		t.Error("expected index 0 to be collapsed after second toggle")
	}
}

func TestViewModel_TagKeyCount(t *testing.T) {
	tests := []struct {
		name     string
		tagKeys  []TagKey
		expected int
	}{
		{"empty", []TagKey{}, 0},
		{"one key", []TagKey{{Name: "Test"}}, 1},
		{"multiple keys", []TagKey{{Name: "A"}, {Name: "B"}, {Name: "C"}}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &ViewModel{
				TagKeys:      tt.tagKeys,
				ExpandedKeys: make(map[int]bool),
			}
			if got := vm.TagKeyCount(); got != tt.expected {
				t.Errorf("TagKeyCount() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestNewViewModel_PreservesTagKeyData(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{
				Name:        "Environment",
				Values:      []string{"Production", "Development", "Staging"},
				EnforcedFor: []string{"ec2:instance", "s3:bucket"},
			},
		},
	}
	vm := NewViewModel(policy)

	if len(vm.TagKeys) != 1 {
		t.Fatalf("expected 1 TagKey, got %d", len(vm.TagKeys))
	}

	tagKey := vm.TagKeys[0]
	if tagKey.Name != "Environment" {
		t.Errorf("expected Name 'Environment', got '%s'", tagKey.Name)
	}
	if len(tagKey.Values) != 3 {
		t.Errorf("expected 3 Values, got %d", len(tagKey.Values))
	}
	if len(tagKey.EnforcedFor) != 2 {
		t.Errorf("expected 2 EnforcedFor, got %d", len(tagKey.EnforcedFor))
	}
}


// Tests for HandleKeyPress - Keyboard Navigation
// Requirements: 3.2, 3.3, 3.5

func TestHandleKeyPress_QuitKey(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production"}},
		},
	}
	view := NewTagPolicyView(policy)

	quit := view.HandleKeyPress(KeyQuit)
	if !quit {
		t.Error("expected quit=true when pressing 'q'")
	}
}

func TestHandleKeyPress_QuitKeyWithEmptyPolicy(t *testing.T) {
	view := NewTagPolicyView(&TagPolicy{Tags: []TagKey{}})

	quit := view.HandleKeyPress(KeyQuit)
	if !quit {
		t.Error("expected quit=true when pressing 'q' with empty policy")
	}
}

func TestHandleKeyPress_DownKey_IncrementsSelectedIndex(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment"},
			{Name: "CostCenter"},
			{Name: "Project"},
		},
	}
	view := NewTagPolicyView(policy)

	// Initial state: SelectedIndex = 0
	if view.Model().SelectedIndex != 0 {
		t.Fatalf("expected initial SelectedIndex 0, got %d", view.Model().SelectedIndex)
	}

	// Press down
	quit := view.HandleKeyPress(KeyDown)
	if quit {
		t.Error("expected quit=false when pressing down")
	}
	if view.Model().SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex 1 after down, got %d", view.Model().SelectedIndex)
	}

	// Press down again
	view.HandleKeyPress(KeyDown)
	if view.Model().SelectedIndex != 2 {
		t.Errorf("expected SelectedIndex 2 after second down, got %d", view.Model().SelectedIndex)
	}
}

func TestHandleKeyPress_DownKey_BoundedAtLastIndex(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment"},
			{Name: "CostCenter"},
		},
	}
	view := NewTagPolicyView(policy)

	// Move to last index
	view.HandleKeyPress(KeyDown)
	if view.Model().SelectedIndex != 1 {
		t.Fatalf("expected SelectedIndex 1, got %d", view.Model().SelectedIndex)
	}

	// Try to move past last index - should stay at 1
	view.HandleKeyPress(KeyDown)
	if view.Model().SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex to stay at 1 (bounded), got %d", view.Model().SelectedIndex)
	}

	// Press down multiple times - should still stay at 1
	view.HandleKeyPress(KeyDown)
	view.HandleKeyPress(KeyDown)
	if view.Model().SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex to stay at 1 after multiple downs, got %d", view.Model().SelectedIndex)
	}
}

func TestHandleKeyPress_UpKey_DecrementsSelectedIndex(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment"},
			{Name: "CostCenter"},
			{Name: "Project"},
		},
	}
	view := NewTagPolicyView(policy)

	// Move to index 2
	view.HandleKeyPress(KeyDown)
	view.HandleKeyPress(KeyDown)
	if view.Model().SelectedIndex != 2 {
		t.Fatalf("expected SelectedIndex 2, got %d", view.Model().SelectedIndex)
	}

	// Press up
	quit := view.HandleKeyPress(KeyUp)
	if quit {
		t.Error("expected quit=false when pressing up")
	}
	if view.Model().SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex 1 after up, got %d", view.Model().SelectedIndex)
	}

	// Press up again
	view.HandleKeyPress(KeyUp)
	if view.Model().SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex 0 after second up, got %d", view.Model().SelectedIndex)
	}
}

func TestHandleKeyPress_UpKey_BoundedAtZero(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment"},
			{Name: "CostCenter"},
		},
	}
	view := NewTagPolicyView(policy)

	// Initial state: SelectedIndex = 0
	if view.Model().SelectedIndex != 0 {
		t.Fatalf("expected initial SelectedIndex 0, got %d", view.Model().SelectedIndex)
	}

	// Try to move past first index - should stay at 0
	quit := view.HandleKeyPress(KeyUp)
	if quit {
		t.Error("expected quit=false when pressing up at index 0")
	}
	if view.Model().SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to stay at 0 (bounded), got %d", view.Model().SelectedIndex)
	}

	// Press up multiple times - should still stay at 0
	view.HandleKeyPress(KeyUp)
	view.HandleKeyPress(KeyUp)
	if view.Model().SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to stay at 0 after multiple ups, got %d", view.Model().SelectedIndex)
	}
}

func TestHandleKeyPress_EnterKey_Expands(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production"}},
		},
	}
	view := NewTagPolicyView(policy)

	// Initially collapsed
	if view.Model().IsExpanded(0) {
		t.Fatal("expected index 0 to be collapsed initially")
	}

	// Press Enter to expand
	quit := view.HandleKeyPress(KeyEnter)
	if quit {
		t.Error("expected quit=false when pressing Enter")
	}
	if !view.Model().IsExpanded(0) {
		t.Error("expected index 0 to be expanded after Enter")
	}

	// Press Enter again - should stay expanded (Enter only expands)
	view.HandleKeyPress(KeyEnter)
	if !view.Model().IsExpanded(0) {
		t.Error("expected index 0 to stay expanded after second Enter")
	}
}

func TestHandleKeyPress_NewlineKey_Expands(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production"}},
		},
	}
	view := NewTagPolicyView(policy)

	// Press newline to expand
	quit := view.HandleKeyPress(KeyNewline)
	if quit {
		t.Error("expected quit=false when pressing newline")
	}
	if !view.Model().IsExpanded(0) {
		t.Error("expected index 0 to be expanded after newline")
	}
}

func TestHandleKeyPress_SpaceKey_Expands(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production"}},
		},
	}
	view := NewTagPolicyView(policy)

	// Press Space to expand
	quit := view.HandleKeyPress(KeySpace)
	if quit {
		t.Error("expected quit=false when pressing Space")
	}
	if !view.Model().IsExpanded(0) {
		t.Error("expected index 0 to be expanded after Space")
	}

	// Press Space again - should stay expanded (Space only expands)
	view.HandleKeyPress(KeySpace)
	if !view.Model().IsExpanded(0) {
		t.Error("expected index 0 to stay expanded after second Space")
	}
}

func TestHandleKeyPress_LeftKey_Collapses(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production"}},
		},
	}
	view := NewTagPolicyView(policy)

	// First expand using Enter
	view.HandleKeyPress(KeyEnter)
	if !view.Model().IsExpanded(0) {
		t.Fatal("expected index 0 to be expanded after Enter")
	}

	// Press Left to collapse
	quit := view.HandleKeyPress(KeyLeft)
	if quit {
		t.Error("expected quit=false when pressing Left")
	}
	if view.Model().IsExpanded(0) {
		t.Error("expected index 0 to be collapsed after Left")
	}
}

func TestHandleKeyPress_RightKey_Expands(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production"}},
		},
	}
	view := NewTagPolicyView(policy)

	// Initially collapsed
	if view.Model().IsExpanded(0) {
		t.Fatal("expected index 0 to be collapsed initially")
	}

	// Press Right to expand
	quit := view.HandleKeyPress(KeyRight)
	if quit {
		t.Error("expected quit=false when pressing Right")
	}
	if !view.Model().IsExpanded(0) {
		t.Error("expected index 0 to be expanded after Right")
	}
}

func TestHandleKeyPress_ToggleOnSelectedIndex(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment"},
			{Name: "CostCenter"},
			{Name: "Project"},
		},
	}
	view := NewTagPolicyView(policy)

	// Move to index 1
	view.HandleKeyPress(KeyDown)
	if view.Model().SelectedIndex != 1 {
		t.Fatalf("expected SelectedIndex 1, got %d", view.Model().SelectedIndex)
	}

	// Toggle expand on index 1
	view.HandleKeyPress(KeyEnter)
	if !view.Model().IsExpanded(1) {
		t.Error("expected index 1 to be expanded")
	}
	if view.Model().IsExpanded(0) {
		t.Error("expected index 0 to remain collapsed")
	}
	if view.Model().IsExpanded(2) {
		t.Error("expected index 2 to remain collapsed")
	}
}

func TestHandleKeyPress_EmptyPolicy_NavigationNoOp(t *testing.T) {
	view := NewTagPolicyView(&TagPolicy{Tags: []TagKey{}})

	// Down should not crash and should not change index
	quit := view.HandleKeyPress(KeyDown)
	if quit {
		t.Error("expected quit=false when pressing down on empty policy")
	}
	if view.Model().SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to stay at 0, got %d", view.Model().SelectedIndex)
	}

	// Up should not crash and should not change index
	quit = view.HandleKeyPress(KeyUp)
	if quit {
		t.Error("expected quit=false when pressing up on empty policy")
	}
	if view.Model().SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to stay at 0, got %d", view.Model().SelectedIndex)
	}
}

func TestHandleKeyPress_EmptyPolicy_ToggleNoOp(t *testing.T) {
	view := NewTagPolicyView(&TagPolicy{Tags: []TagKey{}})

	// Toggle should not crash on empty policy
	quit := view.HandleKeyPress(KeyEnter)
	if quit {
		t.Error("expected quit=false when pressing Enter on empty policy")
	}
	// No keys to expand, so ExpandedKeys should remain empty
	if len(view.Model().ExpandedKeys) != 0 {
		t.Errorf("expected ExpandedKeys to be empty, got %d entries", len(view.Model().ExpandedKeys))
	}
}

func TestHandleKeyPress_UnknownKey_NoOp(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment"},
			{Name: "CostCenter"},
		},
	}
	view := NewTagPolicyView(policy)

	// Press an unknown key
	quit := view.HandleKeyPress('x')
	if quit {
		t.Error("expected quit=false for unknown key")
	}
	if view.Model().SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to stay at 0, got %d", view.Model().SelectedIndex)
	}
	if len(view.Model().ExpandedKeys) != 0 {
		t.Errorf("expected ExpandedKeys to be empty, got %d entries", len(view.Model().ExpandedKeys))
	}
}

func TestHandleKeyPress_SingleTagKey_Navigation(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment"},
		},
	}
	view := NewTagPolicyView(policy)

	// Down should stay at 0 (only one element)
	view.HandleKeyPress(KeyDown)
	if view.Model().SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to stay at 0 with single key, got %d", view.Model().SelectedIndex)
	}

	// Up should stay at 0
	view.HandleKeyPress(KeyUp)
	if view.Model().SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to stay at 0 with single key, got %d", view.Model().SelectedIndex)
	}
}


// Tests for Render function
// Requirements: 3.1, 3.4, 3.6, 3.7

func TestRender_EmptyPolicy(t *testing.T) {
	view := NewTagPolicyView(&TagPolicy{Tags: []TagKey{}})
	output := view.Render()

	// Should contain header
	if !strings.Contains(output, "AWS Tag Policy") {
		t.Error("expected output to contain 'AWS Tag Policy' header")
	}

	// Should contain empty message
	if !strings.Contains(output, "No tag keys found") {
		t.Error("expected output to contain 'No tag keys found' message")
	}

	// Should contain quit help
	if !strings.Contains(output, "q: quit") {
		t.Error("expected output to contain quit help text")
	}
}

func TestRender_SingleTagKey_Collapsed(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production", "Development"}},
		},
	}
	view := NewTagPolicyView(policy)
	output := view.Render()

	// Should contain header
	if !strings.Contains(output, "AWS Tag Policy") {
		t.Error("expected output to contain 'AWS Tag Policy' header")
	}

	// Should contain tag key name
	if !strings.Contains(output, "Environment") {
		t.Error("expected output to contain 'Environment' tag key")
	}

	// Should contain selection indicator and tag key (with possible color codes between)
	// The output format is: "> ▶ \033[33mEnvironment\033[0m"
	if !strings.Contains(output, "> ▶") || !strings.Contains(output, "Environment") {
		t.Errorf("expected output to contain selection indicator and 'Environment', got:\n%s", output)
	}

	// Should NOT contain values (collapsed)
	if strings.Contains(output, "Production") {
		t.Error("expected output to NOT contain 'Production' when collapsed")
	}
	if strings.Contains(output, "Development") {
		t.Error("expected output to NOT contain 'Development' when collapsed")
	}

	// Should contain help text (arrow keys for navigation)
	if !strings.Contains(output, "↑/↓: navigate") {
		t.Error("expected output to contain navigation help text")
	}
}

func TestRender_SingleTagKey_Expanded(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production", "Development"}},
		},
	}
	view := NewTagPolicyView(policy)
	view.Model().SetExpanded(0, true)
	output := view.Render()

	// Should contain expanded indicator and tag key (with possible color codes)
	if !strings.Contains(output, "> ▼") || !strings.Contains(output, "Environment") {
		t.Errorf("expected output to contain expanded indicator and 'Environment', got:\n%s", output)
	}

	// Should contain values (expanded)
	if !strings.Contains(output, "Production") {
		t.Error("expected output to contain 'Production' when expanded")
	}
	if !strings.Contains(output, "Development") {
		t.Error("expected output to contain 'Development' when expanded")
	}
}

func TestRender_MultipleTagKeys_SelectionIndicator(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production"}},
			{Name: "CostCenter", Values: []string{"Engineering"}},
			{Name: "Project", Values: []string{"Alpha"}},
		},
	}
	view := NewTagPolicyView(policy)

	// First key selected by default - check for selection indicator and tag name
	output := view.Render()
	if !strings.Contains(output, "> ▶") || !strings.Contains(output, "Environment") {
		t.Errorf("expected first key to be selected, got:\n%s", output)
	}
	if !strings.Contains(output, "  ▶ CostCenter") {
		t.Errorf("expected second key to not be selected, got:\n%s", output)
	}

	// Move selection to second key
	view.HandleKeyPress(KeyDown)
	output = view.Render()
	if !strings.Contains(output, "  ▶ Environment") {
		t.Errorf("expected first key to not be selected after down, got:\n%s", output)
	}
	// Second key now selected - has color codes
	if !strings.Contains(output, "> ▶") || !strings.Contains(output, "CostCenter") {
		t.Errorf("expected second key to be selected after down, got:\n%s", output)
	}
}

func TestRender_MixedExpandedCollapsed(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production", "Development"}},
			{Name: "CostCenter", Values: []string{"Engineering", "Marketing"}},
			{Name: "Project", Values: []string{"Alpha"}},
		},
	}
	view := NewTagPolicyView(policy)

	// Expand first and third keys
	view.Model().SetExpanded(0, true)
	view.Model().SetExpanded(2, true)

	output := view.Render()

	// First key should be expanded (selected, so has color codes)
	if !strings.Contains(output, "> ▼") || !strings.Contains(output, "Environment") {
		t.Errorf("expected first key to be expanded, got:\n%s", output)
	}
	if !strings.Contains(output, "Production") {
		t.Error("expected 'Production' to be visible")
	}
	if !strings.Contains(output, "Development") {
		t.Error("expected 'Development' to be visible")
	}

	// Second key should be collapsed
	if !strings.Contains(output, "  ▶ CostCenter") {
		t.Errorf("expected second key to be collapsed, got:\n%s", output)
	}
	if strings.Contains(output, "Engineering") {
		t.Error("expected 'Engineering' to NOT be visible (collapsed)")
	}

	// Third key should be expanded
	if !strings.Contains(output, "  ▼ Project") {
		t.Errorf("expected third key to be expanded, got:\n%s", output)
	}
	if !strings.Contains(output, "Alpha") {
		t.Error("expected 'Alpha' to be visible")
	}
}

func TestRender_TagKeyWithNoValues_Expanded(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "EmptyTag", Values: []string{}},
		},
	}
	view := NewTagPolicyView(policy)
	view.Model().SetExpanded(0, true)

	output := view.Render()

	// Should show expanded indicator even with no values (with color codes for selected)
	if !strings.Contains(output, "> ▼") || !strings.Contains(output, "EmptyTag") {
		t.Errorf("expected expanded indicator for empty tag, got:\n%s", output)
	}
}

func TestRender_AllTagKeysPresent(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment"},
			{Name: "CostCenter"},
			{Name: "Project"},
			{Name: "Owner"},
			{Name: "Application"},
		},
	}
	view := NewTagPolicyView(policy)
	output := view.Render()

	// All tag keys should be present in output
	expectedKeys := []string{"Environment", "CostCenter", "Project", "Owner", "Application"}
	for _, key := range expectedKeys {
		if !strings.Contains(output, key) {
			t.Errorf("expected output to contain tag key '%s', got:\n%s", key, output)
		}
	}
}

func TestRender_ValuesIndented(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment", Values: []string{"Production", "Development", "Staging"}},
		},
	}
	view := NewTagPolicyView(policy)
	view.Model().SetExpanded(0, true)

	output := view.Render()

	// Values should be indented (start with spaces)
	lines := strings.Split(output, "\n")
	foundValues := 0
	for _, line := range lines {
		if strings.Contains(line, "Production") ||
			strings.Contains(line, "Development") ||
			strings.Contains(line, "Staging") {
			// Check that the line starts with indentation
			if !strings.HasPrefix(line, "    ") {
				t.Errorf("expected value line to be indented, got: '%s'", line)
			}
			foundValues++
		}
	}
	if foundValues != 3 {
		t.Errorf("expected to find 3 indented values, found %d", foundValues)
	}
}

func TestRender_HelpTextPresent(t *testing.T) {
	policy := &TagPolicy{
		Tags: []TagKey{
			{Name: "Environment"},
		},
	}
	view := NewTagPolicyView(policy)
	output := view.Render()

	// Should contain all help text elements (arrow keys for navigation)
	if !strings.Contains(output, "↑/↓: navigate") {
		t.Error("expected output to contain '↑/↓: navigate'")
	}
	if !strings.Contains(output, "→: expand") {
		t.Error("expected output to contain '→: expand'")
	}
	if !strings.Contains(output, "←: collapse") {
		t.Error("expected output to contain '←: collapse'")
	}
	if !strings.Contains(output, "q: quit") {
		t.Error("expected output to contain 'q: quit'")
	}
}

func TestRender_NilPolicy(t *testing.T) {
	view := NewTagPolicyView(nil)
	output := view.Render()

	// Should handle nil policy gracefully
	if !strings.Contains(output, "AWS Tag Policy") {
		t.Error("expected output to contain header")
	}
	if !strings.Contains(output, "No tag keys found") {
		t.Error("expected output to contain empty message")
	}
}


// =============================================================================
// Property-Based Tests for TUI View
// =============================================================================

// genViewIdentifier generates valid identifier strings for tag names and values.
// Uses alphanumeric characters starting with a letter to ensure valid identifiers.
func genViewIdentifier() gopter.Gen {
	return gen.Identifier().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 50
	})
}

// genViewStringSlice generates a slice of identifier strings.
func genViewStringSlice() gopter.Gen {
	return gen.SliceOf(genViewIdentifier()).Map(func(slice []string) []string {
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

// genViewTagKey generates a random TagKey with valid name, values, and enforced_for.
func genViewTagKey() gopter.Gen {
	return gopter.CombineGens(
		genViewIdentifier(), // Name
		genViewStringSlice(), // Values
		genViewStringSlice(), // EnforcedFor
	).Map(func(vals []interface{}) TagKey {
		return TagKey{
			Name:        vals[0].(string),
			Values:      vals[1].([]string),
			EnforcedFor: vals[2].([]string),
		}
	})
}

// genViewTagPolicy generates a random TagPolicy with unique tag keys.
func genViewTagPolicy() gopter.Gen {
	return gen.SliceOf(genViewTagKey()).Map(func(keys []TagKey) *TagPolicy {
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

// TestRenderContainsAllTagKeys is a property-based test that verifies:
// For any TagPolicy with N tag keys, the rendered TUI output SHALL contain
// all N tag key names.
//
// Feature: aws-tag-policy, Property 4: Render Output Contains All Tag Keys
// **Validates: Requirements 3.1**
func TestRenderContainsAllTagKeys(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("render output contains all tag key names", prop.ForAll(
		func(policy *TagPolicy) bool {
			view := NewTagPolicyView(policy)
			output := view.Render()

			// Verify each tag key name appears in the output
			for _, tagKey := range policy.Tags {
				if !strings.Contains(output, tagKey.Name) {
					t.Logf("Tag key name '%s' not found in rendered output:\n%s", tagKey.Name, output)
					return false
				}
			}
			return true
		},
		genViewTagPolicy(),
	))

	properties.TestingRun(t)
}


// TestNavigationStateChanges is a property-based test that verifies:
// For any view state with selected index I and N total tag keys:
// - Pressing Down SHALL result in index min(I+1, N-1)
// - Pressing Up SHALL result in index max(I-1, 0)
//
// Feature: aws-tag-policy, Property 5: Navigation State Changes Correctly
// **Validates: Requirements 3.2**
func TestNavigationStateChanges(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	// Property: Down key results in min(I+1, N-1)
	properties.Property("down key results in min(I+1, N-1)", prop.ForAll(
		func(numKeys int, selectedIndex int) bool {
			if numKeys <= 0 {
				return true // Skip empty policies
			}

			// Clamp selectedIndex to valid range [0, numKeys-1]
			if selectedIndex < 0 {
				selectedIndex = 0
			}
			if selectedIndex >= numKeys {
				selectedIndex = numKeys - 1
			}

			// Create policy with numKeys tag keys
			policy := createPolicyWithNKeys(numKeys)
			view := NewTagPolicyView(policy)
			view.Model().SelectedIndex = selectedIndex

			// Press down
			view.HandleKeyPress(KeyDown)

			// Expected: min(I+1, N-1)
			expected := selectedIndex + 1
			if expected > numKeys-1 {
				expected = numKeys - 1
			}

			actual := view.Model().SelectedIndex
			if actual != expected {
				t.Logf("Down key: numKeys=%d, startIndex=%d, expected=%d, actual=%d",
					numKeys, selectedIndex, expected, actual)
				return false
			}
			return true
		},
		gen.IntRange(1, 20),  // numKeys: 1 to 20
		gen.IntRange(0, 19),  // selectedIndex: 0 to 19
	))

	// Property: Up key results in max(I-1, 0)
	properties.Property("up key results in max(I-1, 0)", prop.ForAll(
		func(numKeys int, selectedIndex int) bool {
			if numKeys <= 0 {
				return true // Skip empty policies
			}

			// Clamp selectedIndex to valid range [0, numKeys-1]
			if selectedIndex < 0 {
				selectedIndex = 0
			}
			if selectedIndex >= numKeys {
				selectedIndex = numKeys - 1
			}

			// Create policy with numKeys tag keys
			policy := createPolicyWithNKeys(numKeys)
			view := NewTagPolicyView(policy)
			view.Model().SelectedIndex = selectedIndex

			// Press up
			view.HandleKeyPress(KeyUp)

			// Expected: max(I-1, 0)
			expected := selectedIndex - 1
			if expected < 0 {
				expected = 0
			}

			actual := view.Model().SelectedIndex
			if actual != expected {
				t.Logf("Up key: numKeys=%d, startIndex=%d, expected=%d, actual=%d",
					numKeys, selectedIndex, expected, actual)
				return false
			}
			return true
		},
		gen.IntRange(1, 20),  // numKeys: 1 to 20
		gen.IntRange(0, 19),  // selectedIndex: 0 to 19
	))

	properties.TestingRun(t)
}

// createPolicyWithNKeys creates a TagPolicy with N tag keys for testing.
// Each tag key has a unique name like "TagKey0", "TagKey1", etc.
func createPolicyWithNKeys(n int) *TagPolicy {
	tags := make([]TagKey, n)
	for i := 0; i < n; i++ {
		tags[i] = TagKey{
			Name:   fmt.Sprintf("TagKey%d", i),
			Values: []string{fmt.Sprintf("Value%d", i)},
		}
	}
	return &TagPolicy{Tags: tags}
}

// TestExpandCollapseWithArrowKeys is a property-based test that verifies:
// For any view state and selected tag key:
// - Right arrow (or Enter/Space) SHALL expand the selected tag key
// - Left arrow SHALL collapse the selected tag key
// - Expand then Collapse returns to collapsed state
// - Collapse then Expand returns to expanded state
//
// Feature: aws-tag-policy, Property 6: Expand/Collapse with Arrow Keys
// **Validates: Requirements 3.3**
func TestExpandCollapseWithArrowKeys(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	// Property: Right key expands, Left key collapses - round trip returns to original
	properties.Property("right then left returns to collapsed state", prop.ForAll(
		func(numKeys int, selectedIndex int) bool {
			if numKeys <= 0 {
				return true // Skip empty policies
			}

			// Clamp selectedIndex to valid range [0, numKeys-1]
			if selectedIndex < 0 {
				selectedIndex = 0
			}
			if selectedIndex >= numKeys {
				selectedIndex = numKeys - 1
			}

			// Create policy and view (starts collapsed)
			policy := createPolicyWithNKeys(numKeys)
			view := NewTagPolicyView(policy)
			view.Model().SelectedIndex = selectedIndex

			// Verify initially collapsed
			if view.Model().IsExpanded(selectedIndex) {
				t.Logf("Expected initially collapsed")
				return false
			}

			// Press Right to expand
			view.HandleKeyPress(KeyRight)
			if !view.Model().IsExpanded(selectedIndex) {
				t.Logf("Expected expanded after Right key")
				return false
			}

			// Press Left to collapse
			view.HandleKeyPress(KeyLeft)
			if view.Model().IsExpanded(selectedIndex) {
				t.Logf("Expected collapsed after Left key")
				return false
			}

			return true
		},
		gen.IntRange(1, 20),  // numKeys: 1 to 20
		gen.IntRange(0, 19),  // selectedIndex: 0 to 19
	))

	// Property: Left then Right on expanded item returns to expanded state
	properties.Property("left then right on expanded returns to expanded state", prop.ForAll(
		func(numKeys int, selectedIndex int) bool {
			if numKeys <= 0 {
				return true // Skip empty policies
			}

			// Clamp selectedIndex to valid range [0, numKeys-1]
			if selectedIndex < 0 {
				selectedIndex = 0
			}
			if selectedIndex >= numKeys {
				selectedIndex = numKeys - 1
			}

			// Create policy and view, start expanded
			policy := createPolicyWithNKeys(numKeys)
			view := NewTagPolicyView(policy)
			view.Model().SelectedIndex = selectedIndex
			view.Model().SetExpanded(selectedIndex, true)

			// Verify initially expanded
			if !view.Model().IsExpanded(selectedIndex) {
				t.Logf("Expected initially expanded")
				return false
			}

			// Press Left to collapse
			view.HandleKeyPress(KeyLeft)
			if view.Model().IsExpanded(selectedIndex) {
				t.Logf("Expected collapsed after Left key")
				return false
			}

			// Press Right to expand
			view.HandleKeyPress(KeyRight)
			if !view.Model().IsExpanded(selectedIndex) {
				t.Logf("Expected expanded after Right key")
				return false
			}

			return true
		},
		gen.IntRange(1, 20),  // numKeys: 1 to 20
		gen.IntRange(0, 19),  // selectedIndex: 0 to 19
	))

	// Property: Enter key expands (same as Right)
	properties.Property("Enter key expands selected tag key", prop.ForAll(
		func(numKeys int, selectedIndex int) bool {
			if numKeys <= 0 {
				return true // Skip empty policies
			}

			// Clamp selectedIndex to valid range [0, numKeys-1]
			if selectedIndex < 0 {
				selectedIndex = 0
			}
			if selectedIndex >= numKeys {
				selectedIndex = numKeys - 1
			}

			// Create policy and view (starts collapsed)
			policy := createPolicyWithNKeys(numKeys)
			view := NewTagPolicyView(policy)
			view.Model().SelectedIndex = selectedIndex

			// Press Enter to expand
			view.HandleKeyPress(KeyEnter)
			if !view.Model().IsExpanded(selectedIndex) {
				t.Logf("Expected expanded after Enter key")
				return false
			}

			return true
		},
		gen.IntRange(1, 20),  // numKeys: 1 to 20
		gen.IntRange(0, 19),  // selectedIndex: 0 to 19
	))

	// Property: Space key expands (same as Right)
	properties.Property("Space key expands selected tag key", prop.ForAll(
		func(numKeys int, selectedIndex int) bool {
			if numKeys <= 0 {
				return true // Skip empty policies
			}

			// Clamp selectedIndex to valid range [0, numKeys-1]
			if selectedIndex < 0 {
				selectedIndex = 0
			}
			if selectedIndex >= numKeys {
				selectedIndex = numKeys - 1
			}

			// Create policy and view (starts collapsed)
			policy := createPolicyWithNKeys(numKeys)
			view := NewTagPolicyView(policy)
			view.Model().SelectedIndex = selectedIndex

			// Press Space to expand
			view.HandleKeyPress(KeySpace)
			if !view.Model().IsExpanded(selectedIndex) {
				t.Logf("Expected expanded after Space key")
				return false
			}

			return true
		},
		gen.IntRange(1, 20),  // numKeys: 1 to 20
		gen.IntRange(0, 19),  // selectedIndex: 0 to 19
	))

	properties.TestingRun(t)
}

// TestExpandedKeysShowAllValues is a property-based test that verifies:
// For any expanded tag key with M allowed values, the rendered output
// SHALL contain all M values.
//
// Feature: aws-tag-policy, Property 7: Expanded Keys Show All Values
// **Validates: Requirements 3.4**
func TestExpandedKeysShowAllValues(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	// Property: Expanded tag key shows all its values in rendered output
	properties.Property("expanded tag key shows all its values", prop.ForAll(
		func(tagKey TagKey) bool {
			// Create policy with single tag key
			policy := &TagPolicy{Tags: []TagKey{tagKey}}
			view := NewTagPolicyView(policy)

			// Expand the tag key
			view.Model().SetExpanded(0, true)

			// Render
			output := view.Render()

			// Verify all values are present in output
			for _, value := range tagKey.Values {
				if !strings.Contains(output, value) {
					t.Logf("Value '%s' not found in rendered output for tag key '%s':\n%s",
						value, tagKey.Name, output)
					return false
				}
			}
			return true
		},
		genViewTagKey(),
	))

	// Property: Multiple expanded tag keys show all their values
	properties.Property("multiple expanded tag keys show all their values", prop.ForAll(
		func(policy *TagPolicy) bool {
			if len(policy.Tags) == 0 {
				return true // Skip empty policies
			}

			view := NewTagPolicyView(policy)

			// Expand all tag keys
			for i := range policy.Tags {
				view.Model().SetExpanded(i, true)
			}

			// Render
			output := view.Render()

			// Verify all values from all tag keys are present
			for _, tagKey := range policy.Tags {
				for _, value := range tagKey.Values {
					if !strings.Contains(output, value) {
						t.Logf("Value '%s' from tag key '%s' not found in rendered output:\n%s",
							value, tagKey.Name, output)
						return false
					}
				}
			}
			return true
		},
		genViewTagPolicy(),
	))

	// Property: Only expanded tag keys show their values (collapsed keys hide values)
	properties.Property("only expanded tag keys show their values", prop.ForAll(
		func(numKeys int, expandedIndex int) bool {
			if numKeys <= 1 {
				return true // Need at least 2 keys to test this property
			}

			// Clamp expandedIndex to valid range
			if expandedIndex < 0 {
				expandedIndex = 0
			}
			if expandedIndex >= numKeys {
				expandedIndex = numKeys - 1
			}

			// Create policy with unique values per tag key
			tags := make([]TagKey, numKeys)
			for i := 0; i < numKeys; i++ {
				tags[i] = TagKey{
					Name:   fmt.Sprintf("TagKey%d", i),
					Values: []string{fmt.Sprintf("UniqueValue%d", i)},
				}
			}
			policy := &TagPolicy{Tags: tags}
			view := NewTagPolicyView(policy)

			// Expand only one tag key
			view.Model().SetExpanded(expandedIndex, true)

			// Render
			output := view.Render()

			// Verify expanded key's values are present
			expandedValue := fmt.Sprintf("UniqueValue%d", expandedIndex)
			if !strings.Contains(output, expandedValue) {
				t.Logf("Expanded value '%s' not found in output:\n%s", expandedValue, output)
				return false
			}

			// Verify collapsed keys' values are NOT present
			for i := 0; i < numKeys; i++ {
				if i == expandedIndex {
					continue // Skip the expanded key
				}
				collapsedValue := fmt.Sprintf("UniqueValue%d", i)
				if strings.Contains(output, collapsedValue) {
					t.Logf("Collapsed value '%s' should not be in output:\n%s", collapsedValue, output)
					return false
				}
			}
			return true
		},
		gen.IntRange(2, 10),  // numKeys: 2 to 10
		gen.IntRange(0, 9),   // expandedIndex: 0 to 9
	))

	properties.TestingRun(t)
}


// TestRenderReflectsViewState is a property-based test that verifies:
// For any view state:
// - The rendered output SHALL contain a selection indicator at the currently selected index
// - Each tag key SHALL display an expansion indicator matching its expanded/collapsed state
//
// Feature: aws-tag-policy, Property 8: Render Reflects View State
// **Validates: Requirements 3.6, 3.7**
func TestRenderReflectsViewState(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	// Property: Selection indicator appears at selected index
	properties.Property("selection indicator at selected index", prop.ForAll(
		func(numKeys int, selectedIndex int) bool {
			if numKeys <= 0 {
				return true // Skip empty policies
			}

			// Clamp selectedIndex to valid range [0, numKeys-1]
			if selectedIndex < 0 {
				selectedIndex = 0
			}
			if selectedIndex >= numKeys {
				selectedIndex = numKeys - 1
			}

			// Create policy and view
			policy := createPolicyWithNKeys(numKeys)
			view := NewTagPolicyView(policy)
			view.Model().SelectedIndex = selectedIndex

			output := view.Render()
			lines := strings.Split(output, "\n")

			// Find lines that contain tag keys (they have expansion indicators)
			tagKeyLineIndex := 0
			for _, line := range lines {
				// Skip header, empty lines, help text, and value lines
				if strings.Contains(line, "AWS Tag Policy") ||
					strings.Contains(line, "No tag keys found") ||
					strings.Contains(line, "↑/↓: navigate") ||
					strings.Contains(line, "q: quit") ||
					line == "" {
					continue
				}

				// Check if this is a tag key line (contains expansion indicator)
				isTagKeyLine := strings.Contains(line, ExpandedIndicator) || strings.Contains(line, CollapsedIndicator)
				if !isTagKeyLine {
					continue // This is a value line (indented)
				}

				// Check selection indicator
				hasSelectionIndicator := strings.HasPrefix(line, SelectionIndicator+" ")
				hasNoSelectionIndicator := strings.HasPrefix(line, NoSelectionIndicator+" ")

				if tagKeyLineIndex == selectedIndex {
					// This line should have the selection indicator
					if !hasSelectionIndicator {
						t.Logf("Expected selection indicator at index %d, line: '%s'", selectedIndex, line)
						return false
					}
				} else {
					// This line should NOT have the selection indicator
					if !hasNoSelectionIndicator {
						t.Logf("Expected no selection indicator at index %d (selected=%d), line: '%s'",
							tagKeyLineIndex, selectedIndex, line)
						return false
					}
				}

				tagKeyLineIndex++
			}

			// Verify we found all tag keys
			if tagKeyLineIndex != numKeys {
				t.Logf("Expected %d tag key lines, found %d", numKeys, tagKeyLineIndex)
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(0, 9),
	))

	// Property: Expansion indicators match expanded state
	properties.Property("expansion indicators match expanded state", prop.ForAll(
		func(numKeys int, expandedIndices []int) bool {
			if numKeys <= 0 {
				return true // Skip empty policies
			}

			// Create policy and view
			policy := createPolicyWithNKeys(numKeys)
			view := NewTagPolicyView(policy)

			// Build a set of expanded indices (clamped to valid range)
			expandedSet := make(map[int]bool)
			for _, idx := range expandedIndices {
				if idx >= 0 && idx < numKeys {
					expandedSet[idx] = true
					view.Model().SetExpanded(idx, true)
				}
			}

			output := view.Render()
			lines := strings.Split(output, "\n")

			// Find lines that contain tag keys and verify expansion indicators
			tagKeyLineIndex := 0
			for _, line := range lines {
				// Skip header, empty lines, help text, and value lines
				if strings.Contains(line, "AWS Tag Policy") ||
					strings.Contains(line, "No tag keys found") ||
					strings.Contains(line, "↑/↓: navigate") ||
					strings.Contains(line, "q: quit") ||
					line == "" {
					continue
				}

				// Check if this is a tag key line (contains expansion indicator)
				hasExpandedIndicator := strings.Contains(line, ExpandedIndicator)
				hasCollapsedIndicator := strings.Contains(line, CollapsedIndicator)

				if !hasExpandedIndicator && !hasCollapsedIndicator {
					continue // This is a value line (indented)
				}

				// Verify the correct indicator is present
				isExpanded := expandedSet[tagKeyLineIndex]
				if isExpanded {
					if !hasExpandedIndicator {
						t.Logf("Expected expanded indicator (▼) at index %d, line: '%s'", tagKeyLineIndex, line)
						return false
					}
					if hasCollapsedIndicator {
						t.Logf("Unexpected collapsed indicator (▶) at expanded index %d, line: '%s'", tagKeyLineIndex, line)
						return false
					}
				} else {
					if !hasCollapsedIndicator {
						t.Logf("Expected collapsed indicator (▶) at index %d, line: '%s'", tagKeyLineIndex, line)
						return false
					}
					if hasExpandedIndicator {
						t.Logf("Unexpected expanded indicator (▼) at collapsed index %d, line: '%s'", tagKeyLineIndex, line)
						return false
					}
				}

				tagKeyLineIndex++
			}

			// Verify we found all tag keys
			if tagKeyLineIndex != numKeys {
				t.Logf("Expected %d tag key lines, found %d", numKeys, tagKeyLineIndex)
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
		gen.SliceOf(gen.IntRange(0, 9)),
	))

	// Property: Combined selection and expansion indicators are correct
	properties.Property("combined selection and expansion indicators are correct", prop.ForAll(
		func(numKeys int, selectedIndex int, expandedIndices []int) bool {
			if numKeys <= 0 {
				return true // Skip empty policies
			}

			// Clamp selectedIndex to valid range [0, numKeys-1]
			if selectedIndex < 0 {
				selectedIndex = 0
			}
			if selectedIndex >= numKeys {
				selectedIndex = numKeys - 1
			}

			// Create policy and view
			policy := createPolicyWithNKeys(numKeys)
			view := NewTagPolicyView(policy)
			view.Model().SelectedIndex = selectedIndex

			// Build a set of expanded indices (clamped to valid range)
			expandedSet := make(map[int]bool)
			for _, idx := range expandedIndices {
				if idx >= 0 && idx < numKeys {
					expandedSet[idx] = true
					view.Model().SetExpanded(idx, true)
				}
			}

			output := view.Render()
			lines := strings.Split(output, "\n")

			// Find lines that contain tag keys and verify both indicators
			tagKeyLineIndex := 0
			for _, line := range lines {
				// Skip header, empty lines, help text, and value lines
				if strings.Contains(line, "AWS Tag Policy") ||
					strings.Contains(line, "No tag keys found") ||
					strings.Contains(line, "↑/↓: navigate") ||
					strings.Contains(line, "q: quit") ||
					line == "" {
					continue
				}

				// Check if this is a tag key line (contains expansion indicator)
				hasExpandedIndicator := strings.Contains(line, ExpandedIndicator)
				hasCollapsedIndicator := strings.Contains(line, CollapsedIndicator)

				if !hasExpandedIndicator && !hasCollapsedIndicator {
					continue // This is a value line (indented)
				}

				// Determine expected indicators
				isSelected := tagKeyLineIndex == selectedIndex
				isExpanded := expandedSet[tagKeyLineIndex]

				// Build expected prefix
				var expectedPrefix string
				if isSelected {
					expectedPrefix = SelectionIndicator + " "
				} else {
					expectedPrefix = NoSelectionIndicator + " "
				}
				if isExpanded {
					expectedPrefix += ExpandedIndicator + " "
				} else {
					expectedPrefix += CollapsedIndicator + " "
				}

				// Verify the line starts with the expected prefix
				if !strings.HasPrefix(line, expectedPrefix) {
					t.Logf("Expected prefix '%s' at index %d (selected=%v, expanded=%v), line: '%s'",
						expectedPrefix, tagKeyLineIndex, isSelected, isExpanded, line)
					return false
				}

				tagKeyLineIndex++
			}

			// Verify we found all tag keys
			if tagKeyLineIndex != numKeys {
				t.Logf("Expected %d tag key lines, found %d", numKeys, tagKeyLineIndex)
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(0, 9),
		gen.SliceOf(gen.IntRange(0, 9)),
	))

	properties.TestingRun(t)
}
