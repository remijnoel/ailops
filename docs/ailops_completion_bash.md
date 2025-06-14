## ailops completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(ailops completion bash)

To load completions for every new session, execute once:

#### Linux:

	ailops completion bash > /etc/bash_completion.d/ailops

#### macOS:

	ailops completion bash > $(brew --prefix)/etc/bash_completion.d/ailops

You will need to start a new shell for this setup to take effect.


```
ailops completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -c, --config string   Path to configuration file
      --debug           Enable verbose logging
```

### SEE ALSO

* [ailops completion](ailops_completion.md)	 - Generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 13-Jun-2025
