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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OpenAI_completion_out struct {
	Citations  []string
	Content    string
	Tool_calls []OpenAI_completion_msg_Content_ToolCall

	inputTokens  int
	outputTokens int
	totalTokens  int
	time         float64
}

func OpenAI_completion_Run(jsProps []byte, Completion_url string, Api_key string) (OpenAI_completion_out, error) {

	startTime := float64(time.Now().UnixMilli()) / 1000

	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return OpenAI_completion_out{}, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+Api_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return OpenAI_completion_out{}, fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	js, err := io.ReadAll(res.Body)
	if err != nil {
		return OpenAI_completion_out{}, err
	}

	type STChoice struct {
		Message OpenAI_completion_out
		//Delta   STMsg
	}
	type STUsage struct {
		Prompt_tokens     int
		Completion_tokens int
		Total_tokens      int
	}
	type STError struct {
		Message string
	}
	type ST struct {
		Citations []string
		Choices   []STChoice
		Usage     STUsage
		Error     STError
	}
	var st ST
	err = json.Unmarshal(js, &st)
	if err != nil {
		return OpenAI_completion_out{}, err
	}
	if st.Error.Message != "" {
		return OpenAI_completion_out{}, errors.New(st.Error.Message)
	}

	out := OpenAI_completion_out{}
	if len(st.Choices) > 0 {
		out = st.Choices[0].Message
		out.Citations = st.Citations
	} else if len(st.Citations) > 0 {
		out.Citations = st.Citations
	} else {
		return out, err
	}

	out.inputTokens = st.Usage.Prompt_tokens
	out.outputTokens = st.Usage.Completion_tokens
	out.totalTokens = st.Usage.Total_tokens

	out.time = (float64(time.Now().UnixMilli()) / 1000) - startTime

	fmt.Printf("+LLM generated %dtoks which took %.1fsec = %.1f toks/sec\n", st.Usage.Completion_tokens, out.time, float64(st.Usage.Completion_tokens)/out.time)
	fmt.Println("+LLM returns content:", out.Content)
	fmt.Println("+LLM returns tool_calls:", out.Tool_calls)

	if res.StatusCode != 200 {
		return OpenAI_completion_out{}, fmt.Errorf("statusCode %d != 200, response: %v", res.StatusCode, out)
	}

	return out, nil
}
