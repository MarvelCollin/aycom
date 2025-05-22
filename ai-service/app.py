import os
import logging
import pickle
import time
import numpy as np
from flask import Flask, request, jsonify
from flask_cors import CORS
from tensorflow.keras.models import load_model
from tensorflow.keras.preprocessing.sequence import pad_sequences
import pandas as pd
import json

# Configure logging
logging.basicConfig(
    level=os.environ.get("LOG_LEVEL", "INFO"),
    format="%(asctime)s [%(levelname)s] - %(message)s",
)
logger = logging.getLogger(__name__)

app = Flask(__name__)
# Configure CORS to allow credentials and specify origins
CORS(app, resources={r"/*": {"origins": os.environ.get("CORS_ORIGIN", "http://localhost:3000"), "supports_credentials": True}})

# Global variables for models
thread_model = None
tokenizer = None
max_sequence_length = 100  # Max length for input sequences

# Category definitions
categories = {
    "technology": "Technology related topics like computers, software, hardware, programming, AI, algorithms, apps, and digital innovations.",
    "health": "Health topics including fitness, diet, exercise, medical information, diseases, cures, treatments, and wellness.",
    "education": "Education related discussions about schools, universities, learning, studying, colleges, academics, courses, and student life.",
    "entertainment": "Entertainment topics like movies, music, games, plays, concerts, shows, actors, films, and media series.",
    "science": "Scientific topics including research, experiments, labs, scientists, discoveries, physics, chemistry, and biology.",
    "sports": "Sports related content about football, basketball, soccer, games, teams, players, matches, and tournaments.",
    "politics": "Political discussions including government, elections, voting, policies, presidents, congress, and political parties.",
    "business": "Business topics about companies, markets, finance, economy, stocks, investments, startups, and entrepreneurship.",
    "lifestyle": "Lifestyle content about homes, decor, fashion, trends, design, style, living, and clothing.",
    "travel": "Travel related topics including vacations, tours, destinations, hotels, flights, trips, journeys, and tourism.",
    "other": "General topics that don't fit into the other specific categories."
}

# Mapping from numeric output to category names
label_mapping = {
    0: "technology",
    1: "health", 
    2: "education", 
    3: "entertainment",
    4: "science",
    5: "sports",
    6: "politics",
    7: "business",
    8: "lifestyle",
    9: "travel",
    10: "other"
}

def load_models():
    """Load the pre-trained TensorFlow model and tokenizer"""
    global thread_model, tokenizer
    
    try:
        model_paths = [
            os.path.join(os.path.dirname(__file__), "thread_category_model.h5"),
            os.path.join(os.path.dirname(__file__), "models", "thread_category_model.h5"),
            "/app/models/thread_category_model.h5"
        ]
        
        model_loaded = False
        for model_path in model_paths:
            if os.path.exists(model_path):
                logger.info(f"Loading model from {model_path}")
                try:
                    thread_model = load_model(model_path, compile=False)
                    model_loaded = True
                    break
                except ValueError as e:
                    if 'batch_shape' in str(e):
                        # Handle the specific batch_shape error
                        logger.warning("Handling batch_shape error with custom objects")
                        from tensorflow.keras.layers import InputLayer
                        thread_model = load_model(
                            model_path, 
                            compile=False,
                            custom_objects={'InputLayer': InputLayer}
                        )
                        model_loaded = True
                        break
                    else:
                        logger.warning(f"Error loading model from {model_path}: {e}")
                except Exception as e:
                    logger.warning(f"Error loading model from {model_path}: {e}")
                    
        if not model_loaded:
            logger.error("Could not load model from any path")
            return False
                
        thread_model.compile(optimizer='adam', loss='categorical_crossentropy', metrics=['accuracy'])
        
        # Check multiple locations for the tokenizer file
        tokenizer_paths = [
            os.path.join(os.path.dirname(__file__), "tokenizer.pickle"),
            os.path.join(os.path.dirname(__file__), "models", "tokenizer.pickle"),
            "/app/models/tokenizer.pickle"
        ]
        
        # Try each path until a valid one is found
        tokenizer_loaded = False
        for tokenizer_path in tokenizer_paths:
            if os.path.exists(tokenizer_path):
                logger.info(f"Loading tokenizer from {tokenizer_path}")
                try:
                    with open(tokenizer_path, 'rb') as handle:
                        tokenizer = pickle.load(handle)
                    tokenizer_loaded = True
                    break
                except Exception as e:
                    logger.warning(f"Error loading tokenizer from {tokenizer_path}: {e}")
        
        if not tokenizer_loaded:
            logger.error("Could not load tokenizer from any path")
            return False
            
        logger.info("Model and tokenizer loaded successfully")
        return True
    except Exception as e:
        logger.error(f"Error loading models: {e}")
        return False

@app.route("/health", methods=["GET"])
def health_check():
    """Health check endpoint"""
    global thread_model, tokenizer
    
    status = {
        "status": "healthy",
        "timestamp": time.time(),
        "models_loaded": thread_model is not None and tokenizer is not None
    }
    
    logger.info(f"Health check: {status}")
    return jsonify(status)

@app.route("/categories", methods=["GET"])
def get_categories():
    """Return categories for the application"""
    # Convert category dict to list format for API
    categories_list = [
        {"id": category_id, "name": category_id.capitalize()} 
        for category_id in categories.keys()
    ]
    
    return jsonify({
        "success": True,
        "categories": categories_list
    })

@app.route("/predict/category", methods=["POST"])
def predict_category():
    """Predict the category of content using the pre-trained model"""
    global thread_model, tokenizer
    
    try:
        data = request.json
        if not data or 'content' not in data:
            return jsonify({"error": "Missing content field"}), 400
            
        content = data['content']
        logger.info(f"Received category prediction request for: {content[:50]}...")
        
        # Check if models are loaded
        if thread_model is None or tokenizer is None:
            success = load_models()
            if not success:
                return jsonify({"error": "Failed to load prediction models"}), 500
        
        # Preprocess the input text
        sequence = tokenizer.texts_to_sequences([content])
        padded = pad_sequences(sequence, maxlen=max_sequence_length, padding='post')
        
        # Make prediction
        prediction = thread_model.predict(padded)[0]
        
        # Create dictionary of category confidences
        category_scores = {label_mapping[i]: float(score) for i, score in enumerate(prediction)}
        
        # Find highest scoring category
        top_category = max(category_scores, key=category_scores.get)
        top_confidence = category_scores[top_category]
        
        # Return only the category and confidence, not all categories
        result = {
            "category": top_category,
            "confidence": float(top_confidence)
        }
        
        logger.info(f"Prediction result: {top_category} with confidence {top_confidence:.4f}")
        return jsonify(result)
            
    except Exception as e:
        logger.error(f"Error during prediction: {e}")
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    # Load models at startup
    load_models()
    
    # Get configuration from environment
    port = int(os.environ.get("PORT", 5000))
    debug = os.environ.get("FLASK_ENV") == "development"
    
    logger.info(f"Starting AI service on port {port}, debug={debug}")
    app.run(host="0.0.0.0", port=port, debug=debug)