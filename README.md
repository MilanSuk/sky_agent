## What is an AI Agent?
AI agent is an autonomous program that solves user problems.
To solve a high range of problems, **Agents can *not* be fixed flow and can *not* have fixed tools**. Agent must be able to create new tools from scratch.


## The repository
This was my weekend project. Todays companies represent AI Agents as something complex, so I decided to create one from scratch(without using 3rd party libraries).

The result is fully autonomos agent which has around 800 lines of code. It supports any OpenAI-compatible services and local servers.

How it works? If you write prompt and there is no tool, the default tool called "create_new_tool" will write the code and create new tool. Then agent will use the new tool and so on. There is also tool "update_tool" which is good for fixing bugs in the tools.

This repository is basically a manager for compiling, running, and communicating with tools. And calling LLMs.
**The agent and tools do not have any sandboxing!** This repository is for learning purposes. As mentioned, it's a low number of lines of code in a few .go files. It should be easy to hack on.



## Example prompts
TODO



## Compile
Install Go language. It's needed to compile new tools which agent can create.
- https://go.dev/doc/install

Model settings:
- Open models.go and replace <pre><code>your_api_key</code></pre>
- If needed, edit constant g_model_agent, g_model_coder, g_model_search.

Compile:
<pre><code>git clone github.com/milansuk/sky_agent
cd sky_agent
go build
./sky_agent
</code></pre>



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
