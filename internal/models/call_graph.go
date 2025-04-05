package models

// CallGraphNode represents a node in the call graph (a function/method)
type CallGraphNode struct {
	ID         string `json:"id"`                  // Unique identifier (usually package.function or package.receiver.method)
	Package    string `json:"package"`             // Package name
	Function   string `json:"function"`            // Function/method name
	Receiver   string `json:"receiver,omitempty"`  // For methods, the receiver type
	FilePath   string `json:"file_path,omitempty"` // File path containing the function
	Line       int    `json:"line,omitempty"`      // Line number where function starts
	IsExternal bool   `json:"is_external"`         // Whether it's an external function (stdlib or third-party)
}

// CallGraphEdge represents an edge in the call graph (a function call)
type CallGraphEdge struct {
	Source     string `json:"source"`               // Caller function ID
	Target     string `json:"target"`               // Callee function ID
	Line       int    `json:"line,omitempty"`       // Line number where the call occurs
	Parameters string `json:"parameters,omitempty"` // JSON string of parameters
	Count      int    `json:"count"`                // Number of times this call occurs
}

// CallGraph represents a complete call graph for a repository or file
type CallGraph struct {
	Nodes []CallGraphNode `json:"nodes"`
	Edges []CallGraphEdge `json:"edges"`
}

// GetIndexWithCallGraphResponse extends the GetIndexResponse with call graph information
type GetIndexWithCallGraphResponse struct {
	Repository *Repository          `json:"repository"`
	Files      []RepositoryFile     `json:"files,omitempty"`
	Functions  []RepositoryFunction `json:"functions,omitempty"`
	Symbols    []RepositorySymbol   `json:"symbols,omitempty"`
	CallGraph  *CallGraph           `json:"call_graph,omitempty"`
}

// BuildCallGraph builds a call graph from function calls
func BuildCallGraph(functions []RepositoryFunction, calls []FunctionCall, fileMap map[int64]RepositoryFile) *CallGraph {
	// Create a map of function IDs to nodes
	nodeMap := make(map[int64]*CallGraphNode)
	nodes := []CallGraphNode{}

	// Create a map to track external calls
	externalNodes := make(map[string]bool)

	// Create a map to track the number of times each call is made
	callCounts := make(map[string]map[string]int)

	// Create a map to track calls by their source/target
	callDetails := make(map[string]map[string]*CallGraphEdge)

	// Process all functions to create nodes
	for _, fn := range functions {
		id := ""

		// Create node ID based on package and function name
		file, ok := fileMap[fn.FileID]
		if !ok {
			continue
		}

		if fn.Receiver != "" {
			id = file.Package + "." + fn.Receiver + "." + fn.Name
		} else {
			id = file.Package + "." + fn.Name
		}

		node := CallGraphNode{
			ID:         id,
			Package:    file.Package,
			Function:   fn.Name,
			Receiver:   fn.Receiver,
			FilePath:   file.FilePath,
			Line:       fn.Line,
			IsExternal: false,
		}

		nodes = append(nodes, node)
		nodeMap[fn.ID] = &node

		// Initialize call count maps for this function
		if _, ok := callCounts[id]; !ok {
			callCounts[id] = make(map[string]int)
		}

		if _, ok := callDetails[id]; !ok {
			callDetails[id] = make(map[string]*CallGraphEdge)
		}
	}

	// Process all calls to create edges
	for _, call := range calls {
		var sourceID string
		sourceNode, ok := nodeMap[call.CallerID]
		if !ok {
			// Skip calls without a source
			continue
		}
		sourceID = sourceNode.ID

		// Create or get target node
		targetID := call.CalleePackage + "." + call.CalleeName

		// For external calls that we don't have function info for
		if call.CalleeID == nil {
			if _, ok := externalNodes[targetID]; !ok {
				externalNodes[targetID] = true

				// Create an external node
				split := SplitFunctionName(call.CalleeName)

				nodes = append(nodes, CallGraphNode{
					ID:         targetID,
					Package:    call.CalleePackage,
					Function:   split.FunctionName,
					Receiver:   split.ReceiverName,
					IsExternal: true,
				})
			}
		}

		// Update call count
		callCounts[sourceID][targetID]++

		// Create or update call details
		if _, ok := callDetails[sourceID][targetID]; !ok {
			callDetails[sourceID][targetID] = &CallGraphEdge{
				Source:     sourceID,
				Target:     targetID,
				Line:       call.Line,
				Parameters: call.Parameters,
				Count:      1,
			}
		} else {
			callDetails[sourceID][targetID].Count = callCounts[sourceID][targetID]
		}
	}

	// Build edges from call details
	edges := []CallGraphEdge{}
	for _, targetMap := range callDetails {
		for _, edge := range targetMap {
			edges = append(edges, *edge)
		}
	}

	return &CallGraph{
		Nodes: nodes,
		Edges: edges,
	}
}

// FunctionNameParts contains the split parts of a function name
type FunctionNameParts struct {
	ReceiverName string
	FunctionName string
}

// SplitFunctionName attempts to split a function name into receiver and function parts
func SplitFunctionName(name string) FunctionNameParts {
	// Handle simple case with no dot
	if len(name) == 0 || name[0] == '(' {
		return FunctionNameParts{FunctionName: name}
	}

	// Check for method format: Type.Method or (*Type).Method
	dotIndex := -1
	for i := 0; i < len(name); i++ {
		if name[i] == '.' {
			dotIndex = i
			break
		}
	}

	if dotIndex > 0 {
		receiver := name[:dotIndex]
		function := name[dotIndex+1:]
		return FunctionNameParts{
			ReceiverName: receiver,
			FunctionName: function,
		}
	}

	return FunctionNameParts{FunctionName: name}
}
