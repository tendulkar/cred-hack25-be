package structured

// Schema definitions for LLM structured outputs

// FunctionInsightSchema defines the schema for function insights
type FunctionInsightSchema struct {
	Intent struct {
		Problem string `json:"problem"`
		Goal    string `json:"goal"`
		Result  string `json:"result"`
	} `json:"intent"`
	Params []struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Purpose string `json:"purpose"`
	} `json:"params"`
	Returns []struct {
		Type    string `json:"type"`
		Purpose string `json:"purpose"`
	} `json:"returns"`
	Network []struct {
		Protocol string `json:"protocol"`
		Endpoint string `json:"endpoint"`
		Purpose  string `json:"purpose"`
	} `json:"network"`
	Database []struct {
		Engine  string `json:"engine"`
		Action  string `json:"action"`
		Purpose string `json:"purpose"`
	} `json:"database"`
	ObjectStore []struct {
		Provider   string `json:"provider"`
		Bucket     string `json:"bucket"`
		Action     string `json:"action"`
		KeyPattern string `json:"key_pattern"`
		Purpose    string `json:"purpose"`
	} `json:"object_store"`
	Compute []struct {
		Category    string `json:"category"`
		Description string `json:"description"`
	} `json:"compute"`
	Observability []struct {
		Type    string `json:"type"`
		Purpose string `json:"purpose"`
	} `json:"observability"`
	Quality []struct {
		Category    string `json:"category"`
		Description string `json:"description"`
	} `json:"quality"`
	Frameworks []struct {
		Name    string `json:"name"`
		Purpose string `json:"purpose"`
	} `json:"frameworks"`
	Patterns []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"patterns"`
	Related []string `json:"related"`
	Notes   string   `json:"notes"`
}

// SymbolInsightSchema defines the schema for symbol insights
type SymbolInsightSchema struct {
	Concept struct {
		Domain      string `json:"domain"`
		Name        string `json:"name"`
		Description string `json:"description"`
		OntologyURI string `json:"ontology_uri,omitempty"`
	} `json:"concept"`
	Decision struct {
		Problem      string `json:"problem"`
		Rationale    string `json:"rationale"`
		Alternatives string `json:"alternatives"`
	} `json:"decision"`
	UsedBy   []string `json:"used_by"`
	Patterns []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Rationale   string `json:"rationale,omitempty"`
	} `json:"patterns"`
	Quality []struct {
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Metric      string  `json:"metric,omitempty"`
		Value       float64 `json:"value,omitempty"`
		Status      string  `json:"status,omitempty"`
	} `json:"quality"`
}

// StructInsightSchema defines the schema for struct insights
type StructInsightSchema struct {
	Concept struct {
		Domain      string `json:"domain"`
		Name        string `json:"name"`
		Description string `json:"description"`
		OntologyURI string `json:"ontology_uri,omitempty"`
	} `json:"concept"`
	Fields []struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Purpose string `json:"purpose"`
	} `json:"fields"`
	Relations []struct {
		Pattern     string `json:"pattern"`
		Description string `json:"description"`
	} `json:"relations"`
	Persistence struct {
		Engine   string `json:"engine"`
		Table    string `json:"table"`
		Strategy string `json:"strategy"`
	} `json:"persistence"`
	Observability []struct {
		Type    string `json:"type"`
		Purpose string `json:"purpose"`
	} `json:"observability"`
	Quality []struct {
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Metric      string  `json:"metric,omitempty"`
		Value       float64 `json:"value,omitempty"`
		Status      string  `json:"status,omitempty"`
	} `json:"quality"`
	Patterns []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Rationale   string `json:"rationale,omitempty"`
	} `json:"patterns"`
}

// FileInsightSchema defines the schema for file insights
type FileInsightSchema struct {
	Responsibilities struct {
		MainPurpose string `json:"main_purpose"`
		Details     string `json:"details"`
	} `json:"responsibilities"`
	Contains     []string `json:"contains"`
	Dependencies []struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Purpose string `json:"purpose"`
	} `json:"dependencies"`
	Observability []struct {
		Type    string `json:"type"`
		Purpose string `json:"purpose"`
	} `json:"observability"`
	Quality []struct {
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Metric      string  `json:"metric,omitempty"`
		Value       float64 `json:"value,omitempty"`
		Status      string  `json:"status,omitempty"`
	} `json:"quality"`
	Patterns []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Rationale   string `json:"rationale,omitempty"`
	} `json:"patterns"`
}

// RepositoryInsightSchema defines the schema for repository insights
type RepositoryInsightSchema struct {
	Domain struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		OntologyURI string `json:"ontology_uri,omitempty"`
	} `json:"domain"`
	Architecture struct {
		Pattern     string `json:"pattern"`
		Description string `json:"description"`
		Strengths   string `json:"strengths"`
		Weaknesses  string `json:"weaknesses"`
		Reason      string `json:"reason,omitempty"`
	} `json:"architecture"`
	Frameworks []struct {
		Name    string `json:"name"`
		Version string `json:"version,omitempty"`
		Purpose string `json:"purpose"`
	} `json:"frameworks"`
	DesignPatterns []struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Location    string   `json:"location,omitempty"`
		Reason      string   `json:"reason,omitempty"`
		AppliesTo   []string `json:"applies_to,omitempty"`
	} `json:"design_patterns"`
	CodingPatterns []struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Rationale   string `json:"rationale,omitempty"`
		Example     string `json:"example,omitempty"`
	} `json:"coding_patterns"`
	CriticalPaths []string `json:"critical_paths"`
	TechDebt      []struct {
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Severity    string  `json:"severity,omitempty"`
		Metric      string  `json:"metric,omitempty"`
		Value       float64 `json:"value,omitempty"`
		Status      string  `json:"status,omitempty"`
	} `json:"tech_debt"`
}
