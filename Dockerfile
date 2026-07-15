# ============================================================
# Stage 1: Builder
# Compiles the Go source into a WASM binary.
# ============================================================
FROM golang:1.23 AS builder

WORKDIR /usr/src/app

# --- Dependency caching layer ---
# Copy go.mod first so Docker can cache the module download step.
COPY go.mod ./
RUN go mod download 2>/dev/null || true

# Copy source code
COPY . .

# Build the WASM binary with size optimizations
RUN GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/main.wasm .

# Copy the Go WASM support JS from the SDK
RUN cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/ 2>/dev/null \
    || cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/

# ============================================================
# Stage 2: Runtime
# Serves the static frontend + compiled WASM using nginx.
# Alpine-based image keeps the final image small (~40MB).
# ============================================================
FROM nginx:1.27-alpine

# Security hardening: run nginx as non-root
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy built assets from builder
COPY --from=builder /usr/src/app/web/ /usr/share/nginx/html/

RUN cat > /etc/nginx/conf.d/default.conf << 'NGINX'
server {
    listen 8080;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    # --- MIME types ---
    types {
        application/wasm wasm;
        application/javascript js;
        text/css css;
        text/html html;
    }

    # --- Security headers ---
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # --- Compression ---
    gzip on;
    gzip_types application/javascript application/wasm text/css text/html;
    gzip_min_length 256;

    # --- Caching ---
    # WASM and JS: long cache
    location ~* \.(wasm|js)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # CSS and images: moderate cache
    location ~* \.(css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 7d;
        add_header Cache-Control "public";
    }

    # HTML: no cache to always serve the latest version
    location ~* \.html$ {
        expires -1;
        add_header Cache-Control "no-store, no-cache, must-revalidate";
    }

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }
}
NGINX

# Fix permissions for non-root execution
RUN chown -R appuser:appgroup /usr/share/nginx/html && \
    chown -R appuser:appgroup /var/cache/nginx && \
    chown -R appuser:appgroup /var/log/nginx && \
    touch /var/run/nginx.pid && \
    chown appuser:appgroup /var/run/nginx.pid

USER appuser

EXPOSE 8080

# Healthcheck to verify the container is serving
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
