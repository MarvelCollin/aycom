FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies with cache optimization
RUN npm ci --prefer-offline --no-audit

# Set environment variables
ENV NODE_ENV=development
ENV VITE_HMR=true

# Expose port
EXPOSE 3000

# Start development server with hot reload enabled
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "3000"] 