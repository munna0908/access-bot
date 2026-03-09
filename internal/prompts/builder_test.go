package prompts

import (
	"strings"
	"testing"
)

func TestBuildUserPrompt(t *testing.T) {
	tests := []struct {
		name             string
		question         string
		categoryContents []CategoryContent
		wantContains     []string
	}{
		{
			name:     "single category",
			question: "What should I eat?",
			categoryContents: []CategoryContent{
				{Category: "FOOD", Content: "# Food Preferences\nVegetarian"},
			},
			wantContains: []string{
				"User Question:",
				"What should I eat?",
				"## FOOD",
				"# Food Preferences",
				"Vegetarian",
				"Answer the question using only the context provided",
			},
		},
		{
			name:     "multiple categories",
			question: "Order dinner",
			categoryContents: []CategoryContent{
				{Category: "FOOD", Content: "# Food\nVegetarian"},
				{Category: "HEALTH", Content: "# Health\nDiabetes"},
			},
			wantContains: []string{
				"## FOOD",
				"## HEALTH",
				"# Food",
				"Vegetarian",
				"# Health",
				"Diabetes",
			},
		},
		{
			name:             "no categories",
			question:         "Hello",
			categoryContents: []CategoryContent{},
			wantContains: []string{
				"User Question:",
				"Hello",
				"Participant Context:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildUserPrompt(tt.question, tt.categoryContents)
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("expected prompt to contain %q, got:\n%s", want, result)
				}
			}
		})
	}
}

func TestGetSystemPrompt(t *testing.T) {
	prompt := GetSystemPrompt()
	if prompt == "" {
		t.Error("expected non-empty system prompt")
	}

	expectedParts := []string{
		"permission-controlled participant context system",
		"Only use the context provided",
		"Sorry, I can't answer that",
	}

	for _, part := range expectedParts {
		if !strings.Contains(prompt, part) {
			t.Errorf("expected system prompt to contain %q", part)
		}
	}
}
