import os
import logging
import pickle
import time
import numpy as np
import re


os.environ['PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION'] = 'python'

from flask import Flask, request, jsonify
from flask_cors import CORS
from tensorflow.keras.models import load_model, Sequential
from tensorflow.keras.preprocessing.sequence import pad_sequences
from tensorflow.keras.preprocessing.text import Tokenizer
from tensorflow.keras.layers import Dense, Embedding, LSTM, SpatialDropout1D


logging.basicConfig(
    level=os.environ.get("LOG_LEVEL", "INFO"),
    format="%(asctime)s [%(levelname)s] - %(message)s",
)
logger = logging.getLogger(__name__)

app = Flask(__name__)

CORS(app, resources={r"/*": {"origins": os.environ.get("CORS_ORIGIN", "http://localhost:3000")}})


thread_model = None
tokenizer = None
max_sequence_length = 100
using_fresh_model = False


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


label_mapping = {
    0: "technology", 1: "health", 2: "education", 3: "entertainment",
    4: "science", 5: "sports", 6: "politics", 7: "business",
    8: "lifestyle", 9: "travel", 10: "other"
}


category_keywords = {
    "technology": ["technology", "tech", "computer", "software", "hardware", "app", "code", "program", "internet", "smartphone", "device", "iphone", "android", "web", "digital", "online", "system", "electronic", "gadget", "mobile", "laptop", "tablet", "ai", "artificial intelligence", "machine learning"],
    "health": ["health", "medical", "doctor", "hospital", "disease", "treatment", "medicine", "patient", "cure", "symptom", "healthcare", "fitness", "workout", "exercise", "diet", "nutrition", "wellness", "therapy", "drug", "vaccine", "virus", "covid", "pandemic", "mental health"],
    "education": ["education", "school", "student", "teacher", "learn", "course", "study", "college", "university", "degree", "academic", "research", "knowledge", "skill", "class", "lecture", "lesson", "exam", "test", "assignment", "homework", "grade", "campus"],
    "entertainment": ["entertainment", "movie", "film", "music", "song", "concert", "artist", "actor", "actress", "celebrity", "show", "tv", "television", "series", "episode", "play", "theater", "performance", "comedy", "drama", "game", "gaming", "video game", "streaming"],
    "science": ["science", "scientific", "scientist", "research", "discovery", "experiment", "physics", "chemistry", "biology", "astronomy", "space", "planet", "star", "universe", "laboratory", "theory", "hypothesis", "evidence", "data", "climate", "environment", "nature", "animal", "species"],
    "sports": ["sports", "game", "team", "player", "coach", "match", "tournament", "championship", "league", "score", "win", "loss", "competition", "athlete", "olympic", "football", "soccer", "basketball", "baseball", "tennis", "golf", "swimming", "boxing", "cricket"],
    "politics": ["politics", "political", "government", "president", "minister", "election", "vote", "campaign", "law", "policy", "congress", "senate", "representative", "democracy", "republican", "democrat", "party", "bill", "legislation", "regulation", "court", "justice", "decision"],
    "business": ["business", "company", "corporation", "firm", "industry", "market", "product", "service", "customer", "client", "stock", "share", "investor", "profit", "loss", "revenue", "sales", "investment", "finance", "money", "bank", "credit", "loan", "debt", "economy"],
    "lifestyle": ["lifestyle", "life", "living", "home", "house", "apartment", "furniture", "decoration", "design", "style", "fashion", "clothes", "dress", "outfit", "trend", "family", "relationship", "marriage", "wedding", "divorce", "parent", "child", "baby", "pet"],
    "travel": ["travel", "trip", "journey", "vacation", "holiday", "tour", "tourist", "tourism", "destination", "hotel", "resort", "flight", "airplane", "airport", "country", "city", "beach", "mountain", "adventure", "exploration", "sightseeing", "landmark", "passport", "visa"],
    "other": ["other", "miscellaneous", "various", "different", "random", "general"]
}

def create_fresh_model(vocab_size=20000, embedding_dim=128, num_classes=11):
    """Create a fresh model with the same architecture as in the notebook"""
    logger.info(f"Creating a fresh model with vocab_size={vocab_size}, embedding_dim={embedding_dim}, num_classes={num_classes}")
    
    model = Sequential([
        Embedding(input_dim=vocab_size, output_dim=embedding_dim, input_length=max_sequence_length),
        SpatialDropout1D(0.2),
        LSTM(embedding_dim, dropout=0.2, recurrent_dropout=0.2),
        Dense(num_classes, activation='softmax')
    ])
    
    model.compile(
        loss='categorical_crossentropy',
        optimizer='adam',
        metrics=['accuracy']
    )
    
    logger.info("Fresh model created successfully")
    return model

def create_fresh_tokenizer(num_words=20000):
    """Create a new tokenizer if we can't load the saved one"""
    logger.info(f"Creating a fresh tokenizer with num_words={num_words}")
    
    
    new_tokenizer = Tokenizer(num_words=num_words, oov_token="<OOV>")
    
    
    
    common_words = [
        "the", "be", "to", "of", "and", "a", "in", "that", "have", "I",
        "it", "for", "not", "on", "with", "he", "as", "you", "do", "at",
        "technology", "health", "education", "entertainment", "science", 
        "sports", "politics", "business", "lifestyle", "travel"
    ]
    
    
    new_tokenizer.fit_on_texts([" ".join(common_words)])
    
    logger.info(f"Fresh tokenizer created with {len(new_tokenizer.word_index)} words")
    return new_tokenizer

def load_models():
    """Load the pre-trained TensorFlow model and tokenizer"""
    global thread_model, tokenizer, using_fresh_model

    try:
        
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

        
        model_loaded = False
        for model_path in possible_model_paths:
            try:
                if os.path.exists(model_path):
                    logger.info(f"Found model at {model_path}, attempting to load")
                    
                    try:
                        thread_model = load_model(model_path, compile=False)
                    except:
                        thread_model = load_model(model_path, compile=False, custom_objects={})
                    
                    logger.info(f"Successfully loaded model from {model_path}")
                    model_loaded = True
                    break
            except Exception as e:
                logger.info(f"Error loading model from {model_path}: {e}")
                continue
                
        
        if not model_loaded:
            logger.info("Could not load model from disk, creating a fresh model")
            thread_model = create_fresh_model()
            using_fresh_model = True
                
        
        thread_model.compile(optimizer='adam', loss='categorical_crossentropy', metrics=['accuracy'])

        
        tokenizer_loaded = False
        for tokenizer_path in possible_tokenizer_paths:
            try:
                if os.path.exists(tokenizer_path):
                    logger.info(f"Found tokenizer at {tokenizer_path}, attempting to load")
                    with open(tokenizer_path, 'rb') as handle:
                        tokenizer = pickle.load(handle)
                    logger.info(f"Successfully loaded tokenizer from {tokenizer_path}")
                    tokenizer_loaded = True
                    break
            except Exception as e:
                logger.info(f"Error loading tokenizer from {tokenizer_path}: {e}")
                continue
                
        
        if not tokenizer_loaded:
            logger.info("Could not load tokenizer from any location, creating a fresh one")
            tokenizer = create_fresh_tokenizer()
            using_fresh_model = True

        logger.info("Model and tokenizer loaded successfully")
        return True
    except Exception as e:
        logger.error(f"Error loading models: {e}")
        return False

def predict_by_keywords(content):
    """Use keyword matching to predict category if we're using a fresh model"""
    logger.info(f"Using keyword-based prediction for: {content}")
    content = content.lower()
    
    
    category_scores = {}
    for category, keywords in category_keywords.items():
        match_count = 0
        for keyword in keywords:
            
            if re.search(r'\b' + re.escape(keyword) + r'\b', content):
                match_count += 1
        
        
        
        if match_count > 0:
            category_scores[category] = float(match_count) / len(keywords) * 0.8 + 0.1
        else:
            category_scores[category] = 0.02  
    
    
    if not category_scores:
        return {"category": "other", "confidence": 0.1}
    
    top_category = max(category_scores, key=category_scores.get)
    top_confidence = category_scores[top_category]
    
    return {"category": top_category, "confidence": top_confidence}

@app.route("/health", methods=["GET"])
def health_check():
    """Health check endpoint"""
    global thread_model, tokenizer

    status = {
        "status": "healthy",
        "timestamp": time.time(),
        "models_loaded": thread_model is not None and tokenizer is not None,
        "using_fresh_model": using_fresh_model
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
    global thread_model, tokenizer, using_fresh_model

    try:
        data = request.json
        if not data or 'content' not in data:
            return jsonify({"error": "Missing content field"}), 400

        content = data['content']

        
        if thread_model is None or tokenizer is None:
            success = load_models()
            if not success:
                return jsonify({"error": "Failed to load prediction models"}), 500

        
        if using_fresh_model:
            result = predict_by_keywords(content)
            return jsonify(result)

        
        
        sequence = tokenizer.texts_to_sequences([content])
        padded = pad_sequences(sequence, maxlen=max_sequence_length, padding='post')

        
        prediction = thread_model.predict(padded)[0]

        
        category_scores = {label_mapping[i]: float(score) for i, score in enumerate(prediction)}

        
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
    
    load_models()

    
    port = int(os.environ.get("PORT", 5000))
    debug = os.environ.get("FLASK_ENV") == "development"

    logger.info(f"Starting AI service on port {port}")
    app.run(host="0.0.0.0", port=port, debug=debug)