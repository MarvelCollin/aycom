import os
from flask import Flask, request, jsonify
from flask_cors import CORS # Import CORS
import logging

# Configure logging
logging.basicConfig(
    level=os.environ.get("LOG_LEVEL", "INFO"),
    format="%(asctime)s [%(levelname)s] - %(message)s",
)
logger = logging.getLogger(__name__)

app = Flask(__name__)
CORS(app, resources={r"/health": {"origins": "http://localhost:3000"}}) # Enable CORS for /health from frontend origin

@app.route("/health", methods=["GET"])
def health_check():
    logger.info("Health check endpoint accessed")
    return jsonify({"status": "healthy"})

@app.route("/predict", methods=["POST"])
def predict():
    try:
        data = request.json
        logger.info(f"Received prediction request: {data}")
        
        # Mock AI prediction
        # In a real application, this would call your ML model
        result = {"prediction": "example_prediction", "confidence": 0.95}
        
        logger.info(f"Prediction result: {result}")
        return jsonify(result)
    except Exception as e:
        logger.error(f"Error during prediction: {e}")
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    port = int(os.environ.get("PORT", 5000))
    debug = os.environ.get("FLASK_ENV") == "development"
    
    logger.info(f"Starting AI service on port {port}, debug={debug}")
    app.run(host="0.0.0.0", port=port, debug=debug)