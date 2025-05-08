# AYCOM AI Service

This service provides AI-powered functionality for the AYCOM platform, including thread categorization and content analysis.

## Features

- Thread categorization: Automatically categorize threads based on content
- Health checking endpoint
- Docker integration

## API Endpoints

### Health Check
```
GET /health
```
Returns the health status of the service and whether models are loaded.

### Thread Categorization
```
POST /predict/category
```
Categorizes the content of a thread.

Request body:
```json
{
  "content": "This is the text content of the thread"
}
```

Response:
```json
{
  "category": "technology",
  "confidence": 0.85,
  "all_categories": {
    "technology": 0.85,
    "politics": 0.05,
    "entertainment": 0.03,
    "sports": 0.02,
    "business": 0.05
  }
}
```

## Development

### Prerequisites
- Python 3.9+
- Docker and Docker Compose

### Local Setup
1. Install dependencies:
   ```
   pip install -r requirements.txt
   ```

2. Run the service:
   ```
   python app.py
   ```

### Docker Setup
1. Build the Docker image:
   ```
   docker build -t aycom-ai-service .
   ```

2. Run with Docker Compose:
   ```
   docker-compose up ai_service
   ```

## Model Information

The service uses a pre-trained text classification model:
- `thread_category_model.h5`: TensorFlow model for thread categorization
- `tokenizer.pickle`: Text tokenizer for preprocessing input text

## Environment Variables

- `PORT`: Port to run the service on (default: 5000)
- `FLASK_ENV`: Environment setting (`development` or `production`)
- `LOG_LEVEL`: Logging level (default: INFO) 