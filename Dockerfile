# ============================================================
# Stage 1: Builder
# Compiles the Go source into a WASM binary.
# ============================================================
FROM golang:1.23 AS builder

WORKDIR /usr/src/app

# --- Dependency caching layer ---
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
# ============================================================
FROM nginx:1.27-alpine

# Security hardening: run nginx as non-root
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy built assets and nginx config
COPY --from=builder /usr/src/app/web/ /usr/share/nginx/html/
COPY --from=builder /usr/src/app/web/nginx.conf /etc/nginx/conf.d/default.conf

# Fix permissions for non-root execution
RUN chown -R appuser:appgroup /usr/share/nginx/html && \
    chown -R appuser:appgroup /var/cache/nginx && \
    chown -R appuser:appgroup /var/log/nginx && \
    touch /var/run/nginx.pid && \
    chown appuser:appgroup /var/run/nginx.pid

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
