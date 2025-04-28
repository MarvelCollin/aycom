FROM node:18-alpine

WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm install

# Copy the rest of the code
COPY . .

# Set development as default environment
ENV NODE_ENV=development

# Expose the port that matches our docker-compose configuration
EXPOSE 3000

# Default to development mode - binding to all interfaces
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "3000"] 