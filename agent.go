/*
Copyright 2025 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Agent struct {
	server    *NetServer
	passwords *Passwords

	Folder string
	Model  string

	Anthropic_props Anthropic_completion_props
	OpenAI_props    OpenAI_completion_props

	InputTokens  int
	OutputTokens int
	TotalTokens  int
	TotalTime    float64

	Sandbox_violations []string
}

func NewAgent(folder string, use_case string, systemPrompt string, userPrompt string, server *NetServer, passwords *Passwords) *Agent {

	if systemPrompt == "" {
		systemPrompt = `You are an AI tool calling assistant, who enjoys precision and carefully follows the user's requirements."

If you can not find the right tool, use the tool 'create_new_tool'. If there is some problem with a tool(for example, a bug) then use the tool 'update_tool'.
Don't ask to use, change, or create a tool, just do it! Call tools sequentially. Avoid tool call as parameter value.

User informations, device settings and files are store on the disk, read them with 'access_disk' tool.
If the variable was read from disk and it's changed, you should probably write it back with 'access_disk' tool.

Tool parameter values must be real, don't use placeholder(aka example.com) and don't make up numbers or strings! Use 'access_disk' tool to search for the value.
`
	}

	model := Service_findModelFromUse_cases(use_case)
	agent := &Agent{Folder: folder, Model: model, server: server, passwords: passwords}

	if agent.IsModelAnthropic() {
		agent.Anthropic_props.ResetDefault()
		agent.Anthropic_props.Model = model

		agent.Anthropic_props.System = systemPrompt

		{
			msg := Anthropic_completion_msg{Role: "user"}
			msg.AddText(userPrompt)
			agent.Anthropic_props.Messages = append(agent.Anthropic_props.Messages, msg)
		}

	} else {
		if strings.ToLower(use_case) == "search" {
			agent.OpenAI_props.ResetSearch()
		} else {
			agent.OpenAI_props.ResetDefault()
		}
		agent.OpenAI_props.Model = model

		{
			msg := OpenAI_completion_msgPlain{Role: "system", Content: systemPrompt}
			agent.OpenAI_props.Messages = append(agent.OpenAI_props.Messages, msg)
		}

		{
			msg := OpenAI_completion_msg{Role: "user"}
			msg.AddText(userPrompt)
			agent.OpenAI_props.Messages = append(agent.OpenAI_props.Messages, msg)
		}
	}

	toolList, err := GetToolsList(folder)
	if err != nil {
		log.Fatal(err)
	}

	for _, toolName := range toolList {
		path := filepath.Join(folder, toolName)
		if NeedCompileTool(path) {
			err := CompileTool(path)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
		agent.AddTool(path)
	}

	return agent
}

func (agent *Agent) IsModelAnthropic() bool {
	service := Service_findService(agent.Model)
	return service.Anthropic_completion_url != ""
}

func (agent *Agent) Open(path string) error {
	js, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(js, agent)
	if err != nil {
		return err
	}

	return nil
}
func (agent *Agent) Save(save_as_last bool) error {
	js, err := json.MarshalIndent(agent, "", "")
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%d.json", time.Now().UnixMicro()), js, 0644)
	if err != nil {
		return err
	}

	if save_as_last {
		err = os.WriteFile("last.json", js, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func (agent *Agent) AddTool(tool string) {
	if NeedCompileTool(tool) { //must be compiled
		fmt.Printf("Tool '%s' can't be add because it's not compiled\n", tool)
		return
	}

	openaiAPI, anthropicAPI, err := ConvertFileIntoTool(tool)
	if err != nil {
		log.Fatal(err)
	}

	toolName := filepath.Base(tool)
	if agent.IsModelAnthropic() {
		//update
		for i, tool := range agent.Anthropic_props.Tools {
			if tool.Name == toolName {
				agent.Anthropic_props.Tools[i] = anthropicAPI
				return
			}
		}
		//add
		agent.Anthropic_props.Tools = append(agent.Anthropic_props.Tools, anthropicAPI)

	} else {
		//update
		for i, tool := range agent.OpenAI_props.Tools {
			if tool.Function.Name == toolName {
				agent.OpenAI_props.Tools[i] = openaiAPI
				return
			}
		}
		//add
		agent.OpenAI_props.Tools = append(agent.OpenAI_props.Tools, openaiAPI)
	}
}

func (agent *Agent) GetFinalMessage() string {
	if agent.IsModelAnthropic() {
		if len(agent.Anthropic_props.Messages) > 0 {
			msgs := agent.Anthropic_props.Messages[len(agent.Anthropic_props.Messages)-1].Content
			if len(msgs) > 0 {
				return msgs[0].Text
			}
		}

	} else {
		if len(agent.OpenAI_props.Messages) > 0 {
			msg := agent.OpenAI_props.Messages[len(agent.OpenAI_props.Messages)-1]
			switch it := msg.(type) {
			case OpenAI_completion_msg:
				if len(it.Content) > 0 {
					return it.Content[0].Text
				}
			case OpenAI_completion_msgCalls:
				if len(it.Content) > 0 {
					return it.Content
				}
			case OpenAI_completion_msgResult:
				if len(it.Content) > 0 {
					return it.Content
				}
			}
		}
	}
	return ""
}

func (agent *Agent) PrintStats() {
	fmt.Println("---Stats---")

	if agent.IsModelAnthropic() {
		fmt.Println("Model:", agent.Anthropic_props.Model)
		fmt.Println("#Messages:", len(agent.Anthropic_props.Messages))
	} else {
		fmt.Println("Model:", agent.OpenAI_props.Model)
		fmt.Println("#Messages:", len(agent.OpenAI_props.Messages))
	}

	fmt.Println("Tokens(in, out):", agent.InputTokens, agent.OutputTokens)

	if agent.TotalTime > 0 {
		fmt.Println("Toks/sec:", float64(agent.OutputTokens)/agent.TotalTime)
	}

	model := Service_findModel(agent.Model)
	if model != nil {
		input_price := model.Input_price / 1000000
		output_price := model.Output_price / 1000000

		price := float64(agent.InputTokens)*input_price + float64(agent.OutputTokens)*output_price
		fmt.Printf("Price: $%f\n", price)
		fmt.Printf("Runs per $1: %dx\n", int(1/price))
	}

	fmt.Println("Sandbox violations:", len(agent.Sandbox_violations))
	for _, it := range agent.Sandbox_violations {
		fmt.Printf("- %s\n", it)
	}

	fmt.Println("--- ---")
}

func (agent *Agent) Run() bool {
	service := Service_findService(agent.Model)
	if service == nil {
		log.Fatal(fmt.Errorf("model %s not found. Edit g_services", agent.Model))
	}

	if service.Api_key == "<your_api_key>" {
		log.Fatal(fmt.Errorf("no api_key for service '%s'", service.Name))
	}

	if agent.IsModelAnthropic() {
		startTime := float64(time.Now().UnixMilli()) / 1000

		out, err := Anthropic_completion_Run(agent.Anthropic_props, service.Anthropic_completion_url, service.Api_key)
		if err != nil {
			log.Fatal(err)
		}

		dt := (float64(time.Now().UnixMilli()) / 1000) - startTime

		content := ""
		var tool_calls []OpenAI_completion_msg_Content_ToolCall
		if len(out.Content) > 0 {
			for _, it := range out.Content {
				switch it.Type {
				case "text":
					content += it.Text

				case "tool_use":
					args, err := it.Input.MarshalJSON()
					if err != nil {
						log.Fatal(err)
					}
					fn := OpenAI_completion_msg_Content_ToolCall_Function{Name: it.Name, Arguments: string(args)}
					tool_calls = append(tool_calls, OpenAI_completion_msg_Content_ToolCall{Id: it.Id, Type: it.Type, Function: fn})
				}
			}
		}

		agent.InputTokens += out.Usage.Input_tokens
		agent.OutputTokens += out.Usage.Output_tokens
		agent.TotalTokens += out.Usage.Input_tokens + out.Usage.Output_tokens
		agent.TotalTime += dt

		fmt.Printf("+LLM(%s) generated %dtoks which took %.1fsec = %.1f toks/sec\n", agent.Folder, out.Usage.Output_tokens, dt, float64(out.Usage.Output_tokens)/dt)
		fmt.Printf("+LLM(%s) returns content: %s\n", agent.Folder, content)
		fmt.Printf("+LLM(%s) returns tool_calls: %v\n", agent.Folder, tool_calls)

		msg := Anthropic_completion_msg{Role: "assistant", Content: out.Content}
		agent.Anthropic_props.Messages = append(agent.Anthropic_props.Messages, msg)

		agent.callTools(tool_calls)
		return len(tool_calls) > 0

	} else {
		startTime := float64(time.Now().UnixMilli()) / 1000

		out, err := OpenAI_completion_Run(agent.OpenAI_props, service.OpenAI_completion_url, service.Api_key)
		if err != nil {
			log.Fatal(err)
		}

		dt := (float64(time.Now().UnixMilli()) / 1000) - startTime

		agent.InputTokens += out.Usage.Prompt_tokens
		agent.OutputTokens += out.Usage.Completion_tokens
		agent.TotalTokens += out.Usage.Total_tokens
		agent.TotalTime += dt

		var content string
		var tool_calls []OpenAI_completion_msg_Content_ToolCall
		if len(out.Choices) > 0 {
			content = out.Choices[0].Message.Content
			tool_calls = out.Choices[0].Message.Tool_calls
		}

		fmt.Printf("+LLM(%s) generated %dtoks which took %.1fsec = %.1f toks/sec\n", agent.Folder, out.Usage.Completion_tokens, dt, float64(out.Usage.Completion_tokens)/dt)
		fmt.Printf("+LLM(%s) returns content: %s\n", agent.Folder, content)
		fmt.Printf("+LLM(%s) returns tool_calls: %v\n", agent.Folder, tool_calls)

		msg := OpenAI_completion_msgCalls{Role: "assistant"}
		{
			cwCitations := content
			if len(out.Citations) > 0 {
				cwCitations += "\nCitations:\n"
				for _, ct := range out.Citations {
					cwCitations += ct + "\n"
				}
			}
			msg.Content = cwCitations
		}
		msg.Tool_calls = tool_calls
		agent.OpenAI_props.Messages = append(agent.OpenAI_props.Messages, msg)

		agent.callTools(tool_calls)
		return len(tool_calls) > 0
	}
	//return false
}

func (agent *Agent) RunLoop(max_iters int, max_tokens int) {
	orig_max_iters := max_iters
	orig_max_tokens := max_tokens

	if max_iters <= 0 {
		max_iters = 1000000000 //1B
	}
	if max_tokens <= 0 {
		max_tokens = 1000000000 //1B
	}

	for max_iters > 0 {
		if !agent.Run() {
			return
		}

		if agent.TotalTokens >= max_tokens {
			fmt.Printf("Warning: Agent reached max tokens(%d)\n", orig_max_tokens)
			return
		}

		max_iters--
	}

	fmt.Printf("Warning: Agent reached max iters(%d)\n", orig_max_iters)
}

func (agent *Agent) callTools(tool_calls []OpenAI_completion_msg_Content_ToolCall) {
	if len(agent.Anthropic_props.Messages) > 0 {
		for _, it := range tool_calls {
			for _, tool := range agent.Anthropic_props.Tools {
				if tool.Name == it.Function.Name {

					//call
					answerJs := agent.callTool(it.Function.Name, it.Function.Arguments)

					//save answer
					msg := Anthropic_completion_msg{Role: "user"}
					msg.AddToolResult(it.Id, answerJs)
					//msg.AddImage()
					agent.Anthropic_props.Messages = append(agent.Anthropic_props.Messages, msg)

					fmt.Println("+Tool returns:", msg)
				}
			}
		}
	} else if len(agent.OpenAI_props.Messages) > 0 {
		for _, it := range tool_calls {
			for _, tool := range agent.OpenAI_props.Tools {
				if tool.Function.Name == it.Function.Name {

					//call
					answerJs := agent.callTool(it.Function.Name, it.Function.Arguments)

					//save answer
					msg := OpenAI_completion_msgResult{Role: "tool"}
					msg.Name = tool.Function.Name
					msg.Tool_call_id = it.Id
					msg.Content = string(answerJs)
					//msg.AddText(string(js))
					//msg.AddImage()
					agent.OpenAI_props.Messages = append(agent.OpenAI_props.Messages, msg)

					fmt.Println("+Tool returns:", msg)
				}
			}
		}
	}
}

func (agent *Agent) callTool(toolName string, arguments string) string {
	tool := filepath.Join(agent.Folder, toolName)

	//call
	binPath := filepath.Join(tool, "bin")
	cmd := exec.Command("./"+binPath, strconv.Itoa(agent.server.port))
	cmd.Dir = ""
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		fmt.Println("Error:", err)
	}

	cl, err := agent.server.Accept()
	if err != nil {
		fmt.Println("Error:", err)
	}
	err = cl.WriteArray([]byte(arguments))
	if err != nil {
		fmt.Println("Error:", err)
	}

	var js []byte
	var tp uint64
	for tp != 1 {
		tp, err = cl.ReadInt()
		if err != nil {
			break
		}

		switch tp {
		case 1: //result
			js, _ = cl.ReadArray()

		case 2: //SDK_RunAgent
			max_iters, _ := cl.ReadInt()
			max_tokens, _ := cl.ReadInt()
			use_cases, _ := cl.ReadArray()
			systemPrompt, _ := cl.ReadArray()
			userPrompt, _ := cl.ReadArray()

			//init
			agent2 := NewAgent(tool, string(use_cases), string(systemPrompt), string(userPrompt), agent.server, agent.passwords)
			defer agent2.Save(false)

			//run
			agent2.RunLoop(int(max_iters), int(max_tokens))

			//send result back
			cl.WriteArray([]byte(agent2.GetFinalMessage()))
			agent2.PrintStats()

		case 3: //SDK_SetToolCode
			toolName, _ := cl.ReadArray()
			toolCode, _ := cl.ReadArray()

			path := filepath.Join(tool, string(toolName))
			os.MkdirAll(path, os.ModePerm)
			err := os.WriteFile(filepath.Join(path, "tool.go"), toolCode, 0644)
			if err != nil {
				fmt.Println(err)
			}

			err = CompileTool(path)
			if err == nil {
				//ok
				cl.WriteArray(nil)
			} else {
				//error
				cl.WriteArray([]byte(fmt.Sprintf("Tool '%s' was created, but compiler reported error: %v", path, err)))
			}

			if agent != nil {
				agent.AddTool(path)
			}

		case 4: //SDK_Sandbox_violation
			info, _ := cl.ReadArray()
			if agent != nil {
				agent.Sandbox_violations = append(agent.Sandbox_violations, string(info))
				fmt.Println("Sandbox violation:", string(info))
			}
			cl.WriteInt(1) //block it

		case 5: //SDK_GetPassword
			id, _ := cl.ReadArray()
			if agent != nil {
				password := agent.passwords.Find(string(id))
				cl.WriteArray([]byte(password))
				fmt.Println("Search for password:", string(id))
			}
			cl.WriteInt(1) //block it

		}
	}

	err = cmd.Wait()
	if err != nil {
		//tool crashed
		js = []byte(fmt.Sprintf("Tool '%s' crashed with log.Fatal: %s", tool, err.Error()))
	}

	return string(js)
}
