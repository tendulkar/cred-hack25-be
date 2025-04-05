package models

// CodeAnalysisRequest represents a request to analyze a GitHub repository
type CodeAnalysisRequest struct {
	RepoURL   string `json:"repo_url" binding:"required"`
	AuthToken string `json:"auth_token,omitempty"`
}

// FileAnalysisResult represents the analysis result for a single file
type FileAnalysisResult struct {
	Path          string             `json:"path"`
	Dependencies  []string           `json:"dependencies"`
	GlobalVars    []VariableInfo     `json:"global_vars"`
	Constants     []VariableInfo     `json:"constants"`
	InitFunction  *FunctionInfo      `json:"init_function,omitempty"`
	Structs       []StructInfo       `json:"structs"`
	Methods       []MethodInfo       `json:"methods"`
	WorkflowSteps []WorkflowStepInfo `json:"workflow_steps"`
}

// VariableInfo represents information about a variable or constant
type VariableInfo struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value,omitempty"`
}

// FunctionInfo represents information about a function
type FunctionInfo struct {
	Name          string         `json:"name"`
	InputParams   []VariableInfo `json:"input_params"`
	OutputParams  []VariableInfo `json:"output_params"`
	Functionality string         `json:"functionality"`
}

// StructInfo represents information about a struct
type StructInfo struct {
	Name   string         `json:"name"`
	Fields []VariableInfo `json:"fields"`
}

// MethodInfo represents information about a method
type MethodInfo struct {
	Name          string         `json:"name"`
	Receiver      string         `json:"receiver"`
	InputParams   []VariableInfo `json:"input_params"`
	OutputParams  []VariableInfo `json:"output_params"`
	Functionality string         `json:"functionality"`
}

// WorkflowStepInfo represents information about a workflow step
type WorkflowStepInfo struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	TypeDetails  string   `json:"type_details"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
	InputVars    []string `json:"input_vars"`
	OutputVars   []string `json:"output_vars"`
	WorkflowName string   `json:"workflow_name"`
}

// RepositoryAnalysisResult represents the analysis result for a repository
type RepositoryAnalysisResult struct {
	RepoURL string               `json:"repo_url"`
	Owner   string               `json:"owner"`
	Name    string               `json:"name"`
	Files   []FileAnalysisResult `json:"files"`
}
