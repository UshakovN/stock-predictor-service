// Code generated by swaggo/swag. DO NOT EDIT.

package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "Health method check http server health",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Health"
                ],
                "summary": "Health check method",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/common.HealthResponse"
                        }
                    }
                }
            }
        },
        "/redirect": {
            "get": {
                "security": [
                    {
                        "HttpBasicAuth": []
                    }
                ],
                "description": "Redirect method made redirect to swagger documentation url for service",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Documentation"
                ],
                "summary": "Redirect to swagger doc method",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Service Name",
                        "name": "service",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "302": {
                        "description": "Found"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/errs.Error"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/errs.Error"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/errs.Error"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/errs.Error"
                        }
                    }
                }
            }
        },
        "/services": {
            "get": {
                "security": [
                    {
                        "HttpBasicAuth": []
                    }
                ],
                "description": "Service method provide names of services support swagger",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Documentation"
                ],
                "summary": "Services method for service names",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/internal_handler.ServicesResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/errs.Error"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/errs.Error"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/errs.Error"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/errs.Error"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "common.HealthResponse": {
            "type": "object",
            "properties": {
                "success": {
                    "type": "boolean"
                }
            }
        },
        "errs.Error": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "internal_handler.ServicesResponse": {
            "type": "object",
            "properties": {
                "service_names": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "success": {
                    "type": "boolean"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:8086",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "Swagger Service API",
	Description:      "API for single swagger service",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
