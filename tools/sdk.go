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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

var _sdk_client *SDK_NetClient

func main() {
	if len(os.Args) < 2 {
		log.Fatal("missing 'port' argument: ", os.Args)
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	//connect to server
	_sdk_client = SDK_NewNetClient("localhost", port)
	defer _sdk_client.Destroy()

	//get tool input
	input := _sdk_client.ReadArray()
	var st _replace_with_tool_structure_
	err = json.Unmarshal(input, &st)
	if err != nil {
		log.Fatal(err)
	}

	//exe tool
	output, err := json.Marshal(st.run())
	if err != nil {
		log.Fatal(err)
	}

	//send back result
	_sdk_client.WriteInt(1)
	_sdk_client.WriteArray(output)

}

// use_case = "agent", "coder", "search"
func SDK_RunAgent(use_case string, max_iters int, max_tokens int, systemPrompt string, userPrompt string) string {
	_sdk_client.WriteInt(2)
	_sdk_client.WriteInt(uint64(max_iters))
	_sdk_client.WriteInt(uint64(max_tokens))
	_sdk_client.WriteArray([]byte(use_case))
	_sdk_client.WriteArray([]byte(systemPrompt))
	_sdk_client.WriteArray([]byte(userPrompt))

	js := _sdk_client.ReadArray()
	return string(js)
}

func SDK_SetToolCode(tool string, code string) string {
	_sdk_client.WriteInt(3)
	_sdk_client.WriteArray([]byte(tool))
	_sdk_client.WriteArray([]byte(code))

	js := _sdk_client.ReadArray()
	return string(js)
}

type SDK_NetClient struct {
	conn *net.TCPConn
}

func SDK_NewNetClient(addr string, port int) *SDK_NetClient {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	return &SDK_NetClient{conn: conn}
}
func (client *SDK_NetClient) Destroy() {
	client.conn.Close()
}

func (client *SDK_NetClient) ReadInt() uint64 {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if err != nil {
		log.Fatal(err)
	}

	return binary.LittleEndian.Uint64(sz[:])
}

func (client *SDK_NetClient) WriteInt(value uint64) {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if err != nil {
		log.Fatal(err)
	}
}

func (client *SDK_NetClient) ReadArray() []byte {
	//recv size
	size := client.ReadInt()

	//recv data
	data := make([]byte, size)
	p := 0
	for p < int(size) {
		n, err := client.conn.Read(data[p:])
		if err != nil {
			log.Fatal(err)
		}
		p += n
	}

	return data
}

func (client *SDK_NetClient) WriteArray(data []byte) {
	//send size
	client.WriteInt(uint64(len(data)))

	//send data
	_, err := client.conn.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}
