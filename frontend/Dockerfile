FROM node:18

WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm install

# Copy the rest of the code
COPY . .

# Set development as default environment
ENV NODE_ENV=development

# Expose Vite's default port
EXPOSE 5173

# Default to development mode
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"] 