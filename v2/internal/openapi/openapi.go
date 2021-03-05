package openapi

type Contact struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	URL   string `yaml:"url,omitempty" json:"url,omitempty"`
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
}

type License struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	URL  string `yaml:"url,omitempty" json:"url,omitempty"`
}

type Info struct {
	Title          string   `yaml:"title,omitempty" json:"title,omitempty"`
	Description    string   `yaml:"description,omitempty" json:"description,omitempty"`
	TermsOfService string   `yaml:"termsOfService,omitempty" json:"termsOfService,omitempty"`
	Contact        *Contact `yaml:"contact,omitempty" json:"contact,omitempty"`
	License        *License `yaml:"license,omitempty" json:"license,omitempty"`
	Version        string   `yaml:"version,omitempty" json:"version,omitempty"`
}

type ExternalDocs struct {
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	URL         string `yaml:"url,omitempty" json:"url,omitempty"`
}

type Tag struct {
	Name         string       `yaml:"name,omitempty" json:"name,omitempty"`
	Description  string       `yaml:"description,omitempty" json:"description,omitempty"`
	ExternalDocs ExternalDocs `yaml:"externalDocs,omitempty" json:"externalDocs,omitempty"`
}

type Properties map[string]*Schema

type Schema struct {
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Ref         string      `yaml:"$ref,omitempty" json:"$ref,omitempty"`
	Type        string      `yaml:"type,omitempty" json:"type,omitempty"`
	Format      string      `yaml:"format,omitempty" json:"format,omitempty"`
	Properties  Properties  `yaml:"properties,omitempty" json:"properties,omitempty"`
	Items       *Schema     `yaml:"items,omitempty" json:"items,omitempty"`
	AnyOf       []Schema    `yaml:"anyOf,omitempty" json:"anyOf,omitempty"`
	Enum        []string    `yaml:"enum,omitempty" json:"enum,omitempty"`
	Example     interface{} `yaml:"example,omitempty" json:"example,omitempty"`
}

type Parameter struct {
	Ref         string  `yaml:"$ref,omitempty" json:"$ref,omitempty"`
	In          string  `yaml:"in,omitempty" json:"in,omitempty"`
	Name        string  `yaml:"name,omitempty" json:"name,omitempty"`
	Description string  `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool    `yaml:"required,omitempty" json:"required,omitempty"`
	Schema      *Schema `yaml:"schema,omitempty" json:"schema,omitempty"`
}

type Media struct {
	Schema *Schema `yaml:"schema,omitempty" json:"schema,omitempty"`
}

type Content map[string]Media

type Response struct {
	Description string  `yaml:"description,omitempty" json:"description,omitempty"`
	Content     Content `yaml:"content,omitempty" json:"content,omitempty"`
}

type Responses map[string]Response

type RequestBody struct {
	Description string  `yaml:"description,omitempty" json:"description,omitempty"`
	Content     Content `yaml:"content,omitempty" json:"content,omitempty"`
	Required    bool    `yaml:"required,omitempty" json:"required,omitempty"`
}

type Operation struct {
	Tags        []string     `yaml:"tags,omitempty" json:"tags,omitempty"`
	Summary     string       `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string       `yaml:"description,omitempty" json:"description,omitempty"`
	OperationID string       `yaml:"operationId,omitempty" json:"operationId,omitempty"`
	Consumes    []string     `yaml:"consumes,omitempty" json:"consumes,omitempty"`
	Produces    []string     `yaml:"produces,omitempty" json:"produces,omitempty"`
	Parameters  []Parameter  `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	RequestBody *RequestBody `yaml:"requestBody,omitempty" json:"requestBody,omitempty"`
	Responses   Responses    `yaml:"responses,omitempty" json:"responses,omitempty"`
}

type Path struct {
	Ref         string     `yaml:"$ref,omitempty" json:"$ref,omitempty"`
	Summary     string     `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string     `yaml:"description,omitempty" json:"description,omitempty"`
	Get         *Operation `yaml:"get,omitempty" json:"get,omitempty"`
	Post        *Operation `yaml:"post,omitempty" json:"post,omitempty"`
	Patch       *Operation `yaml:"patch,omitempty" json:"patch,omitempty"`
	Put         *Operation `yaml:"put,omitempty" json:"put,omitempty"`
	Delete      *Operation `yaml:"delete,omitempty" json:"delete,omitempty"`
}

type Variable struct {
	Enum        []string `yaml:"enum,omitempty" json:"enum,omitempty"`
	Default     string   `yaml:"default,omitempty" json:"default,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
}

type Server struct {
	URL         string              `yaml:"url,omitempty" json:"url,omitempty"`
	Description string              `yaml:"description,omitempty" json:"description,omitempty"`
	Variables   map[string]Variable `yaml:"variables,omitempty" json:"variables,omitempty"`
}

type Schemas map[string]*Schema

type Components struct {
	Schemas Schemas `yaml:"schemas,omitempty" json:"schemas,omitempty"`
}

type OpenAPI struct {
	OpenAPI    string           `yaml:"openapi" json:"openapi"`
	Info       Info             `yaml:"info,omitempty" json:"info,omitempty"`
	Servers    []Server         `yaml:"servers,omitempty" json:"servers,omitempty"`
	Tags       []Tag            `yaml:"tags,omitempty" json:"tags,omitempty"`
	Schemes    []string         `yaml:"schemes,omitempty" json:"schemes,omitempty"`
	Paths      map[string]*Path `yaml:"paths,omitempty" json:"paths,omitempty"`
	Components Components       `yaml:"components,omitempty" json:"components,omitempty"`
}
