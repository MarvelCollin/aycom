FROM python:3.8-slim

# Set environment variables
ENV PYTHONUNBUFFERED=1 \
    TF_CPP_MIN_LOG_LEVEL=2 \
    PYTHONDONTWRITEBYTECODE=1 \
    PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION=python

# Install dependencies
WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Create models directory
RUN mkdir -p /app/models

# Copy application code
COPY app.py /app/
COPY thread_category_model.h5 /app/models/
COPY tokenizer.pickle /app/models/

# Expose port
EXPOSE 5000

# Run app
CMD ["python", "app.py"] 