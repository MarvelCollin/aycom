FROM tensorflow/tensorflow:2.13.0 AS builder

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    gcc \
    python3-dev \
    && apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ENV MAKEFLAGS="-j$(nproc)" \
    PIP_NO_CACHE_DIR=1 \
    PYTHONUNBUFFERED=1

WORKDIR /build
COPY requirements.txt .

RUN grep -v "tensorflow" requirements.txt > filtered_requirements.txt || true

RUN pip install --upgrade pip && \
    pip install wheel && \
    pip wheel --no-cache-dir -r filtered_requirements.txt -w /wheels

FROM tensorflow/tensorflow:2.13.0-slim

ENV PYTHONUNBUFFERED=1 \
    TF_CPP_MIN_LOG_LEVEL=2 \
    PYTHONDONTWRITEBYTECODE=1 \
    TF_ENABLE_ONEDNN_OPTS=1 \
    OMP_NUM_THREADS=4 \
    NUMBA_NUM_THREADS=4 \
    OPENBLAS_NUM_THREADS=4

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    && apt-get clean && \
    rm -rf /var/lib/apt/lists/*

    WORKDIR /app
COPY --from=builder /wheels /wheels

RUN pip install --no-cache-dir --no-index --find-links=/wheels /wheels/* && \
    rm -rf /wheels

COPY app.py thread_category_model.h5 tokenizer.pickle /app/

VOLUME /app/models

RUN groupadd -r appuser && useradd -r -g appuser appuser && \
    chown -R appuser:appuser /app
USER appuser

EXPOSE 5000

CMD echo "Starting app with optimized settings..." && \
    python -c "import tensorflow as tf; print('TensorFlow version:', tf.__version__); print('Available devices:', tf.config.list_physical_devices())" && \
    python app.py 