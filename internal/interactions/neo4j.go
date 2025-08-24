package interactions

import (
	"context"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog/log"
)

type QueryResult struct {
	Query  string           `json:"query"`
	Result []map[string]any `json:"result"`
}

func newDriver(ctx context.Context, uri, user, password string) (neo4j.DriverWithContext, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create neo4j driver: %w", err)
	}
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify neo4j connectivity: %w", err)
	}
	log.Debug().Msg("neo4j driver created and connected successfully")
	return driver, nil
}

func executeQuery(ctx context.Context, session neo4j.SessionWithContext, query string) ([]map[string]any, error) {
	result, err := session.Run(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to run query '%s': %w", query, err)
	}
	records, err := result.Collect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect results for query '%s': %w", query, err)
	}
	var results []map[string]any
	for _, record := range records {
		results = append(results, record.AsMap())
	}
	return results, nil
}

func ExecuteNeo4jQueries(ctx context.Context, uri, user, password, database string, queries []string, writeMode bool) ([]QueryResult, error) {
	driver, err := newDriver(ctx, uri, user, password)
	if err != nil {
		return nil, err
	}
	defer driver.Close(ctx)
	mode := neo4j.AccessModeRead
	if writeMode {
		mode = neo4j.AccessModeWrite
	}
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: database,
		AccessMode:   mode,
	})
	defer session.Close(ctx)
	var allResults []QueryResult
	for _, query := range queries {
		log.Debug().Msgf("executing query: %s", query)
		records, err := executeQuery(ctx, session, query)
		if err != nil {
			log.Error().Err(err).Msgf("error executing query, but continuing...")
			allResults = append(allResults, QueryResult{
				Query:  query,
				Result: []map[string]any{{"error": err.Error()}},
			})
			continue
		}
		allResults = append(allResults, QueryResult{
			Query:  query,
			Result: records,
		})
	}
	return allResults, nil
}

func ExecuteNeo4jQueriesFromFile(ctx context.Context, uri, user, password, database, filePath string, writeMode bool) ([]QueryResult, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read query file '%s': %w", filePath, err)
	}
	var queries []string
	if err := yaml.Unmarshal(fileContent, &queries); err != nil {
		return nil, fmt.Errorf("failed to parse YAML query file: %w", err)
	}
	if len(queries) == 0 {
		return nil, fmt.Errorf("no queries found in the file: %s", filePath)
	}
	return ExecuteNeo4jQueries(ctx, uri, user, password, database, queries, writeMode)
}
