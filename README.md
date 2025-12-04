# Atelier Go

```bash
export ATELIER_HOST=localhost
export ATELIER_REMOTE_PATH=$HOME/.local/bin/atelier-server 
~/tmp/atelier-go/bash/atelier-client launch 
```

(Note: SSH to localhost might require setup keys)

## Local Testing

To test the scripts locally from within the source directory:

1. Ensure you can SSH to localhost (`ssh localhost` should work).
2. Run the client pointing to the local server script:

\`\`\`bash
ATELIER_HOST=localhost ATELIER_REMOTE_PATH="$(pwd)/atelier-server" ./atelier-client
\`\`\`
