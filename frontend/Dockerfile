# Multi-stage build for production
FROM node:20-alpine AS builder

# Build argument for API URL
ARG VITE_API_URL=https://api.odowell.pro

WORKDIR /app

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci --prefer-offline --no-audit

# Copy source code
COPY . .

# Build for production with environment variable
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# Production stage - nginx with pinned version
FROM nginx:1.25-alpine

# Install wget for healthcheck
RUN apk add --no-cache wget

# Copy built files from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Create directories and fix permissions for non-root nginx user
RUN mkdir -p /var/cache/nginx/client_temp \
             /var/cache/nginx/proxy_temp \
             /var/cache/nginx/fastcgi_temp \
             /var/cache/nginx/uwsgi_temp \
             /var/cache/nginx/scgi_temp \
             /var/run && \
    chown -R nginx:nginx /var/cache/nginx /var/run /usr/share/nginx/html && \
    chmod -R 755 /usr/share/nginx/html && \
    # Allow nginx to write pid file
    touch /var/run/nginx.pid && \
    chown nginx:nginx /var/run/nginx.pid

# Use non-root user (nginx user already exists in nginx:alpine with uid 101)
USER nginx

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
