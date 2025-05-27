package docs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/swaggo/swag"
)

var SwaggerInfo = &swag.Spec{
	Version:     "1.0",
	Host:        "localhost:8083",
	BasePath:    "/api/v1",
	Schemes:     []string{"http", "https"},
	Title:       "AYCOM API",
	Description: "This is the API Gateway for the AYCOM platform.",
}

var doc = `{
    "swagger": "2.0",
    "info": {
        "description": "This is the API Gateway for the AYCOM platform.",
        "title": "AYCOM API",
        "contact": {},
        "version": "1.0"
    },
    "host": "localhost:8081",
    "basePath": "/api/v1",
    "paths": {
        "/auth/login": {
            "post": {
                "description": "Authenticates a user and returns tokens",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "User login",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.AuthServiceResponse"
                        }
                    }
                }
            }
        },
        "/trends": {
            "get": {
                "description": "Returns trending topics",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Trends"
                ],
                "summary": "Get trends",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "trends": {
                                    "type": "array",
                                    "items": {
                                        "$ref": "#/definitions/handlers.Trend"
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "handlers.AuthServiceResponse": {
            "type": "object",
            "properties": {
                "access_token": {
                    "type": "string"
                },
                "expires_in": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                },
                "refresh_token": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "token_type": {
                    "type": "string"
                },
                "user": {
                    "type": "object"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "handlers.Trend": {
            "type": "object",
            "properties": {
                "category": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "post_count": {
                    "type": "integer"
                },
                "title": {
                    "type": "string"
                }
            }
        }
    }
}`

type s struct{}

func (s *s) ReadDoc() string {

	jsonPath := filepath.Join("docs", "swagger.json")
	if _, err := os.Stat(jsonPath); err == nil {
		if data, err := ioutil.ReadFile(jsonPath); err == nil {
			return string(data)
		}
	}

	return doc
}

func init() {
	swag.Register(swag.Name, &s{})
}
