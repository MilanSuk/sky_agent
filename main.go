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
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	UserPrompt := "Search the web for How many stars are in the universe?" //default, if program doesn't have argument

	if len(os.Args) > 1 {
		UserPrompt = os.Args[1]
	}

	server := NewNetServer(8090)
	defer server.Destroy()

	mainAgent := NewAgent("agent", "", UserPrompt)
	if strings.ToLower(UserPrompt) == "continue" {
		mainAgent.Open("last.json") //recover previous state
	}
	defer mainAgent.Save(true)

	toolList, err := GetToolsList()
	if err != nil {
		log.Fatal(err)
	}

	for _, toolName := range toolList {
		if NeedCompileTool(toolName) {
			err := CompileTool(toolName)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	//add tools to mainAgent
	for _, toolName := range toolList {
		mainAgent.AddTool(toolName)
	}

	//run
	mainAgent.RunLoop(20, 20000, server)

	mainAgent.PrintStats()
	fmt.Println("Final answer:", mainAgent.GetFinalMessage())
}
