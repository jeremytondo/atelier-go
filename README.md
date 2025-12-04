# Atelier Go

There are two environment variables that need to be set. Below are the
default settings.

```bash
export ATELIER_HOST=localhost
export ATELIER_REMOTE_PATH=$HOME/.local/bin/atelier-server 
```

## Local Testing

To test the scripts locally from within the source directory:

1. Ensure you can SSH to localhost (`ssh localhost` should work).
2. Run the client pointing to the local server script:

```bash
atelier-client
```
