package repointel

import (
	"cred.com/hack25/backend/internal/insights"
)

// Alias types from the insights package to maintain backward compatibility
type InsightRequest = insights.InsightRequest
type InsightType = insights.InsightType
type FunctionInsight = insights.FunctionInsight
type SymbolInsight = insights.SymbolInsight
type StructInsight = insights.StructInsight
type FileInsight = insights.FileInsight
type RepositoryInsight = insights.RepositoryInsight
type InsightRecord = insights.InsightRecord
type LLMRequest = insights.LLMRequest
type LLMMessage = insights.LLMMessage
type LLMResponse = insights.LLMResponse

// Constants for insight types
const (
	InsightTypeFunction   = insights.FunctionInsightType
	InsightTypeSymbol     = insights.SymbolInsightType
	InsightTypeStruct     = insights.StructInsightType
	InsightTypeFile       = insights.FileInsightType
	InsightTypeRepository = insights.RepositoryInsightType
)
