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
	"strconv"
	"strings"
	"time"
)

type Agent struct {
	Model string

	Props OpenAI_completion_props

	InputTokens  int
	OutputTokens int
	TotalTokens  int
	TotalTime    float64

	Sandbox_violations []string
}

func NewAgent(use_case string, systemPrompt string, userPrompt string) *Agent {
	if systemPrompt == "" {
		systemPrompt = "You are an AI programming assistant, who enjoys precision and carefully follows the user's requirements. Take advantage of function(tool) calling, they are very helpfull! If you can't find right function(tool) then use function 'create_new_tool'. If there is some problem with tool(for example bug) then use function 'update_tool'. If the user message mentioning file, you probably need to use(or create) tool to work with the file."
	}

	model := Service_findModelFromUse_cases(use_case)

	agent := &Agent{Model: model}
	if strings.ToLower(use_case) == "search" {
		agent.Props.ResetSearch()
	} else {
		agent.Props.ResetDefault()
	}
	agent.Props.Model = model

	msg := OpenAI_completion_msg{Role: "system"}
	msg.AddText(systemPrompt)
	agent.Props.Messages = append(agent.Props.Messages, msg)

	msg = OpenAI_completion_msg{Role: "user"}
	msg.AddText(userPrompt)
	agent.Props.Messages = append(agent.Props.Messages, msg)

	return agent
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

func (agent *Agent) AddTool(toolName string) {
	if NeedCompileTool(toolName) { //must be compiled
		fmt.Printf("Tool '%s' can't be add because it's not compiled\n", toolName)
		return
	}

	api, err := ConvertFileIntoTool(toolName)
	if err != nil {
		log.Fatal(err)
	}

	//update
	for i, tool := range agent.Props.Tools {
		if tool.Function.Name == toolName {
			agent.Props.Tools[i] = api
			return
		}
	}

	//add
	agent.Props.Tools = append(agent.Props.Tools, api)
}

func (agent *Agent) GetFinalMessage() string {
	if len(agent.Props.Messages) > 0 {
		msgs := agent.Props.Messages[len(agent.Props.Messages)-1].Content
		if len(msgs) > 0 {
			return msgs[0].Text
		}
	}
	return ""
}

func (agent *Agent) PrintStats() {
	fmt.Println("---Stats---")

	fmt.Println("Model:", agent.Props.Model)

	fmt.Println("#Messages:", len(agent.Props.Messages))

	fmt.Println("Tokens(in, out):", agent.InputTokens, agent.OutputTokens)

	if agent.TotalTime > 0 {
		fmt.Println("Toks/sec:", float64(agent.OutputTokens)/agent.TotalTime)
	}

	model := Service_findModel(agent.Model)
	if model != nil {
		input_price := model.Input_price / 1000000
		output_price := model.Output_price / 1000000

		fmt.Println("Price($):", float64(agent.InputTokens)*input_price+float64(agent.OutputTokens)*output_price)
	}

	fmt.Println("--- ---")
}

func (agent *Agent) Run(server *NetServer) bool {
	var out OpenAI_completion_out

	jsProps, err := json.Marshal(agent.Props)
	if err != nil {
		log.Fatal(err)
	}

	service := Service_findService(agent.Model)
	if service == nil {
		log.Fatal(fmt.Errorf("model %s not found. Edit g_services", agent.Model))
	}

	if service.Api_key == "<your_api_key>" {
		log.Fatal(fmt.Errorf("no api_key for service '%s'", service.Name))
	}

	out, err = OpenAI_completion_Run(jsProps, service.Completion_url, service.Api_key)
	if err != nil {
		log.Fatal(err)
	}

	agent.InputTokens += out.inputTokens
	agent.OutputTokens += out.outputTokens
	agent.TotalTokens += out.totalTokens
	agent.TotalTime += out.time

	msg := OpenAI_completion_msg{Role: "assistant"}
	{
		txt := out.Content
		if len(out.Citations) > 0 {
			txt += "\nCitations:\n"
			for _, ct := range out.Citations {
				txt += ct + "\n"
			}
		}
		msg.AddText(txt)
	}
	msg.Tool_calls = out.Tool_calls
	agent.Props.Messages = append(agent.Props.Messages, msg)

	agent.callTools(out.Tool_calls, server)

	return len(out.Tool_calls) > 0
}

func (agent *Agent) RunLoop(max_iters int, max_tokens int, server *NetServer) {
	orig_max_iters := max_iters
	orig_max_tokens := max_tokens

	if max_iters <= 0 {
		max_iters = 1000000000 //1B
	}
	if max_tokens <= 0 {
		max_tokens = 1000000000 //1B
	}

	for max_iters > 0 {
		if !agent.Run(server) {
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

func (agent *Agent) callTools(tool_calls []OpenAI_completion_msg_Content_ToolCall, server *NetServer) {
	for _, it := range tool_calls {
		for _, tool := range agent.Props.Tools {
			if tool.Function.Name == it.Function.Name {

				//call
				cmd := exec.Command(fmt.Sprintf("./tools/%s/bin", it.Function.Name), strconv.Itoa(server.port))
				cmd.Dir = ""
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Start()
				if err != nil {
					fmt.Println("Error:", err)
				}

				cl, err := server.Accept()
				if err != nil {
					fmt.Println("Error:", err)
				}
				err = cl.WriteArray([]byte(it.Function.Arguments))
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

					case 2: //RunAgent
						max_iters, _ := cl.ReadInt()
						max_tokens, _ := cl.ReadInt()
						use_cases, _ := cl.ReadArray()
						systemPrompt, _ := cl.ReadArray()
						userPrompt, _ := cl.ReadArray()

						//init
						agent2 := NewAgent(string(use_cases), string(systemPrompt), string(userPrompt))
						defer agent2.Save(false)

						//run
						agent2.RunLoop(int(max_iters), int(max_tokens), server)

						//send result back
						cl.WriteArray([]byte(agent2.GetFinalMessage()))
						agent2.PrintStats()

					case 3: //SetToolCode
						toolName, _ := cl.ReadArray()
						toolCode, _ := cl.ReadArray()

						os.MkdirAll("tools/"+string(toolName), os.ModePerm)
						err := os.WriteFile(fmt.Sprintf("tools/%s/tool.go", toolName), toolCode, 0644)
						if err != nil {
							fmt.Println(err)
						}

						err = CompileTool(string(toolName))
						if err == nil {
							//ok
							cl.WriteArray(nil)
						} else {
							//error
							cl.WriteArray([]byte(fmt.Sprintf("Tool '%s' was created, but compiler reported error: %v", toolName, err)))
						}

						agent.AddTool(string(toolName))

					case 4: //Sandbox_violation
						info, _ := cl.ReadArray()
						agent.Sandbox_violations = append(agent.Sandbox_violations, string(info))
						fmt.Println("Sandbox violation:", string(info))
						cl.WriteInt(1) //block it
					}
				}

				err = cmd.Wait()
				if err != nil {
					//tool crashed
					js = []byte(fmt.Sprintf("Tool '%s' crashed with log.Fatal: %s", tool.Function.Name, err.Error()))
				}

				//save
				msg := OpenAI_completion_msg{Role: "tool"}
				msg.Tool_call_id = it.Id
				msg.AddText(string(js))
				//msg.AddImage()
				agent.Props.Messages = append(agent.Props.Messages, msg)

				fmt.Println("+Tool returns:", msg)
			}
		}
	}
}
