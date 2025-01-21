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

import "strings"

type Model struct {
	Name         string
	Input_price  float64
	Output_price float64
}

type Service struct {
	Name           string
	Completion_url string
	Api_key        string

	Models        []Model
	Default_model string
}

const g_model_agent = "gpt-4o"
const g_model_coder = "gpt-4o"
const g_model_search = "llama-3.1-sonar-large-128k-online"

var g_services = []Service{
	{Name: "xai", Completion_url: "https://api.x.ai/v1/chat/completions", Api_key: "<your_api_key>",
		Models: []Model{
			{Name: "grok-2-vision", Input_price: 2, Output_price: 10},
			{Name: "grok-2", Input_price: 2, Output_price: 10},
			{Name: "grok-vision-beta", Input_price: 2, Output_price: 15},
			{Name: "grok-beta", Input_price: 2, Output_price: 15},
		},
	},

	{Name: "openai", Completion_url: "https://api.openai.com/v1/chat/completions", Api_key: "<your_api_key>",
		Models: []Model{
			{Name: "gpt-3.5-turbo", Input_price: 0.5, Output_price: 1.5},
			{Name: "gpt-4", Input_price: 30, Output_price: 60},
			{Name: "gpt-4-turbo", Input_price: 10, Output_price: 30},
			{Name: "gpt-4o", Input_price: 2.5, Output_price: 10},
			{Name: "gpt-4o-mini", Input_price: 0.15, Output_price: 0.6},
			{Name: "o1", Input_price: 15, Output_price: 60},
			{Name: "o1-mini", Input_price: 3, Output_price: 12},
		},
	},

	{Name: "perplexity", Completion_url: "https://api.perplexity.ai/chat/completions", Api_key: "<your_api_key>",
		Models: []Model{
			{Name: "llama-3.1-sonar-small-128k-online", Input_price: 0, Output_price: 0.2},
			{Name: "llama-3.1-sonar-large-128k-online", Input_price: 0, Output_price: 1},
			{Name: "llama-3.1-sonar-huge-128k-online", Input_price: 0, Output_price: 5},
		},
	},

	{Name: "local", Completion_url: "http://localhost:8090/v1/chat/completions", //8090 = port number. Replace it
		Models: []Model{
			//{Name: "name_of_model", Input_price: 0, Output_price: 0},
		},
	},
}

func Service_findModel(model string) *Model {
	model = strings.ToLower(model)
	for _, srv := range g_services {
		for _, md := range srv.Models {
			if md.Name == model {
				return &md
			}
		}
	}
	return nil
}
func Service_findService(model string) *Service {
	model = strings.ToLower(model)
	for _, srv := range g_services {
		for _, md := range srv.Models {
			if md.Name == model {
				return &srv
			}
		}
	}
	return nil
}

func Service_findModelFromUse_cases(use_case string) string {
	switch strings.ToLower(use_case) {
	case "agent":
		return g_model_agent
	case "coder":
		return g_model_coder
	case "search":
		return g_model_search
	}
	return g_model_agent //default
}
