package openapi

func getOpenapiJSONRPCErrorSchemas() Schemas {
	return Schemas{
		"ParseError": {
			Type: "object",
			Properties: Properties{
				"jsonrpc": &Schema{
					Type:    "string",
					Example: "2.0",
				},
				"id": &Schema{
					Type:    "string",
					Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
				},
				"error": &Schema{
					Type: "object",
					Properties: Properties{
						"code": &Schema{
							Type:    "integer",
							Example: -32700,
						},
						"message": &Schema{
							Type:    "string",
							Example: "Parse error",
						},
					},
				},
			},
		},
		"InvalidRequestError": {
			Type: "object",
			Properties: Properties{
				"jsonrpc": &Schema{
					Type:    "string",
					Example: "2.0",
				},
				"id": &Schema{
					Type:    "string",
					Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
				},
				"error": &Schema{
					Type: "object",
					Properties: Properties{
						"code": &Schema{
							Type:    "integer",
							Example: -32600,
						},
						"message": &Schema{
							Type:    "string",
							Example: "Invalid Request",
						},
					},
				},
			},
		},
		"MethodNotFoundError": {
			Type: "object",
			Properties: Properties{
				"jsonrpc": &Schema{
					Type:    "string",
					Example: "2.0",
				},
				"id": &Schema{
					Type:    "string",
					Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
				},
				"error": &Schema{
					Type: "object",
					Properties: Properties{
						"code": &Schema{
							Type:    "integer",
							Example: -32601,
						},
						"message": &Schema{
							Type:    "string",
							Example: "Method not found",
						},
					},
				},
			},
		},
		"InvalidParamsError": {
			Type: "object",
			Properties: Properties{
				"jsonrpc": &Schema{
					Type:    "string",
					Example: "2.0",
				},
				"id": &Schema{
					Type:    "string",
					Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
				},
				"error": &Schema{
					Type: "object",
					Properties: Properties{
						"code": &Schema{
							Type:    "integer",
							Example: -32602,
						},
						"message": &Schema{
							Type:    "string",
							Example: "Invalid params",
						},
					},
				},
			},
		},
		"InternalError": {
			Type: "object",
			Properties: Properties{
				"jsonrpc": &Schema{
					Type:    "string",
					Example: "2.0",
				},
				"id": &Schema{
					Type:    "string",
					Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
				},
				"error": &Schema{
					Type: "object",
					Properties: Properties{
						"code": &Schema{
							Type:    "integer",
							Example: -32603,
						},
						"message": &Schema{
							Type:    "string",
							Example: "Internal error",
						},
					},
				},
			},
		},
	}
}

func makeOpenapiSchemaJRPCError(code int64) *Schema {
	return &Schema{
		Type: "object",
		Properties: Properties{
			"jsonrpc": &Schema{
				Type:    "string",
				Example: "2.0",
			},
			"id": &Schema{
				Type:    "string",
				Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
			},
			"error": &Schema{
				Type: "object",
				Properties: Properties{
					"code": &Schema{
						Type:    "integer",
						Example: code,
					},
					"message": &Schema{
						Type: "string",
					},
				},
			},
		},
	}
}

func makeOpenapiSchemaRESTError(errCode string) *Schema {
	return &Schema{
		Type: "object",
		Properties: Properties{
			"data": &Schema{
				Type: "object",
				Properties: Properties{
					"key": &Schema{AnyOf: []Schema{
						{Type: "string"},
						{Type: "number"},
						{Type: "integer"},
						{Type: "array"},
						{Type: "object"},
					}},
				},
			},
			"code": &Schema{
				Type:    "string",
				Example: errCode,
			},
			"error": &Schema{
				Type: "string",
			},
		},
	}
}
