// Package docs provides API documentation for the AYCOM API Gateway.
//
// Documentation for the AYCOM API Gateway
//
//	Schemes: http, https
//	Host: localhost:8080
//	BasePath: /api/v1
//	Version: 1.0.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Security:
//	- bearer
//
//	SecurityDefinitions:
//	bearer:
//	     type: apiKey
//	     name: Authorization
//	     in: header
package docs

import "github.com/swaggo/swag"

var (
	// SwaggerInfo holds exported Swagger Info so clients can modify it
	SwaggerInfo = &swag.Spec{
		Version:          "1.0",
		Host:             "localhost:8080",
		BasePath:         "/api/v1",
		Schemes:          []string{"http", "https"},
		Title:            "AYCOM API Gateway",
		Description:      "API Gateway for AYCOM microservices",
		InfoInstanceName: "swagger",
		SwaggerTemplate:  docTemplate,
	}
)

const docTemplate = `{
  "swagger": "2.0",
  "info": {
    "description": "{{.Description}}",
    "title": "{{.Title}}",
    "contact": {
      "name": "API Support",
      "url": "https://github.com/your-username/aycom",
      "email": "support@example.com"
    },
    "license": {
      "name": "MIT",
      "url": "https://opensource.org/licenses/MIT"
    },
    "version": "{{.Version}}"
  },
  "host": "{{.Host}}",
  "basePath": "{{.BasePath}}",
  "paths": {
    "/health": {
      "get": {
        "description": "Get the status of the API",
        "consumes": ["application/json"],
        "produces": ["application/json"],
        "tags": ["health"],
        "summary": "Health check endpoint",
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "additionalProperties": true
            }
          }
        }
      }
    }
  },
  "securityDefinitions": {
    "BearerAuth": {
      "type": "apiKey",
      "name": "Authorization",
      "in": "header"
    }
  }
}`
