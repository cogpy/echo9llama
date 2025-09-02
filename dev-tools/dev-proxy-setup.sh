#!/bin/bash
set -e

# EchoLlama Development Mode Loop - Universal Proxy Setup
# Sets up a proxy listening on 0.0.0.0:11434 that redirects all traffic to 127.0.0.1:11435
# This ensures every incoming request - external or internal, from any interface - is routed to local echollama

echo "=== EchoLlama Development Proxy Setup ==="
echo "Setting up universal proxy redirect: 0.0.0.0:11434 → 127.0.0.1:11435"
echo

# Check if nginx is installed
if ! command -v nginx &> /dev/null; then
    echo "Installing nginx..."
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        sudo apt-get update
        sudo apt-get install -y nginx
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install nginx
    else
        echo "Please install nginx for your system"
        exit 1
    fi
fi

# Create nginx configuration for universal proxy
echo "Creating nginx proxy configuration..."
NGINX_CONFIG_DIR="/tmp/echollama-dev"
mkdir -p "$NGINX_CONFIG_DIR"

cat > "$NGINX_CONFIG_DIR/nginx.conf" <<EOF
worker_processes 1;
error_log /tmp/echollama-dev/error.log;
pid /tmp/echollama-dev/nginx.pid;

events {
    worker_connections 1024;
}

http {
    access_log /tmp/echollama-dev/access.log;
    
    upstream echollama {
        server 127.0.0.1:11435;
    }
    
    server {
        listen 0.0.0.0:11434;           # Listen on ALL interfaces
        server_name _;                   # Match ANY host
        
        location / {
            proxy_pass http://echollama;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
            
            # Ensure no timeout issues during model operations
            proxy_read_timeout 300s;
            proxy_connect_timeout 10s;
            proxy_send_timeout 10s;
        }
    }
}
EOF

# Function to cleanup on exit
cleanup() {
    echo
    echo "=== Cleaning up ==="
    if [[ -f "/tmp/echollama-dev/nginx.pid" ]]; then
        echo "Stopping nginx proxy..."
        nginx -s quit -c "$NGINX_CONFIG_DIR/nginx.conf" 2>/dev/null || true
    fi
    if pgrep -f "echollama serve" > /dev/null; then
        echo "Stopping echollama server..."
        pkill -f "echollama serve" || true
    fi
    echo "Development proxy stopped."
}

trap cleanup EXIT INT TERM

# Start nginx proxy
echo "Starting nginx proxy on 0.0.0.0:11434..."
nginx -c "$NGINX_CONFIG_DIR/nginx.conf"

# Build echollama if needed
if [[ ! -f "../echollama" ]]; then
    echo "Building echollama..."
    cd .. && go build -o echollama . && cd dev-tools
fi

# Start echollama server on port 11435
echo "Starting echollama server on 127.0.0.1:11435..."
OLLAMA_HOST=127.0.0.1:11435 ../echollama serve &
ECHOLLAMA_PID=$!

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 5

# Test the proxy setup
echo
echo "=== Testing Proxy Setup ==="
echo "Testing direct echollama connection (127.0.0.1:11435)..."
if curl -s "http://127.0.0.1:11435/api/version" > /dev/null; then
    echo "✅ Direct echollama connection: OK"
else
    echo "❌ Direct echollama connection: FAILED"
fi

echo "Testing proxy connection (localhost:11434)..."
if curl -s "http://localhost:11434/api/version" > /dev/null; then
    echo "✅ Proxy connection: OK"
else
    echo "❌ Proxy connection: FAILED"
fi

echo "Testing external interface proxy (0.0.0.0:11434)..."
if curl -s "http://0.0.0.0:11434/api/version" > /dev/null; then
    echo "✅ External proxy connection: OK"
else
    echo "❌ External proxy connection: FAILED"
fi

echo
echo "=== Development Environment Ready ==="
echo "✅ Universal proxy active: ALL traffic to any interface:11434 → 127.0.0.1:11435"
echo "✅ EchoLlama server running on 127.0.0.1:11435"
echo "✅ Proxy listening on 0.0.0.0:11434"
echo
echo "Test with:"
echo "  curl http://localhost:11434/api/version"
echo "  curl http://127.0.0.1:11434/api/version"
echo "  curl http://0.0.0.0:11434/api/version"
echo
echo "All requests will be routed to the local echollama server."
echo "Press Ctrl+C to stop the development environment."
echo

# Keep running until interrupted
wait $ECHOLLAMA_PID