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
	Name                     string
	OpenAI_completion_url    string
	Anthropic_completion_url string
	Api_key                  string

	Models        []Model
	Default_model string
}

// grok-2
// gpt-4o-mini
// mistral-large-latest
// mistral-small-latest
// codestral-latest
const g_model_agent = "grok-2"
const g_model_coder = "grok-2"
const g_model_search = "llama-3.1-sonar-large-128k-online"

var g_services = []Service{
	{Name: "xai", OpenAI_completion_url: "https://api.x.ai/v1/chat/completions" /*, Anthropic_completion_url: "https://api.x.ai/v1/messages"*/, Api_key: "<your_api_key>",
		Models: []Model{
			//https://docs.x.ai/docs/models
			{Name: "grok-2-vision", Input_price: 2, Output_price: 10},
			{Name: "grok-2", Input_price: 2, Output_price: 10},
			{Name: "grok-vision-beta", Input_price: 2, Output_price: 15},
			{Name: "grok-beta", Input_price: 2, Output_price: 15},
		},
	},

	{Name: "openai", OpenAI_completion_url: "https://api.openai.com/v1/chat/completions", Api_key: "<your_api_key>",
		Models: []Model{
			//https://platform.openai.com/docs/pricing
			//{Name: "gpt-3.5-turbo", Input_price: 0.5, Output_price: 1.5},
			{Name: "gpt-4", Input_price: 30, Output_price: 60},
			{Name: "gpt-4-turbo", Input_price: 10, Output_price: 30},
			{Name: "gpt-4o", Input_price: 2.5, Output_price: 10},
			{Name: "gpt-4o-mini", Input_price: 0.15, Output_price: 0.6},
			//{Name: "o1", Input_price: 15, Output_price: 60},
			//{Name: "o1-mini", Input_price: 3, Output_price: 12},
		},
	},

	{Name: "anthropic", Anthropic_completion_url: "https://api.anthropic.com/v1/messages", Api_key: "<your_api_key>",
		Models: []Model{
			//https://www.anthropic.com/pricing#anthropic-api
			{Name: "claude-3-5-haiku-latest", Input_price: 0.8, Output_price: 4},
			{Name: "claude-3-5-sonnet-latest", Input_price: 3, Output_price: 15},
		},
	},

	{Name: "mistral", OpenAI_completion_url: "https://api.mistral.ai/v1/chat/completions", Api_key: "<your_api_key>",
		Models: []Model{
			//https://mistral.ai/technology/#pricing
			{Name: "mistral-large-latest", Input_price: 2, Output_price: 6},
			{Name: "pixtral-large-latest", Input_price: 2, Output_price: 6},
			{Name: "mistral-small-latest", Input_price: 0.2, Output_price: 0.6},
			{Name: "codestral-latest", Input_price: 0.3, Output_price: 0.9},
			{Name: "pixtral-12b-2409", Input_price: 0.15, Output_price: 0.15},  //free?
			{Name: "open-mistral-nemo", Input_price: 0.15, Output_price: 0.15}, //free?
		},
	},

	{Name: "groq", OpenAI_completion_url: "https://api.groq.com/openai/v1/chat/completions", Api_key: "<your_api_key>",
		Models: []Model{
			//https://groq.com/pricing/
			{Name: "llama-3.3-70b-versatile", Input_price: 0.59, Output_price: 0.79},
			{Name: "llama-3.3-70b-specdec", Input_price: 0.59, Output_price: 0.99},
			{Name: "llama-3.1-8b-instant", Input_price: 0.05, Output_price: 0.08},
			{Name: "gemma2-9b-it", Input_price: 0.2, Output_price: 0.2},
			//{Name: "deepseek-r1-distill-llama-70b", Input_price: 0.59, Output_price: 0.79},
		},
	},

	{Name: "google", OpenAI_completion_url: "https://generativelanguage.googleapis.com/v1beta/chat/completions", Api_key: "<your_api_key>-8",
		Models: []Model{
			//...
			{Name: "gemini-1.5-flash", Input_price: 0, Output_price: 0},     //price? ...
			{Name: "gemini-2.0-flash-exp", Input_price: 0, Output_price: 0}, //price? ...
		},
	},

	{Name: "perplexity", OpenAI_completion_url: "https://api.perplexity.ai/chat/completions", Api_key: "<your_api_key>",
		Models: []Model{
			//https://docs.perplexity.ai/guides/pricing
			{Name: "llama-3.1-sonar-small-128k-online", Input_price: 0.2, Output_price: 0.2},
			{Name: "llama-3.1-sonar-large-128k-online", Input_price: 1, Output_price: 1},
			{Name: "llama-3.1-sonar-huge-128k-online", Input_price: 5, Output_price: 5},
		},
	},

	{Name: "local", OpenAI_completion_url: "http://localhost:8090/v1/chat/completions", //8090 = port number. Replace it
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
