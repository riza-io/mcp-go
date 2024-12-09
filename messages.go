package mcp

import "encoding/json"

type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// Capabilities represents the available feature capabilities
type ClientCapabilities struct {
	Roots    Roots    `json:"roots"`
	Sampling Sampling `json:"sampling"`
}

// Roots contains root-level capabilities
type Roots struct {
	ListChanged bool `json:"listChanged"`
}

// Sampling represents sampling-related capabilities
// Currently empty but structured for future expansion
type Sampling struct{}

// ClientInfo contains information about the connected client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResponse struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ClientInfo contains information about the connected client
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Capabilities represents the available feature capabilities
type ServerCapabilities struct {
	Logging *Logging `json:"logging,omitempty"`
	Tools   *Tools   `json:"tools,omitempty"`
}

type Logging struct{}

type Tools struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ListToolsRequest struct {
	Cursor string `json:"cursor,omitempty"`
}

type ListToolsResponse struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type CallToolRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type CallToolResponse struct {
	IsError bool      `json:"isError"`
	Content []Content `json:"content"`
}

type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

type Prompt struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Arguments   []Argument `json:"arguments"`
}

type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type ListPromptsRequest struct {
	Cursor string `json:"cursor,omitempty"`
}

type ListPromptsResponse struct {
	Prompts    []Prompt `json:"prompts"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

type GetPromptRequest struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments"`
}

type GetPromptResponse struct {
	Description string    `json:"description"`
	Messages    []Message `json:"messages"`
}

type Message struct {
	Role    string  `json:"role"`
	Content Content `json:"content"`
}

type ListResourcesRequest struct {
	Cursor string `json:"cursor,omitempty"`
}

type ListResourcesResponse struct {
	Resources  []Resource `json:"resources"`
	NextCursor string     `json:"nextCursor,omitempty"`
}

type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

type ReadResourceRequest struct {
	URI string `json:"uri"`
}

type ReadResourceResponse struct {
	Contents []ResourceContent `json:"contents"`
}

type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

type ListResourceTemplatesRequest struct {
	Cursor string `json:"cursor,omitempty"`
}

type ListResourceTemplatesResponse struct {
	Templates  []ResourceTemplate `json:"resourceTemplates"`
	NextCursor string             `json:"nextCursor,omitempty"`
}

type ResourceTemplate struct {
	URITemplate string `json:"uriTemplate"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

type CompletionRequest struct {
	Ref      CompletionRef      `json:"ref"`
	Argument CompletionArgument `json:"argument"`
}

type CompletionRef struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type CompletionArgument struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CompletionResponse struct {
	Completion CompletionResult `json:"completion"`
}

type CompletionResult struct {
	Values  []string `json:"values"`
	HasMore bool     `json:"hasMore"`
	Total   int      `json:"total"`
}
