@echo off
echo ================================================
echo      AI Service - Local Development Server
echo ================================================
echo.
echo This script starts the AI service locally on Windows.
echo.

:: Set environment variables
set PORT=5000
set FLASK_ENV=development
set LOG_LEVEL=INFO
set CORS_ORIGIN=http://localhost:3000
set TF_CPP_MIN_LOG_LEVEL=2
set PYTHONDONTWRITEBYTECODE=1
set AI_DEBUG_MODE=false

:: Check if Python is installed
python --version > nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Python is not installed or not in PATH. Please install Python.
    exit /b 1
)

:: Check for virtual environment
if exist venv (
    echo Using existing virtual environment...
    call venv\Scripts\activate
) else (
    echo Creating new virtual environment...
    python -m venv venv
    call venv\Scripts\activate
    echo Installing dependencies...
    pip install -r requirements.txt
)

echo.
echo Starting AI service on port 5000...
echo Press Ctrl+C to stop the server.
echo.

:: Start the server
python app.py

:: Deactivate virtual environment when done
call venv\Scripts\deactivate 