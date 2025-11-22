package interactionsCmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/interactions"
)

var neo4jCmdFlags struct {
	uri        string
	user       string
	password   string
	database   string
	query      string
	queryFile  string
	outputFile string
	writeMode  bool
}

var Neo4jCmd = &cobra.Command{
	Use:   "neo4j",
	Short: "Interact with a Neo4j database",
	Long: `Run Cypher queries against a Neo4j database from a string or a file.
The query file should be a YAML file containing a list of queries.
Multi-line queries are supported using the '|' character in YAML.

Example (single query):
  anbu interactions neo4j -q "MATCH (n) RETURN n LIMIT 5"

Example (query file):
  anbu interactions neo4j --query-file ./queries.yaml --output-file results.json`,
	Run: func(cmd *cobra.Command, args []string) {
		if neo4jCmdFlags.query != "" && neo4jCmdFlags.queryFile != "" {
			log.Fatal().Msg("please provide either a query or a query file, not both")
		}
		if neo4jCmdFlags.query == "" && neo4jCmdFlags.queryFile == "" {
			log.Fatal().Msg("a query or a query file is required")
		}

		ctx := context.Background()
		var results []interactions.QueryResult
		var err error
		if neo4jCmdFlags.query != "" {
			log.Debug().Msgf("executing single query: %s", neo4jCmdFlags.query)
			results, err = interactions.ExecuteNeo4jQueries(ctx, neo4jCmdFlags.uri, neo4jCmdFlags.user, neo4jCmdFlags.password, neo4jCmdFlags.database, []string{neo4jCmdFlags.query}, neo4jCmdFlags.writeMode)
		} else {
			log.Debug().Msgf("executing queries from file: %s", neo4jCmdFlags.queryFile)
			results, err = interactions.ExecuteNeo4jQueriesFromFile(ctx, neo4jCmdFlags.uri, neo4jCmdFlags.user, neo4jCmdFlags.password, neo4jCmdFlags.database, neo4jCmdFlags.queryFile, neo4jCmdFlags.writeMode)
		}
		if err != nil {
			log.Fatal().Err(err).Msg("failed to execute neo4j queries")
		}

		// Marshal the results to JSON
		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to marshal results to JSON")
		}
		err = os.WriteFile(neo4jCmdFlags.outputFile, jsonData, 0644)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to write results to file: %s", neo4jCmdFlags.outputFile)
		}
		log.Info().Msgf("Successfully executed queries and saved results to %s", neo4jCmdFlags.outputFile)
	},
}

func init() {
	Neo4jCmd.Flags().StringVarP(&neo4jCmdFlags.uri, "uri", "r", "neo4j://localhost:7687", "Neo4j URI")
	Neo4jCmd.Flags().StringVarP(&neo4jCmdFlags.user, "user", "u", "neo4j", "Neo4j user")
	Neo4jCmd.Flags().StringVarP(&neo4jCmdFlags.password, "password", "p", "p4SSw0rd", "Neo4j password")
	Neo4jCmd.Flags().StringVarP(&neo4jCmdFlags.database, "database", "d", "neo4j", "Neo4j database")
	Neo4jCmd.Flags().StringVarP(&neo4jCmdFlags.query, "query", "q", "", "Single Cypher query to execute")
	Neo4jCmd.Flags().StringVar(&neo4jCmdFlags.queryFile, "query-file", "", "Path to a YAML file with a list of Cypher queries")
	Neo4jCmd.Flags().StringVarP(&neo4jCmdFlags.outputFile, "output-file", "o", "neo4j-query-result.json", "Output file for the query results")
	Neo4jCmd.Flags().BoolVar(&neo4jCmdFlags.writeMode, "write", false, "Open connection in write mode")
}
