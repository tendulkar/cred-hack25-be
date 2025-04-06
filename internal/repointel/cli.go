package repointel

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"strconv"
// 	"strings"

// 	"cred.com/hack25/backend/internal/repository"
// 	"github.com/spf13/cobra"
// )

// // RegisterCommands registers CLI commands for repository intelligence
// func RegisterCommands(rootCmd *cobra.Command, codeAnalyzerRepo *repository.CodeAnalyzerRepository) {
// 	// Create services and repositories
// 	service := NewService(
// 		codeAnalyzerRepo,
// 		os.Getenv("LITELLM_URL"),
// 		os.Getenv("LITELLM_API_KEY"),
// 		getDefaultModel(),
// 	)

// 	// Create both old and new repositories to support the transition
// 	insightRepo := NewRepository(codeAnalyzerRepo.DB.DB)
// 	insightsRepo := NewInsightsRepository(codeAnalyzerRepo.DB.DB)

// 	// Create the repointel command
// 	repointelCmd := &cobra.Command{
// 		Use:   "repointel",
// 		Short: "Generate and manage repository intelligence",
// 		Long:  `Generate and manage intelligent insights for developers from repository code analysis`,
// 	}

// 	// Command to generate function insights
// 	generateFunctionCmd := &cobra.Command{
// 		Use:   "function [repoID] [functionID/name]",
// 		Short: "Generate insights for a function",
// 		Long:  `Generate intelligent insights for a specific function in the repository`,
// 		Args:  cobra.ExactArgs(2),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			repoID, err := strconv.ParseInt(args[0], 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("invalid repository ID: %w", err)
// 			}

// 			// Check if the second argument is a function ID or name
// 			functionID, err := strconv.ParseInt(args[1], 10, 64)
// 			if err != nil {
// 				// If it's not a numeric ID, we need to find the function by name
// 				// Build a query to get the function ID directly
// 				query := `
// 					SELECT id FROM code_analyzer.repository_functions
// 					WHERE repository_id = $1 AND name = $2
// 					LIMIT 1
// 				`
// 				var id int64
// 				err = codeAnalyzerRepo.DB.Get(&id, query, repoID, args[1])
// 				if err != nil {
// 					return fmt.Errorf("error finding function '%s': %w", args[1], err)
// 				}
// 				functionID = id
// 			}

// 			modelName, _ := cmd.Flags().GetString("model")

// 			insight, err := service.GenerateFunctionInsight(repoID, functionID, modelName)
// 			if err != nil {
// 				return fmt.Errorf("failed to generate function insight: %w", err)
// 			}

// 			// Save the insight using the new specialized repository
// 			_, err = insightsRepo.SaveFunctionInsight(repoID, functionID, insight, modelName)
// 			if err != nil {
// 				return fmt.Errorf("failed to save function insight: %w", err)
// 			}

// 			// Also save to the old repository for backward compatibility during transition
// 			insightJSON, err := json.Marshal(insight)
// 			if err != nil {
// 				return fmt.Errorf("failed to marshal function insight: %w", err)
// 			}

// 			insightRecord := &InsightRecord{
// 				RepositoryID: repoID,
// 				FunctionID:   &functionID,
// 				Type:         InsightTypeFunction,
// 				Data:         string(insightJSON),
// 				Model:        modelName,
// 			}

// 			err = insightRepo.SaveInsight(insightRecord)
// 			if err != nil {
// 				// Just log the error but don't fail if old format save fails
// 				fmt.Printf("Note: Failed to save to legacy insight format: %v\n", err)
// 			}

// 			// Output formatted insight with the new structure
// 			fmt.Printf("Function Insight:\n")
// 			fmt.Printf("Problem: %s\n", insight.Intent.Problem)
// 			fmt.Printf("Goal: %s\n", insight.Intent.Goal)
// 			fmt.Printf("Result: %s\n\n", insight.Intent.Result)

// 			// Display database operations if present
// 			if len(insight.Database) > 0 {
// 				fmt.Printf("Database Operations:\n")
// 				for _, db := range insight.Database {
// 					fmt.Printf("- %s on %s: %s\n", db.Action, db.Engine, db.Purpose)
// 				}
// 				fmt.Println()
// 			}

// 			// Display network calls if present
// 			if len(insight.Network) > 0 {
// 				fmt.Printf("Network Calls:\n")
// 				for _, net := range insight.Network {
// 					fmt.Printf("- %s to %s: %s\n", net.Protocol, net.Endpoint, net.Purpose)
// 				}
// 				fmt.Println()
// 			}

// 			// Display notes if present
// 			if insight.Notes != "" {
// 				fmt.Printf("Notes: %s\n", insight.Notes)
// 			}

// 			// Display object storage operations if present
// 			if len(insight.ObjectStore) > 0 {
// 				fmt.Printf("Storage Operations:\n")
// 				for _, store := range insight.ObjectStore {
// 					fmt.Printf("- %s on %s/%s: %s\n", store.Action, store.Provider, store.Bucket, store.Purpose)
// 				}
// 				fmt.Println()
// 			}

// 			// Display related functions if present
// 			if len(insight.Related) > 0 {
// 				fmt.Printf("Related Functions:\n")
// 				for _, fn := range insight.Related {
// 					fmt.Printf("- %s\n", fn)
// 				}
// 				fmt.Println()
// 			}

// 			// Display framework usage if present
// 			if len(insight.Frameworks) > 0 {
// 				fmt.Printf("Frameworks Used:\n")
// 				for _, fw := range insight.Frameworks {
// 					fmt.Printf("- %s: %s\n", fw.Name, fw.Purpose)
// 				}
// 				fmt.Println()
// 			}

// 			return nil
// 		},
// 	}
// 	generateFunctionCmd.Flags().String("model", "", "LLM model to use for generation")

// 	// Command to generate symbol insights
// 	generateSymbolCmd := &cobra.Command{
// 		Use:   "symbol [repoID] [symbolID/name]",
// 		Short: "Generate insights for a symbol",
// 		Long:  `Generate intelligent insights for a specific symbol in the repository`,
// 		Args:  cobra.ExactArgs(2),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			repoID, err := strconv.ParseInt(args[0], 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("invalid repository ID: %w", err)
// 			}

// 			// Check if the second argument is a symbol ID or name
// 			symbolID, err := strconv.ParseInt(args[1], 10, 64)
// 			if err != nil {
// 				// If it's not a numeric ID, we need to find the symbol by name
// 				// Build a query to get the symbol ID directly
// 				query := `
// 					SELECT id FROM code_analyzer.repository_symbols
// 					WHERE repository_id = $1 AND name = $2
// 					LIMIT 1
// 				`
// 				var id int64
// 				err = codeAnalyzerRepo.DB.Get(&id, query, repoID, args[1])
// 				if err != nil {
// 					return fmt.Errorf("error finding symbol '%s': %w", args[1], err)
// 				}
// 				symbolID = id
// 			}

// 			modelName, _ := cmd.Flags().GetString("model")

// 			isStruct, _ := cmd.Flags().GetBool("struct")
// 			if isStruct {
// 				insight, err := service.GenerateStructInsight(repoID, symbolID, modelName)
// 				if err != nil {
// 					return fmt.Errorf("failed to generate struct insight: %w", err)
// 				}

// 				// Save the insight using the new specialized repository
// 				_, err = insightsRepo.SaveStructInsight(repoID, symbolID, insight, modelName)
// 				if err != nil {
// 					return fmt.Errorf("failed to save struct insight: %w", err)
// 				}

// 				// Also save to the old repository for backward compatibility during transition
// 				insightJSON, err := json.Marshal(insight)
// 				if err != nil {
// 					return fmt.Errorf("failed to marshal struct insight: %w", err)
// 				}

// 				insightRecord := &InsightRecord{
// 					RepositoryID: repoID,
// 					SymbolID:     &symbolID,
// 					Type:         InsightTypeStruct,
// 					Data:         string(insightJSON),
// 					Model:        modelName,
// 				}

// 				err = insightRepo.SaveInsight(insightRecord)
// 				if err != nil {
// 					// Just log the error but don't fail if old format save fails
// 					fmt.Printf("Note: Failed to save to legacy insight format: %v\n", err)
// 				}

// 				// Output formatted insight
// 				fmt.Printf("Struct Insight:\n")
// 				// fmt.Printf("Knowledge Graph: %s\n", insight.KnowledgeGraph)
// 				// fmt.Printf("Usage: %s\n", insight.Usage)
// 				// fmt.Printf("Rationale: %s\n", insight.Rationale)
// 				// fmt.Printf("Data Model: %s\n", insight.DataModel)

// 				// if len(insight.RelatedStructs) > 0 {
// 				// 	fmt.Printf("\nRelated Structs:\n")
// 				// 	for _, s := range insight.RelatedStructs {
// 				// 		fmt.Printf("- %s\n", s)
// 				// 	}
// 				// }
// 			} else {
// 				insight, err := service.GenerateSymbolInsight(repoID, symbolID, modelName)
// 				if err != nil {
// 					return fmt.Errorf("failed to generate symbol insight: %w", err)
// 				}

// 				// Save the insight using the new specialized repository
// 				_, err = insightsRepo.SaveSymbolInsight(repoID, symbolID, insight, modelName)
// 				if err != nil {
// 					return fmt.Errorf("failed to save symbol insight: %w", err)
// 				}

// 				// Also save to the old repository for backward compatibility during transition
// 				insightJSON, err := json.Marshal(insight)
// 				if err != nil {
// 					return fmt.Errorf("failed to marshal symbol insight: %w", err)
// 				}

// 				insightRecord := &InsightRecord{
// 					RepositoryID: repoID,
// 					SymbolID:     &symbolID,
// 					Type:         InsightTypeSymbol,
// 					Data:         string(insightJSON),
// 					Model:        modelName,
// 				}

// 				err = insightRepo.SaveInsight(insightRecord)
// 				if err != nil {
// 					// Just log the error but don't fail if old format save fails
// 					fmt.Printf("Note: Failed to save to legacy insight format: %v\n", err)
// 				}

// 				// Output formatted insight
// 				fmt.Printf("Symbol Insight:\n")
// 				// fmt.Printf("Knowledge Graph: %s\n", insight.KnowledgeGraph)
// 				// fmt.Printf("Usage: %s\n", insight.Usage)
// 				// fmt.Printf("Rationale: %s\n", insight.Rationale)

// 				// if insight.AdditionalInfo != "" {
// 				// 	fmt.Printf("Additional Information: %s\n", insight.AdditionalInfo)
// 				// }
// 			}

// 			return nil
// 		},
// 	}
// 	generateSymbolCmd.Flags().String("model", "", "LLM model to use for generation")
// 	generateSymbolCmd.Flags().Bool("struct", false, "Generate struct-specific insights")

// 	// Command to list available LLMs
// 	listModelsCmd := &cobra.Command{
// 		Use:   "list-models",
// 		Short: "List available LLM models",
// 		Long:  `List all available LLM models that can be used for generating insights`,
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			fmt.Println("Available LLM Models:")
// 			fmt.Println("- claude-instant-1")
// 			fmt.Println("- claude-2")
// 			fmt.Println("- claude-3-haiku-20240307")
// 			fmt.Println("- claude-3-sonnet-20240229")
// 			fmt.Println("- claude-3-opus-20240229")
// 			fmt.Println("- gpt-3.5-turbo")
// 			fmt.Println("- gpt-4-turbo")
// 			fmt.Println("- gpt-4o")
// 			fmt.Println("- llama-3-8b")
// 			fmt.Println("- llama-3-70b")
// 			fmt.Println("- gemini-pro")
// 			fmt.Println("- mistral-large")
// 			return nil
// 		},
// 	}

// 	// Command to analyze a file
// 	analyzeFileCmd := &cobra.Command{
// 		Use:   "file [repoID] [fileID/path]",
// 		Short: "Generate insights for a file",
// 		Long:  `Generate intelligent insights for a specific file in the repository`,
// 		Args:  cobra.ExactArgs(2),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			repoID, err := strconv.ParseInt(args[0], 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("invalid repository ID: %w", err)
// 			}

// 			// Check if the second argument is a file ID or path
// 			fileID, err := strconv.ParseInt(args[1], 10, 64)
// 			if err != nil {
// 				// Try to find the file by path
// 				path := args[1]
// 				file, err := codeAnalyzerRepo.GetRepositoryFileByPath(repoID, path)
// 				if err != nil {
// 					return fmt.Errorf("error finding file '%s': %w", path, err)
// 				}
// 				if file == nil {
// 					return fmt.Errorf("file '%s' not found in repository", path)
// 				}
// 				fileID = file.ID
// 			}

// 			modelName, _ := cmd.Flags().GetString("model")

// 			// Generate the insight
// 			fmt.Printf("Analyzing file with ID %d\n", fileID)
// 			insight, err := service.GenerateFileInsight(repoID, fileID, modelName)
// 			if err != nil {
// 				return fmt.Errorf("failed to generate file insight: %w", err)
// 			}

// 			// Save the insight
// 			insightJSON, err := json.Marshal(insight)
// 			if err != nil {
// 				return fmt.Errorf("failed to marshal file insight: %w", err)
// 			}

// 			insightRecord := &InsightRecord{
// 				RepositoryID: repoID,
// 				FileID:       &fileID,
// 				Type:         InsightTypeFile,
// 				Data:         string(insightJSON),
// 				Model:        modelName,
// 			}

// 			err = insightRepo.SaveInsight(insightRecord)
// 			if err != nil {
// 				return fmt.Errorf("failed to save file insight: %w", err)
// 			}

// 			// Output formatted insight
// 			fmt.Printf("\nFile Insight:\n")
// 			// fmt.Printf("Purpose: %s\n\n", insight.Purpose)

// 			// fmt.Printf("Main Components and Responsibilities:\n%s\n\n", insight.Responsibilities)
// 			// if len(insight.Components) > 0 {
// 			// 	fmt.Printf("Components:\n")
// 			// 	for _, comp := range insight.Components {
// 			// 		fmt.Printf("- %s\n", comp)
// 			// 	}
// 			// 	fmt.Printf("\n")
// 			// }

// 			// if len(insight.Dependencies) > 0 {
// 			// 	fmt.Printf("Dependencies:\n")
// 			// 	for _, dep := range insight.Dependencies {
// 			// 		fmt.Printf("- %s\n", dep)
// 			// 	}
// 			// 	fmt.Printf("\n")
// 			// }

// 			// fmt.Printf("Data Flow:\n%s\n", insight.DataFlows)

// 			return nil
// 		},
// 	}
// 	analyzeFileCmd.Flags().String("model", "", "LLM model to use for generation")

// 	// Command to analyze a repository at a high level (architecture, key components, etc.)
// 	analyzeRepoArchCmd := &cobra.Command{
// 		Use:   "architecture [repoID]",
// 		Short: "Analyze repository architecture",
// 		Long:  `Generate high-level insights about the repository architecture, key components, and data flows`,
// 		Args:  cobra.ExactArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			repoID, err := strconv.ParseInt(args[0], 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("invalid repository ID: %w", err)
// 			}

// 			modelName, _ := cmd.Flags().GetString("model")

// 			// Generate the insight
// 			fmt.Printf("Analyzing repository architecture for ID %d\n", repoID)
// 			insight, err := service.GenerateRepositoryInsight(repoID, modelName)
// 			if err != nil {
// 				return fmt.Errorf("failed to generate repository insight: %w", err)
// 			}

// 			// Save the insight
// 			insightJSON, err := json.Marshal(insight)
// 			if err != nil {
// 				return fmt.Errorf("failed to marshal repository insight: %w", err)
// 			}

// 			insightRecord := &InsightRecord{
// 				RepositoryID: repoID,
// 				Type:         InsightTypeRepository,
// 				Data:         string(insightJSON),
// 				Model:        modelName,
// 			}

// 			err = insightRepo.SaveInsight(insightRecord)
// 			if err != nil {
// 				return fmt.Errorf("failed to save repository insight: %w", err)
// 			}

// 			// Output formatted insight
// 			fmt.Printf("\nRepository Architecture Insight:\n")
// 			// fmt.Printf("Purpose: %s\n\n", insight.Purpose)
// 			// fmt.Printf("Architecture Overview:\n%s\n\n", insight.Architecture)

// 			// if len(insight.KeyComponents) > 0 {
// 			// 	fmt.Printf("Key Components:\n")
// 			// 	for _, comp := range insight.KeyComponents {
// 			// 		fmt.Printf("- %s\n", comp)
// 			// 	}
// 			// 	fmt.Printf("\n")
// 			// }

// 			// fmt.Printf("Data Flows:\n%s\n\n", insight.DataFlows)

// 			// if len(insight.Dependencies) > 0 {
// 			// 	fmt.Printf("External Dependencies:\n")
// 			// 	for _, dep := range insight.Dependencies {
// 			// 		fmt.Printf("- %s\n", dep)
// 			// 	}
// 			// 	fmt.Printf("\n")
// 			// }

// 			// fmt.Printf("Recommendations:\n%s\n", insight.Recommendations)

// 			return nil
// 		},
// 	}
// 	analyzeRepoArchCmd.Flags().String("model", "", "LLM model to use for generation")

// 	// Command to analyze a whole repository (detailed component analysis)
// 	analyzeRepoCmd := &cobra.Command{
// 		Use:   "analyze [repoID]",
// 		Short: "Analyze a full repository",
// 		Long:  `Generate insights for all key components in a repository`,
// 		Args:  cobra.ExactArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			repoID, err := strconv.ParseInt(args[0], 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("invalid repository ID: %w", err)
// 			}

// 			modelName, _ := cmd.Flags().GetString("model")

// 			// Get repository info
// 			repo, err := codeAnalyzerRepo.GetRepositoryByID(repoID)
// 			if err != nil {
// 				return fmt.Errorf("failed to get repository: %w", err)
// 			}
// 			if repo == nil {
// 				return fmt.Errorf("repository with ID %d not found", repoID)
// 			}

// 			fmt.Printf("Analyzing repository: %s (%s)\n", repo.Name, repo.URL)

// 			// Get functions and symbols
// 			functions, err := codeAnalyzerRepo.GetRepositoryFunctions(repoID, 0)
// 			if err != nil {
// 				return fmt.Errorf("failed to get functions: %w", err)
// 			}

// 			symbols, err := codeAnalyzerRepo.GetRepositorySymbols(repoID, 0)
// 			if err != nil {
// 				return fmt.Errorf("failed to get symbols: %w", err)
// 			}

// 			// First, analyze the overall repository architecture
// 			fmt.Println("Generating repository architecture overview...")
// 			_, err = service.GenerateRepositoryInsight(repoID, modelName)
// 			if err != nil {
// 				fmt.Printf("  Warning: Failed to generate repository architecture: %v\n", err)
// 			} else {
// 				fmt.Println("  Repository architecture analysis complete")
// 			}

// 			// Filter to structs only
// 			var structs []int64
// 			for _, sym := range symbols {
// 				if strings.Contains(strings.ToLower(sym.Kind), "struct") {
// 					structs = append(structs, sym.ID)
// 				}
// 			}

// 			// Limit analysis to most important components
// 			maxComponents := 10
// 			fmt.Printf("Found %d functions and %d structs\n", len(functions), len(structs))

// 			functionsToAnalyze := min(maxComponents, len(functions))
// 			structsToAnalyze := min(maxComponents/2, len(structs))

// 			fmt.Printf("Will analyze %d functions and %d structs\n", functionsToAnalyze, structsToAnalyze)

// 			// Analyze functions
// 			for i := 0; i < functionsToAnalyze; i++ {
// 				fn := functions[i]
// 				fmt.Printf("Analyzing function %d/%d: %s\n", i+1, functionsToAnalyze, fn.Name)

// 				_, err := service.GenerateFunctionInsight(repoID, fn.ID, modelName)
// 				if err != nil {
// 					fmt.Printf("  Error: %v\n", err)
// 					continue
// 				}

// 				fmt.Printf("  Generated insights for %s\n", fn.Name)
// 			}

// 			// Analyze structs
// 			for i := 0; i < structsToAnalyze && i < len(structs); i++ {
// 				symID := structs[i]
// 				var symName string
// 				for _, sym := range symbols {
// 					if sym.ID == symID {
// 						symName = sym.Name
// 						break
// 					}
// 				}

// 				fmt.Printf("Analyzing struct %d/%d: %s\n", i+1, structsToAnalyze, symName)

// 				_, err := service.GenerateStructInsight(repoID, symID, modelName)
// 				if err != nil {
// 					fmt.Printf("  Error: %v\n", err)
// 					continue
// 				}

// 				fmt.Printf("  Generated insights for %s\n", symName)
// 			}

// 			fmt.Println("Repository analysis complete. Insights have been generated and saved.")
// 			return nil
// 		},
// 	}
// 	analyzeRepoCmd.Flags().String("model", "", "LLM model to use for generation")

// 	// Add commands to parent command
// 	repointelCmd.AddCommand(generateFunctionCmd)
// 	repointelCmd.AddCommand(generateSymbolCmd)
// 	repointelCmd.AddCommand(listModelsCmd)
// 	repointelCmd.AddCommand(analyzeRepoCmd)
// 	repointelCmd.AddCommand(analyzeFileCmd)
// 	repointelCmd.AddCommand(analyzeRepoArchCmd)

// 	rootCmd.AddCommand(repointelCmd)
// }

// // getDefaultModel returns the default LLM model to use
// func getDefaultModel() string {
// 	model := os.Getenv("LITELLM_DEFAULT_MODEL")
// 	if model == "" {
// 		return "claude-3-sonnet-20240229"
// 	}
// 	return model
// }

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }
