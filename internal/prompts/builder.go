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
- If the context is insufficient, respond exactly: Sorry, I can't answer that.
- Be concise and helpful.

IMPORTANT — Food ordering requests:
When the question asks to order food or list food options, respond with ONLY a numbered list of exactly 3 suitable dishes based on the participant's food preferences and health restrictions. Use this exact format with no other text:
1. [Dish Name]
2. [Dish Name]
3. [Dish Name]

IMPORTANT — Delivery address requests:
The question will tell you exactly which address type to return (e.g. "my work address", "my home address", "my gym address").
Find that address in the context and respond with ONLY the full postal address on a single line: building/company name, flat/floor, street, area, city, state and PIN code. No gate codes, no delivery notes, no explanation.`

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
