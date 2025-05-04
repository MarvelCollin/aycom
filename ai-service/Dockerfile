FROM python:3.9-slim AS builder

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy entire codebase
COPY . .

# Final stage
FROM python:3.9-slim

# Add certificates and create non-root user
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    groupadd -r appgroup && \
    useradd -r -g appgroup appuser

WORKDIR /app

# Copy installed packages and code from builder
COPY --from=builder --chown=appuser:appgroup /usr/local/lib/python3.9/site-packages /usr/local/lib/python3.9/site-packages
COPY --from=builder --chown=appuser:appgroup /app /app

# Set environment variables
ENV PYTHONUNBUFFERED=1

# Use non-root user
USER appuser

# Expose ports
EXPOSE 5000

# Start the application
CMD ["python", "app.py"] 