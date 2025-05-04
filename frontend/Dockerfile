FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm install

# Copy entire codebase
COPY . .

# Build the application
RUN npm run build

# Final stage
FROM node:18-alpine

# Add certificates and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

# Copy necessary files from builder
COPY --from=builder --chown=appuser:appgroup /app/package*.json ./
COPY --from=builder --chown=appuser:appgroup /app/node_modules ./node_modules
COPY --from=builder --chown=appuser:appgroup /app/dist ./dist
COPY --from=builder --chown=appuser:appgroup /app/public ./public
COPY --from=builder --chown=appuser:appgroup /app/vite.config.ts ./

# Set environment variables
ENV NODE_ENV=production

# Use non-root user
USER appuser

# Expose ports
EXPOSE 3000

# Start the application
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "3000"] 