# AYCOM AI Service

## Overview

The AYCOM AI Service provides machine learning capabilities to enhance the platform's functionality. Its primary features include automatic thread categorization, content analysis, and recommendation systems.

## Features

### 1. Automatic Thread Categorization

The service uses a trained neural network model to analyze thread content and predict the most appropriate category. This helps organize discussions more effectively and improves content discoverability.

#### How It Works

1. When a user creates a new thread, the content is sent to the AI service's `/predict/category` endpoint
2. The service preprocesses the text and passes it through the TensorFlow model
3. The model predicts category probabilities across all supported categories
4. The highest-confidence category is returned, along with confidence scores for all categories
5. The UI can auto-select categories with high confidence (>70%) or show them as suggestions

#### Supported Categories

| Category ID | Description |
|-------------|-------------|
| technology | Technology related topics like computers, software, hardware, programming, AI, algorithms, apps, and digital innovations |
| health | Health topics including fitness, diet, exercise, medical information, diseases, cures, treatments, and wellness |
| education | Education related discussions about schools, universities, learning, studying, colleges, academics, courses, and student life |
| entertainment | Entertainment topics like movies, music, games, plays, concerts, shows, actors, films, and media series |
| science | Scientific topics including research, experiments, labs, scientists, discoveries, physics, chemistry, and biology |
| sports | Sports related content about football, basketball, soccer, games, teams, players, matches, and tournaments |
| politics | Political discussions including government, elections, voting, policies, presidents, congress, and political parties |
| business | Business topics about companies, markets, finance, economy, stocks, investments, startups, and entrepreneurship |
| lifestyle | Lifestyle content about homes, decor, fashion, trends, design, style, living, and clothing |
| travel | Travel related topics including vacations, tours, destinations, hotels, flights, trips, journeys, and tourism |
| other | General topics that don't fit into the other specific categories |

#### API Response Format

```json
{
  "category": "technology",
  "confidence": 0.87,
  "all_categories": {
    "technology": 0.87,
    "science": 0.05,
    "education": 0.03,
    "entertainment": 0.02,
    "business": 0.01,
    "politics": 0.01,
    "health": 0.00,
    "sports": 0.00,
    "lifestyle": 0.00,
    "travel": 0.00,
    "other": 0.01
  }
}
```

## API Endpoints

### GET /health

Health check endpoint to verify the service is running and models are loaded.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": 1681234567,
  "models_loaded": true
}
```

### GET /categories

Returns all available categories with their descriptions.

**Response:**
```json
{
  "success": true,
  "categories": [
    {"id": "technology", "name": "Technology"},
    {"id": "health", "name": "Health"},
    ...
  ]
}
```

### POST /predict/category

Analyzes text content and predicts the most appropriate category.

**Request:**
```json
{
  "content": "I've been learning about machine learning and neural networks for the past few months and it's fascinating how these technologies are transforming industries."
}
```

**Response:**
```json
{
  "category": "technology",
  "confidence": 0.87,
  "all_categories": {
    "technology": 0.87,
    "science": 0.05,
    ...
  }
}
```

## Technical Implementation

### Model Architecture

The thread categorization model uses a deep neural network architecture:
- Input layer with text tokenization
- Embedding layer
- Bidirectional LSTM layers
- Dense output layer with softmax activation

### Training Data

The model was trained on a dataset of categorized social media posts, carefully labeled to represent the 11 supported categories.

### Preprocessing

1. Text is tokenized using a custom tokenizer
2. Sequences are padded to a fixed length (100 tokens)
3. Special characters and formatting are normalized

### Deployment

The model is deployed using TensorFlow Serving, with the Flask app providing the API interface.

## Integration with Thread Creation

The AI service integrates with the thread creation workflow:

1. As the user types content in the thread composer, the frontend makes debounced API calls to the suggestion service
2. Suggestions appear in real-time with confidence indicators
3. High-confidence suggestions (>70%) are auto-selected
4. Users can override suggestions by manually selecting categories
5. The selected category is included when the thread is submitted to the backend

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Port for the Flask server | 5000 |
| FLASK_ENV | Environment (development/production) | development |
| AI_DEBUG_MODE | Enable simplified category matching | false |
| CORS_ORIGIN | Allowed CORS origin | http://localhost:3000 |
| LOG_LEVEL | Logging level | INFO |

## Development

### Requirements

- Python 3.8+
- TensorFlow 2.x
- Flask
- Additional requirements in requirements.txt

### Local Setup

1. Create a virtual environment: `python -m venv venv`
2. Activate the environment: `source venv/bin/activate` (Linux/Mac) or `venv\Scripts\activate` (Windows)
3. Install dependencies: `pip install -r requirements.txt`
4. Run the development server: `python run_local.py`

### Docker

Build and run with Docker:
```
docker build -t aycom-ai-service .
docker run -p 5000:5000 aycom-ai-service
``` 