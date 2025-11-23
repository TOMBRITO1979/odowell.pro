# Multi-stage build for production
FROM node:18-alpine AS builder

# Build argument for API URL
ARG VITE_API_URL=https://drapi.crwell.pro

WORKDIR /app

# Copy package files
COPY package.json ./

# Install dependencies
RUN npm install

# Copy source code
COPY . .

# Build for production with environment variable
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# Production stage - nginx
FROM nginx:alpine

# Copy built files from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Fix permissions for all files
RUN chmod -R 755 /usr/share/nginx/html && \
    find /usr/share/nginx/html -type f -exec chmod 644 {} \;

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
