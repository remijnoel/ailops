# ailops

Ailops is a CLI tool to perform debugging with the help of a large language model (OpenAI for now).

## Usage

To diagnose an issue on a host, you can use the `diagnose` command:

```bash
ailops diagnose --description "Describe the issue here" --interactive
```

To a remote host, you can use the `--remote` option:

```bash
ailops diagnose -i -d "Describe the issue here" --remote user@host
```

The `--interactive` or `-i` flag will run the diagnosis but ask you to confirm each new step. If you want to run the diagnosis in a non-interactive mode, you can omit this flag. In that case, the diagnosis will run without an visual feedback and will print the results to the console.

To generate a markdown report of the diagnosis, you can use the `--report` option. This will create a `.ailops` directory in the current working directory with the report files.

```bash
ailops diagnose --description "Describe the issue here" --report
```

To use an alternate base URL for the OpenAI API (tested with LiteLLM only), you can either:

- Set the `--base-url` option when running the command
- Set the `AILOPS_BASE_URL` environment variable
- Or use the configuration file key `base_url`

## Build

To build the project, just use the Makefile:

```bash
make build # Will build using the system's specs

# For cross-compilation, you can use:
make build-linux-amd64
make build-linux-arm64
make build-darwin-amd64
```

## Configuration

### API Key

To use the OpenAI API, set the `OPENAI_API_KEY` environment variable with your key

```bash
export OPENAI_API_KEY=your_openai_api_key
```

To use an Azure endpoint, set the API key and endpoint as follows:

```bash
export AZURE_OPENAI_API_KEY=<your_azure_openai_api_key>
export AILOPS_AZURE_OPENAI_ENDPOINT=https://<your-azure-endpoint>.openai.azure.com/
export AILOPS_AZURE_OPENAI_API_VERSION=2023-05-15
```

### Other Configurations

The tool can be configured using a configuration file or environment variables. The following keys are allowed:

- `log_level`: The log level for the application (default: `warn`)
- `cmd_whitelist`: A list of commands that are allowed to be executed (default: `[]`)
- `cmd_blacklist`: A list of commands that are not allowed to be executed (default: `[]`)
- `initial_commands`: A list of commands that will be executed at the start
  - Default:
    - "top -b -n1 | head -20"
    - "ps aux | head -10"
    - "df -h"
    - "free -h"
    - "dmesg | tail -n 50"

### Loading Configuration

```bash
ailops --config /path/to/config.yaml
```

Environment variables can also be used to configure the tool. The environment variable names are prefixed with `AILOPS_` and the keys are converted to uppercase. For example, to set the `log_level`, you can use:

```bash
export AILOPS_LOG_LEVEL=info
```
