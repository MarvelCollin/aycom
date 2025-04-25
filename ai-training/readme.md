# News Category Classification

This project implements a neural network model for classifying news articles into 4 categories.

## Dataset

The dataset consists of news articles categorized into 4 classes:
1. World (Class Index 1)
2. Sports (Class Index 2)
3. Business (Class Index 3)
4. Science/Technology (Class Index 4)

### Files
- `train.csv`: 120,000 news articles for training
- `test.csv`: 7,600 news articles for testing

Each file has three columns:
- `Class Index`: The category label (1-4)
- `Title`: The title of the news article
- `Description`: The content or description of the article

## Model Architecture

The model uses a deep learning approach with:
- Word embeddings to represent textual data
- LSTM (Long Short-Term Memory) neural network for sequence analysis
- Spatial Dropout for regularization
- Dense layer with softmax activation for classification

## Implementation Details

- Text preprocessing: Title and Description are combined into a single "content" field
- Tokenization: Text is converted to numerical sequences
- The model achieves approximately 90% accuracy on the test set
- Key libraries: TensorFlow, Keras, pandas, numpy, matplotlib

## Usage

### Requirements

```
tensorflow>=2.0.0
pandas
numpy
matplotlib
scikit-learn
```

### Training the Model

Run the Jupyter notebook `category_prediction.ipynb` to:
1. Load and preprocess the data
2. Train the model
3. Evaluate performance
4. Save the trained model

### Making Predictions

```python
def predict_thread_category(text_content):
    # Load the saved model
    loaded_model = tf.keras.models.load_model('thread_category_model.h5')
    
    # Load the tokenizer
    with open('tokenizer.pickle', 'rb') as handle:
        loaded_tokenizer = pickle.load(handle)
    
    # Preprocess and predict
    sequence = loaded_tokenizer.texts_to_sequences([text_content])
    padded = pad_sequences(sequence, maxlen=100, padding='post', truncating='post')
    prediction = loaded_model.predict(padded)[0]
    
    # Get the predicted class
    predicted_class = np.argmax(prediction) + 1
    category_name = {1: "World", 2: "Sports", 3: "Business", 4: "Science/Technology"}[predicted_class]
    
    return category_name, float(prediction[predicted_class-1]) * 100
```

## Files in this Repository

- `category_prediction.ipynb`: Jupyter notebook with the full model implementation
- `train.csv`: Training dataset
- `test.csv`: Testing dataset
- `thread_category_model.h5`: Saved model file
- `tokenizer.pickle`: Saved tokenizer for text preprocessing
- `requirements.txt`: Required Python packages
- `readme.md`: This documentation file