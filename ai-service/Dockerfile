FROM python:3.9-slim

# Add system dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /app

# Install Python dependencies
COPY requirements.txt /app/
RUN pip install --no-cache-dir -r requirements.txt

# Copy application files
COPY app.py /app/
COPY thread_category_model.h5 /app/
COPY tokenizer.pickle /app/

# Create a persistent volume for models
VOLUME /app/models

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV TF_CPP_MIN_LOG_LEVEL=2
ENV PYTHONDONTWRITEBYTECODE=1

# Expose the port
EXPOSE 5000

# Start the application
CMD ["python", "app.py"] 