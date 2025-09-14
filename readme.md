## 📋 Architecture

### Backend Services

AYCOM follows a microservices architecture with the following components:

- **API Gateway** - Central entry point for all client requests, handles routing and authentication
- **User Service** - Handles user authentication, profiles, social connections, and blocking functionality
- **Thread Service** - Manages threads, replies, media, and content interactions (likes, bookmarks)
- **Community Service** - Provides community management, membership, and chat functionality
- **Event Bus** - Handles asynchronous communication between services using RabbitMQ
- **AI Service** - Provides AI-based content categorization and analysis with TensorFlow

### Frontend

- Built with Svelte for reactive UI components
- TypeScript for type-safety
- Modern responsive design with custom theme support (dark/light mode)
- WebSocket integration for real-time notifications and chat
- Responsive layout suitable for mobile and desktop devices

### Infrastructure

- Docker containerization for consistent deployment across environments
- PostgreSQL databases (separate for each service to ensure independence)
- Redis for caching, session management, and real-time features
- RabbitMQ for asynchronous messaging and event broadcasting
- TensorFlow for AI content categorization capabilities

### Backend
- **Language**: Go (Golang) for high-performance microservices
- **Communication**: gRPC for efficient inter-service communication
- **Database**: PostgreSQL (separate instances for each service)
- **Caching**: Redis for performance optimization and WebSocket support
- **Message Queue**: RabbitMQ for event-driven architecture
- **Authentication**: JWT-based token system with refresh capabilities
- **API Documentation**: Swagger/OpenAPI

### Frontend
- **Framework**: Svelte for reactive UI components
- **Language**: TypeScript for type safety
- **Styling**: Modern CSS with custom theming system (dark/light mode)
- **API Client**: Custom fetch wrapper with standardized error handling
- **State Management**: Svelte stores for global state
- **WebSockets**: Real-time notifications and chat functionality

### AI Components
- **API**: Flask-based REST API
- **ML Framework**: TensorFlow/Keras for machine learning models
- **NLP**: Natural language processing for content categorization
- **Model Training**: Jupyter notebooks for model development and training
- **Model Deployment**: Containerized inference service
