FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm install

# Copy the rest of the code
COPY . .

# Set environment variables
ENV NODE_ENV=development

# Create a non-root user
FROM node:18-alpine

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy built app from builder stage
COPY --from=builder --chown=appuser:appgroup /app ./

# Set permissions
RUN chown -R appuser:appgroup /app

# Use the non-root user
USER appuser

# Expose the port
EXPOSE 3000

# Start the application
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "3000"] 