FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies with cache optimization
RUN npm ci --prefer-offline --no-audit

# Set environment variables
ENV NODE_ENV=development
ENV VITE_HMR=true
ENV VITE_GOOGLE_CLIENT_ID=161144128362-3jdhmpm3kfr253crkmv23jfqa9ubs2o8.apps.googleusercontent.com
ENV VITE_GOOGLE_REDIRECT_URI=http://localhost:3000/register
ENV VITE_RECAPTCHA_SITE_KEY=6Ld6UysrAAAAAPW3XRLe-M9bGDgOPJ2kml1yCozA

# Expose port
EXPOSE 3000

# Start development server with hot reload enabled
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "3000"] 