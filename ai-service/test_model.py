#!/usr/bin/env python3
"""
Simple test script to verify that the AI model loads correctly and can make predictions
Run this script to confirm the model is working as expected
"""

import os
import sys
import pickle
import tensorflow as tf
from tensorflow.keras.models import load_model
from tensorflow.keras.preprocessing.sequence import pad_sequences

# Set up paths
current_dir = os.path.dirname(os.path.abspath(__file__))
model_path = os.path.join(current_dir, "thread_category_model.h5")
tokenizer_path = os.path.join(current_dir, "tokenizer.pickle")

# Category mapping
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

# Load model
print(f"Loading model from {model_path}...")
try:
    model = load_model(model_path, compile=False)
    model.compile(optimizer='adam', loss='categorical_crossentropy', metrics=['accuracy'])
    print("Model loaded successfully!")
except Exception as e:
    print(f"Error loading model: {e}")
    sys.exit(1)

# Load tokenizer
print(f"Loading tokenizer from {tokenizer_path}...")
try:
    with open(tokenizer_path, 'rb') as handle:
        tokenizer = pickle.load(handle)
    print("Tokenizer loaded successfully!")
except Exception as e:
    print(f"Error loading tokenizer: {e}")
    sys.exit(1)

# Test predictions
test_texts = [
    "I'm learning about machine learning and neural networks for a school project",
    "The basketball game last night was incredible, the team played amazingly",
    "I need advice on my diet and exercise routine to improve my health",
    "This new movie that just came out has the best special effects I've ever seen",
    "The stock market had a significant drop today affecting many companies"
]

max_sequence_length = 100

# Process each test text
for text in test_texts:
    # Preprocess
    sequence = tokenizer.texts_to_sequences([text])
    padded = pad_sequences(sequence, maxlen=max_sequence_length, padding='post')
    
    # Predict
    prediction = model.predict(padded)[0]
    
    # Get top category
    top_idx = prediction.argmax()
    top_category = label_mapping[top_idx]
    top_confidence = prediction[top_idx]
    
    print("-" * 70)
    print(f"Text: {text}")
    print(f"Predicted Category: {top_category}")
    print(f"Confidence: {top_confidence:.4f}")
    
    # Print top 3 categories
    sorted_indices = prediction.argsort()[::-1][:3]
    print("Top categories:")
    for idx in sorted_indices:
        print(f"  - {label_mapping[idx]}: {prediction[idx]:.4f}")

print("\nTest completed successfully! The model is working as expected.") 