# Linux Terminal AI - go-figure

`go-figure` is a utility tool that leverages OpenAI to assist with Linux command suggestions. It provides step-by-step guidance in JSON format, explaining the purpose and reasoning behind each command. It can execute commands directly or write them to a file.

## Features
- Generate Linux command steps based on a user query.
- Execute commands interactively.
- Save commands to a file for later use.

## Requirements
- Go 1.20 or later
- OpenAI API Key

## Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/karma-ashutosh/go-figure.git
   cd go-figure```

2. Build the project
   ```./build.sh```
   This will create an executable named go-figure.

3. Set up your OpenAI API Key as an environment variable:
   ```
   export OPENAI_API_KEY=your-api-key
   ```

4. Run the utility in your terminal:
   ```
   ./go-figure
   ```
   Modes

The utility supports two modes:

* Execute Mode (Default): Commands are executed interactively.
    ```
    ./go-figure --mode execute
    ```
    
* Write-to-File Mode: Commands are written to a specified file.
    ```
    ./go-figure --mode write-to-file --file output.txt
    ```
