**This repository was merged into Skyalt(https://github.com/milanSuk/skyalt) which provides GUI for agents. From now on, all fixes and improvements are done in Skyalt.**


## What is an AI Agent?
AI agent is an autonomous program that solves user problems.
To solve a high range of problems, **Agents can *not* be fixed flow and can *not* have fixed tools**. Agent must be able to create new tools from scratch.



## The repository
This was my weekend project. Todays companies represent AI Agents as something complex, so I decided to create one from scratch(without using 3rd party libraries).

The result is general and fully autonomous agent which has around 800 lines of code. It supports any OpenAI-compatible services and local servers.

How it works? If you write prompt and there is no tool, the default tool called `create_new_tool` will write the code for the new tool. Then agent will use that new tool and so on. There is also tool `update_tool` which is good for fixing bugs in the tools.

This repository is basically a manager for compiling, running, and communicating with tools. And calling LLMs.
The agent and tools have simple sandboxing, check `tools/sdk_sandbox_fns.txt`. This repository is for learning purposes. As mentioned, it's a low number of lines of code in a few .go files. It should be easy to hack on.



## Example prompts
*"What is the population of Prague, Paris and Los Angeles? Use OpenStreetMap's Nominatim API to get latest data."*
- agent calls `create_new_tool` to create new tool `get_city_population`.
- agent calls 3x `get_city_population` to get population numbers.
- agent answers.

*"Search the web for How many stars are in the universe?"*
- agent call `web_search`(Perplexity returns few paragraphs)
- agent answers.

*"Compute amount of burn calories in 'morning_run.gpx' file."*
- agent calls `create_new_tool` to create new tool `calculate_calories_from_gpx_file`.
- agent calls `access_disk` to get weight of the user.
- agent writes and execute SQL query to get weight of the user.
- agent calls `calculate_calories_from_gpx_file` to get calories.
- agent answers.
- *note: you need to have file 'morning_run.gpx' in repo folder!*

*"Send email to <email_address>. Subject: Test mail. Body: Hi there!"*
- agent calls `create_new_tool` to create new tool `send_email`.
- agent calls `access_disk` to get STMP login information
- agent writes and execute SQL query to get STMP login information from database.
- agent calls `send_email`.
- *note: put email login into database(password_id) and passwords.json(plain passwoord)



## Compile & Run
Model settings:
- Open `models.go` and replace `<your_api_key>`.
- If needed, edit constants `g_model_agent`, `g_model_coder`, `g_model_search`.

Install Go language. It's needed to compile new tools which agent can create.
- https://go.dev/doc/install
- install tools:
<pre><code>go install golang.org/x/tools/cmd/gopls@latest
go install golang.org/x/tools/cmd/goimports@latest
</code></pre>

Compile:
<pre><code>git clone https://github.com/milansuk/sky_agent
cd sky_agent
go mod tidy
go build
</code></pre>

Run:
<pre><code>./sky_agent "Search the web for How many stars are in the universe?"</code></pre>



## Author
Milan Suk

Email: milan@skyalt.com

X: https://x.com/milansuk/

**Sponsor**: https://github.com/sponsors/milansuk

*Feel free to follow or contact me with any idea, question or problem.*



## Contributing
Your feedback and code are welcome!

For bug report or question, please use [GitHub's Issues](https://github.com/milansuk/sky_agent/issues)

Sky_agent is licensed under **Apache v2.0** license. This repository includes 100% of the code.