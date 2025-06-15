FROM python:3.8-slim

ENV PYTHONUNBUFFERED=1 \
    TF_CPP_MIN_LOG_LEVEL=2 \
    PYTHONDONTWRITEBYTECODE=1 \
    PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION=python

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

RUN mkdir -p /app/models

COPY app.py /app/
COPY thread_category_model.h5 /app/models/
COPY tokenizer.pickle /app/models/

EXPOSE 5000

CMD ["python", "app.py"] 