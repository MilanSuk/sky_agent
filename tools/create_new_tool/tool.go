package main

import (
	"fmt"
	"strings"
)

// Create new tool from description.
type create_new_tool struct {
	Name        string //tool name. No spaces(use '_' instead) or special characters.
	Description string //Prompt with the name of tool, parameters(name, type, description) and detail description of functionality.
}

func (st *create_new_tool) run() string {
	SystemPrompt := "You are an AI programming assistant, who enjoys precision and carefully follows the user's requirements. You write code in Go-lang."

	UserPrompt := "This is prompt from user:\n"
	UserPrompt += st.Description
	UserPrompt += "\n"

	UserPrompt += "Based on this prompt modify this template:"
	UserPrompt += "```go\n"
	UserPrompt += fmt.Sprintf(`package main
//<tool_description>
type %s struct {
	<tool_input_parameters_with_descriptions_as_comments>
}
func (st *%s) run() <one_tool_return_type> {
	<tool_implementation>	//If there is error, use log.Fatalf
}`, st.Name, st.Name)
	UserPrompt += "\n```"

	UserPrompt += "\n"
	UserPrompt += "If an error occurs, use log.Fatalf. Output only modified template above. Implement everything, no placeholders! Don't add main() function to the code."

	fmt.Println("create_new_tool UserPrompt:", UserPrompt)

	code_answer := SDK_RunAgent("coder", 20, 20000, SystemPrompt, UserPrompt)

	var ok bool
	code_answer, ok = strings.CutPrefix(code_answer, "```go")
	if ok {
		code_answer, ok = strings.CutSuffix(code_answer, "```")
		if ok {
			compile_answer := SDK_SetToolCode(st.Name, code_answer)
			if compile_answer == "" {
				return "success"
			} else {
				return compile_answer //error
			}
		}
	}

	return "failed"
}
