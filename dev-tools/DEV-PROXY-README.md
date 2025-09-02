# Development Proxy Setup for EchoLlama

This directory contains tools for setting up a universal proxy that redirects ALL traffic from any interface to the local EchoLlama server, ensuring complete isolation and control during development.

## What it does

The proxy listens on `0.0.0.0:11434` and forwards all requests to `127.0.0.1:11435` where EchoLlama runs. This ensures:

- ✅ **Unbreakable redirection**: No matter how requests are constructed (local, remote, any alias) - they're caught and sent to local EchoLlama
- ✅ **Perfect for testing and isolation**: Ensures experimental modes never call upstream APIs accidentally
- ✅ **Zero client configuration**: Clients can connect to any IP/hostname - proxy handles the routing

## Usage Options

### Option 1: Bash Script with nginx (Recommended)

```bash
cd dev-tools
./dev-proxy-setup.sh
```

This script:
1. Installs nginx if needed
2. Configures proxy on 0.0.0.0:11434 → 127.0.0.1:11435
3. Builds and starts EchoLlama on port 11435
4. Tests the proxy setup
5. Runs until Ctrl+C

### Option 2: Go Proxy (Minimal)

```bash
# Terminal 1: Start EchoLlama on port 11435
OLLAMA_HOST=127.0.0.1:11435 ./echollama serve

# Terminal 2: Start Go proxy
cd dev-tools
go run dev-proxy.go
```

### Option 3: Manual nginx Configuration

Create `/etc/nginx/sites-available/echollama-proxy`:

```nginx
server {
  listen 11434;                     # Listen on all interfaces
  server_name _;                    # Match any host
  location / {
    proxy_pass http://127.0.0.1:11435;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
}
```

## Testing

Once running, test with any of these - they should all work:

```bash
curl http://localhost:11434/api/version
curl http://127.0.0.1:11434/api/version  
curl http://0.0.0.0:11434/api/version
```

## Architecture

```
Client Request (any interface/hostname:11434)
    ↓
Proxy (0.0.0.0:11434) 
    ↓
EchoLlama (127.0.0.1:11435)
```

This setup guarantees that **all traffic lands at the local EchoLlama server** regardless of client configuration, providing absolute control and assurance over API request routing for advanced development, integration, and testing scenarios.