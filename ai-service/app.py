import os
import logging
import pickle
import time
import numpy as np
from flask import Flask, request, jsonify
from flask_cors import CORS
from tensorflow.keras.models import load_model

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

def load_models():
    """Load ML models at startup"""
    global thread_model, tokenizer
    
    try:
        # Load the thread categorization model
        model_path = os.path.join(os.path.dirname(__file__), 'thread_category_model.h5')
        tokenizer_path = os.path.join(os.path.dirname(__file__), 'tokenizer.pickle')
        
        logger.info(f"Loading model from {model_path}")
        thread_model = load_model(model_path)
        
        logger.info(f"Loading tokenizer from {tokenizer_path}")
        with open(tokenizer_path, 'rb') as handle:
            tokenizer = pickle.load(handle)
            
        logger.info("Models loaded successfully")
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

@app.route("/predict/category", methods=["POST"])
def predict_category():
    """Predict the category of a thread based on its content"""
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
                return jsonify({"error": "Models not available"}), 503
        
        # In a real application, preprocess the text and make a prediction
        # For now, returning a mock result
        categories = ["technology", "politics", "entertainment", "sports", "business"]
        confidence_scores = np.random.random(5)
        confidence_scores = confidence_scores / np.sum(confidence_scores)  # Normalize
        
        # Get the highest confidence category
        top_category_idx = np.argmax(confidence_scores)
        top_category = categories[top_category_idx]
        top_confidence = float(confidence_scores[top_category_idx])
        
        result = {
            "category": top_category,
            "confidence": top_confidence,
            "all_categories": {cat: float(score) for cat, score in zip(categories, confidence_scores)}
        }
        
        logger.info(f"Prediction result: {result}")
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