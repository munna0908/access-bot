package prompts

import (
	"fmt"
	"strings"
)

// SystemPrompt is the system instruction for the LLM.
const SystemPrompt = `You are answering a question inside a permission-controlled participant context system.

You will receive:
1. A user question
2. Markdown files containing participant context

Rules:
- Only use the context provided.
- Do not assume additional participant data.
- Do not fabricate missing information.
- If the context is insufficient, respond exactly:
Sorry, I can't answer that.
- Be concise and helpful.`

// CategoryContent holds the category name and its markdown content.
type CategoryContent struct {
	Category string
	Content  string
}

// BuildUserPrompt constructs the user prompt with the question and context.
func BuildUserPrompt(question string, categoryContents []CategoryContent) string {
	var sb strings.Builder

	sb.WriteString("User Question:\n")
	sb.WriteString(question)
	sb.WriteString("\n\nParticipant Context:\n")

	for _, cc := range categoryContents {
		sb.WriteString(fmt.Sprintf("\n## %s\n", cc.Category))
		sb.WriteString(cc.Content)
		sb.WriteString("\n")
	}

	sb.WriteString("\nAnswer the question using only the context provided.\n")
	sb.WriteString("If the context is insufficient say exactly:\n")
	sb.WriteString("Sorry, I can't answer that.")

	return sb.String()
}

// GetSystemPrompt returns the system prompt for the LLM.
func GetSystemPrompt() string {
	return SystemPrompt
}
