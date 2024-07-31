// Package docs Code generated by swaggo/swag. DO NOT EDIT
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
        "/account": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get account details",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Get account details",
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "object"
                        }
                    }
                }
            }
        },
        "/analytics/analysis-data": {
            "get": {
                "description": "Get the analysis data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "AnalysisData"
                ],
                "summary": "Get the analysis data",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.AnalysisData"
                        }
                    }
                }
            }
        },
        "/analytics/download-market-data": {
            "get": {
                "description": "Download the market data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "DownloadMarketData"
                ],
                "summary": "Download the market data",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Start Date",
                        "name": "startDate",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "End Date",
                        "name": "endDate",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/analytics/meeting-runners": {
            "get": {
                "description": "Get the meeting runners",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GetMeetingRunners"
                ],
                "summary": "Get the meeting runners",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Event Name",
                        "name": "event_name",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.TodayRunners"
                        }
                    }
                }
            }
        },
        "/analytics/save-market-data": {
            "get": {
                "description": "Save the market data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "SaveMarketData"
                ],
                "summary": "Save the market data",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.MarketData"
                        }
                    }
                }
            }
        },
        "/analytics/today-meeting": {
            "get": {
                "description": "Get the today meeting",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GetTodayMeeting"
                ],
                "summary": "Get the today meeting",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/login": {
            "post": {
                "description": "Login",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Login",
                "parameters": [
                    {
                        "description": "User credentials",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.SignIn"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "object"
                        }
                    }
                }
            }
        },
        "/logout": {
            "get": {
                "description": "Logout",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Logout",
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "object"
                        }
                    }
                }
            }
        },
        "/register": {
            "post": {
                "description": "Register",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Register",
                "parameters": [
                    {
                        "description": "User credentials",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.SignUp"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "object"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "models.AnalysisData": {
            "type": "object",
            "properties": {
                "bsp": {
                    "type": "number"
                },
                "created_at": {
                    "type": "string"
                },
                "event_dt": {
                    "type": "string"
                },
                "event_id": {
                    "type": "integer"
                },
                "event_name": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "ip_traded_vol": {
                    "type": "number"
                },
                "ipmax": {
                    "type": "number"
                },
                "ipmin": {
                    "type": "number"
                },
                "menu_hint": {
                    "type": "string"
                },
                "morning_traded_vol": {
                    "type": "number"
                },
                "morning_wap": {
                    "type": "number"
                },
                "pp_traded_vol": {
                    "type": "number"
                },
                "ppmax": {
                    "type": "number"
                },
                "ppmin": {
                    "type": "number"
                },
                "ppwap": {
                    "type": "number"
                },
                "selection_id": {
                    "type": "integer"
                },
                "selection_name": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                },
                "win_lose": {
                    "type": "string"
                },
                "win_lose_float": {
                    "type": "number"
                }
            }
        },
        "models.MarketData": {
            "type": "object",
            "properties": {
                "bsp": {
                    "type": "number"
                },
                "created_at": {
                    "type": "string"
                },
                "event_dt": {
                    "type": "string"
                },
                "event_id": {
                    "type": "integer"
                },
                "event_name": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "ip_traded_vol": {
                    "type": "number"
                },
                "ipmax": {
                    "type": "number"
                },
                "ipmin": {
                    "type": "number"
                },
                "menu_hint": {
                    "type": "string"
                },
                "morning_traded_vol": {
                    "type": "number"
                },
                "morning_wap": {
                    "type": "number"
                },
                "pp_traded_vol": {
                    "type": "number"
                },
                "ppmax": {
                    "type": "number"
                },
                "ppmin": {
                    "type": "number"
                },
                "ppwap": {
                    "type": "number"
                },
                "selection_id": {
                    "type": "integer"
                },
                "selection_name": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                },
                "win_lose": {
                    "type": "string"
                },
                "win_lose_float": {
                    "type": "number"
                }
            }
        },
        "models.SignIn": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "models.SignUp": {
            "type": "object",
            "required": [
                "email",
                "full_name",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "full_name": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "password": {
                    "type": "string"
                },
                "phone_number": {
                    "type": "string"
                }
            }
        },
        "models.TodayRunners": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "event_date": {
                    "type": "string"
                },
                "event_name": {
                    "type": "string"
                },
                "event_time": {
                    "type": "string"
                },
                "horse_name": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "price": {
                    "type": "string"
                },
                "selection_id": {
                    "type": "integer"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
