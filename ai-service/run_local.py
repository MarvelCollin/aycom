#!/usr/bin/env python
"""
Local development server for the AI service.
Run this script to start the AI service locally instead of in Docker.
"""

import os
import sys
import subprocess

# Set environment variables
os.environ["PORT"] = "5000"
os.environ["FLASK_ENV"] = "development"
os.environ["LOG_LEVEL"] = "INFO"
os.environ["CORS_ORIGIN"] = "http://localhost:3000"
os.environ["TF_CPP_MIN_LOG_LEVEL"] = "2"  # Suppress TensorFlow logs
os.environ["PYTHONDONTWRITEBYTECODE"] = "1"
os.environ["AI_DEBUG_MODE"] = "false"

def check_dependencies():
    """Check if required Python packages are installed."""
    try:
        with open("requirements.txt", "r") as f:
            requirements = f.read().splitlines()
        
        required_packages = []
        for req in requirements:
            if req and not req.startswith("#"):
                package = req.split("==")[0] if "==" in req else req
                required_packages.append(package)
        
        print("Checking required packages...")
        missing_packages = []
        
        # Check each package
        for package in required_packages:
            try:
                __import__(package.replace("-", "_"))
                print(f"✅ {package}")
            except ImportError:
                missing_packages.append(package)
                print(f"❌ {package}")
        
        if missing_packages:
            print("\nMissing packages. Installing...")
            subprocess.check_call([sys.executable, "-m", "pip", "install", "-r", "requirements.txt"])
            print("Dependencies installed successfully.")
        else:
            print("\nAll dependencies are installed.")
            
    except Exception as e:
        print(f"Error checking dependencies: {e}")
        return False
    
    return True

def start_server():
    """Start the Flask development server."""
    try:
        print("\n🚀 Starting AI service locally on port 5000...")
        subprocess.run([sys.executable, "app.py"])
    except KeyboardInterrupt:
        print("\n🛑 Stopping AI service...")
    except Exception as e:
        print(f"\n❌ Error starting AI service: {e}")

if __name__ == "__main__":
    print("=" * 50)
    print("   🤖 AI Service - Local Development Server 🤖")
    print("=" * 50)
    print("\nThis script starts the AI service locally, outside of Docker.")
    print("The API Gateway in Docker will connect to this local instance.")
    
    if check_dependencies():
        start_server() 