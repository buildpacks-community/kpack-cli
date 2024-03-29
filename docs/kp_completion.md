## kp completion

Generate completion script

### Synopsis

To load completions:

Bash:

$ source <(kp completion bash)

# To load completions for each session, execute once:
Linux:
  $ kp completion bash > /etc/bash_completion.d/kp
MacOS:
  $ kp completion bash > /usr/local/etc/bash_completion.d/kp

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ kp completion zsh > "${fpath[1]}/_kp"

# You will need to start a new shell for this setup to take effect.

Fish:

$ kp completion fish | source

# To load completions for each session, execute once:
$ kp completion fish > ~/.config/fish/completions/kp.fish


```
kp completion [bash|zsh|fish|powershell]
```

### Options

```
  -h, --help   help for completion
```

### SEE ALSO

* [kp](kp.md)	 - 

