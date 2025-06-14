import os
import logging
import pickle
import time
import numpy as np

# Fix for protobuf compatibility issues with TensorFlow
os.environ['PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION'] = 'python'

from flask import Flask, request, jsonify
from flask_cors import CORS
from tensorflow.keras.models import load_model
from tensorflow.keras.preprocessing.sequence import pad_sequences

# Configure logging
logging.basicConfig(
    level=os.environ.get("LOG_LEVEL", "INFO"),
    format="%(asctime)s [%(levelname)s] - %(message)s",
)
logger = logging.getLogger(__name__)

app = Flask(__name__)
# Configure CORS
CORS(app, resources={r"/*": {"origins": os.environ.get("CORS_ORIGIN", "http://localhost:3000")}})

# Global variables for models
thread_model = None
tokenizer = None
max_sequence_length = 100

# Category definitions
categories = {
    "technology": "Technology",
    "health": "Health",
    "education": "Education",
    "entertainment": "Entertainment",
    "science": "Science",
    "sports": "Sports",
    "politics": "Politics",
    "business": "Business",
    "lifestyle": "Lifestyle",
    "travel": "Travel",
    "other": "Other"
}

# Mapping from numeric output to category names
label_mapping = {
    0: "technology", 1: "health", 2: "education", 3: "entertainment",
    4: "science", 5: "sports", 6: "politics", 7: "business",
    8: "lifestyle", 9: "travel", 10: "other"
}

def load_models():
    """Load the pre-trained TensorFlow model and tokenizer"""
    global thread_model, tokenizer

    try:
        # Try all possible model locations
        possible_model_paths = [
            "/app/models/thread_category_model.h5",
            "/app/thread_category_model.h5",
            "./thread_category_model.h5",
            "./models/thread_category_model.h5"
        ]
        
        possible_tokenizer_paths = [
            "/app/models/tokenizer.pickle",
            "/app/tokenizer.pickle",
            "./tokenizer.pickle",
            "./models/tokenizer.pickle"
        ]
        
        # Handle the 'batch_shape' argument issue by adding it to custom_objects
        custom_objects = {'batch_shape': lambda x: x}

        # Try to load the model from any location
        for model_path in possible_model_paths:
            try:
                if os.path.exists(model_path):
                    logger.info(f"Found model at {model_path}, attempting to load")
                    thread_model = load_model(model_path, compile=False, custom_objects=custom_objects)
                    logger.info(f"Successfully loaded model from {model_path}")
                    break
            except Exception as e:
                logger.info(f"Error loading model from {model_path}: {e}")
                continue
                
        # If model is still None, we couldn't load it
        if thread_model is None:
            raise Exception("Could not load model from any location")

        # Compile the model
        thread_model.compile(optimizer='adam', loss='categorical_crossentropy', metrics=['accuracy'])

        # Try to load tokenizer from either location
        for tokenizer_path in possible_tokenizer_paths:
            try:
                if os.path.exists(tokenizer_path):
                    logger.info(f"Found tokenizer at {tokenizer_path}, attempting to load")
                    with open(tokenizer_path, 'rb') as handle:
                        tokenizer = pickle.load(handle)
                    logger.info(f"Successfully loaded tokenizer from {tokenizer_path}")
                    break
            except Exception as e:
                logger.info(f"Error loading tokenizer from {tokenizer_path}: {e}")
                continue
                
        # If tokenizer is still None, we couldn't load it
        if tokenizer is None:
            raise Exception("Could not load tokenizer from any location")

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

    return jsonify(status)

@app.route("/categories", methods=["GET"])
def get_categories():
    """Return categories for the application"""
    categories_list = [
        {"id": category_id, "name": name} 
        for category_id, name in categories.items()
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

        result = {
            "category": top_category,
            "confidence": float(top_confidence)
        }

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

    logger.info(f"Starting AI service on port {port}")
    app.run(host="0.0.0.0", port=port, debug=debug)