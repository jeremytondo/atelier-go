# Remote Execution Plan

I'm working on a simple method for using Atelier Go on a remote server that does
not require a server running on the remote host. The general idea behind this is
that we just run the app via a simple ssh exec command and interact with it completely
on the server.

This could be as simple as setting up an ssh config like this for each reomte
server you want to work with:

```bash
Host ag
  HostName workstation   
  User jeremytondo
  ControlMaster auto
  ControlPath ~/.ssh/cm-%C
  ControlPersist 10m
  ServerAliveInterval 60
  ServerAliveCountMax 3
  RequestTTY yes
  LogLevel QUIET
  RemoteCommand /home/jeremytondo/.local/bin/atelier-go
```

Then in order to run Atelier Go on the remote server, you just need to run the 
ssh command like this:

```bash
ssh ag
```

If this works well, we don't need any server app for this at all and it greatly
simplifies the remote use case.

The biggest challenge right now is that running the remote command like I've
shown in the example does not give you a shell where you have things like the
PATH set up. This means that the current way we start sessions won't work on the
remote server since we can't just run `zmx attach`. If we do, it can't find the
application zmx due to the shell not really being a full shell.

Notice in the example that the remote command needs to have the full path. This
is one way around the issue and works. While this may work fine for this initial
command since it is easy for a user to find this path and set it up once in a 
config like this, it doesn't really work well once we get into starting sessions 
and running actions. In these cases we don't really know what apps a user may
want to run and asking them to get the full path to each would be a bit annoying.

So, we need a solution that is rock solid, maintains the quick native like feel
we're looking for, and works on both local and remote use cases.

## Proposed Solution: PATH Harvesting

To resolve the environment issues when running via `RemoteCommand` (non-interactive, non-login shell), we will implement a "PATH Harvesting" strategy. This allows the app to stay lightweight and fast while ensuring it has access to the user's full environment.

1.  **Detection**: At startup, `atelier-go` checks for the `SSH_CONNECTION` environment variable to determine if it's running in a remote context.
2.  **Login Shell Discovery**: The app identifies the user's actual login shell using `os/user.Current().Shell`. This is more reliable than the `$SHELL` environment variable in restricted SSH environments.
3.  **Harvesting**: If running over SSH, the app executes the discovered shell in interactive login mode (e.g., `/bin/zsh -l -i -c 'echo $PATH'`) to retrieve the full, correct `PATH`.
4.  **Robustness**: The app will handle potential "noise" from interactive shells (e.g., MOTD or startup messages) by taking the last valid-looking path line from the shell output.
5.  **Application**: The app updates its own process environment using `os.Setenv("PATH", harvestedPath)`. Since all subsequent child processes (like `zmx` or user actions) inherit the environment, this solves the path issue globally without requiring modifications to action execution logic.

## Implementation Steps

1.  **Environment Management**: Create a new utility or update `internal/config` to handle environment initialization.
2.  **Detection Logic**: Implement the check for `SSH_CONNECTION`.
3.  **Shell Discovery**: Implement logic to get the user's shell via `os/user.Current()`.
4.  **Shell Execution & Parsing**: Add logic to safely execute the shell to harvest the PATH and parse the last valid line of output.
5.  **Integration**: Call the environment initialization early in the `cmd/atelier-go/main.go` or within the root command initialization.

## Checklist
- [ ] Implement `SSH_CONNECTION` detection.
- [ ] Implement shell discovery via `os/user.Current().Shell`.
- [ ] Implement PATH harvesting with robustness for shell output noise.
- [ ] Update process environment with harvested PATH.
- [ ] Verify `zmx` attachment works correctly over SSH.

## Implementation Notes
- The harvesting step should be fast, but we should ensure it doesn't hang if the shell has heavy startup logic.
- Using `os/user.Current().Shell` ensures we use the shell the user expects, even if `$SHELL` is not propagated or is set to a generic value.
- By taking the last line of the output, we skip any banner text or warnings printed by the shell during initialization.
- This approach avoids the need for a complex client/server architecture while solving the most common friction point of remote execution.
