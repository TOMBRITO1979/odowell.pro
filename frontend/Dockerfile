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

# Fix permissions for all files
RUN chmod -R 755 /usr/share/nginx/html && \
    find /usr/share/nginx/html -type f -exec chmod 644 {} \;

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
