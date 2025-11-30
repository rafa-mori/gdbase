// Package dockerstack provides migration management for PostgreSQL
package dockerstack

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/kubex-ecosystem/gdbase/internal/bootstrap"
	"github.com/kubex-ecosystem/logz"
	gl "github.com/kubex-ecosystem/logz"

	_ "github.com/lib/pq"
)

// MigrationResult represents the result of a migration execution
type MigrationResult struct {
	FileName        string
	TotalStatements int
	SuccessfulStmts int
	FailedStmts     int
	Errors          []StatementError
	Duration        time.Duration
}

// StatementError represents an error in a specific SQL statement
type StatementError struct {
	Statement string
	Error     string
	Line      int
}

// MigrationManager handles database initialization and migrations with error recovery
type MigrationManager struct {
	dsn string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(dsn string, logger *logz.LoggerZ) *MigrationManager {
	return &MigrationManager{
		dsn: dsn,
	}
}

// WaitForPostgres waits for PostgreSQL to be ready with exponential backoff
func (m *MigrationManager) WaitForPostgres(ctx context.Context, maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)
	attempt := 1

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		db, err := sql.Open("postgres", m.dsn)
		if err != nil {
			gl.Log("debug", fmt.Sprintf("Attempt %d: Connection failed: %v", attempt, err))
			time.Sleep(time.Duration(attempt) * time.Second)
			attempt++
			continue
		}

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		if err := db.PingContext(ctx); err != nil {
			cancel()
			db.Close()
			gl.Log("debug", fmt.Sprintf("Attempt %d: Ping failed: %v", attempt, err))
			time.Sleep(time.Duration(attempt) * time.Second)
			attempt++
			continue
		}
		cancel()
		db.Close()

		gl.Log("info", fmt.Sprintf("✅ PostgreSQL is ready (attempt %d)", attempt))
		return nil
	}

	return fmt.Errorf("PostgreSQL not ready after %v (tried %d times)", maxWait, attempt-1)
}

// RunMigrations executes all SQL files in order with error recovery
func (m *MigrationManager) RunMigrations(ctx context.Context) ([]MigrationResult, error) {
	db, err := sql.Open("postgres", m.dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Execute migrations in order
	migrations := []string{"001_init.sql", "002_hardening.sql"}
	results := make([]MigrationResult, 0, len(migrations))

	gl.Log("info", "🚀 Starting PostgreSQL migrations with error recovery...")

	for _, filename := range migrations {
		result := m.executeSQLFileWithRecovery(ctx, db, filename)
		results = append(results, result)

		// Log summary for this file
		if result.FailedStmts == 0 {
			gl.Log("info", fmt.Sprintf("✅ %s: %d/%d statements executed successfully (%.2fs)",
				filename, result.SuccessfulStmts, result.TotalStatements, result.Duration.Seconds()))
		} else {
			gl.Log("warn", fmt.Sprintf("⚠️  %s: %d/%d statements succeeded, %d failed (%.2fs)",
				filename, result.SuccessfulStmts, result.TotalStatements, result.FailedStmts, result.Duration.Seconds())) // Log first few errors for debugging
			for i, err := range result.Errors {
				if i >= 3 { // Limit error logging
					gl.Log("warn", fmt.Sprintf("... and %d more errors", len(result.Errors)-i))
					break
				}
				gl.Log("error", fmt.Sprintf("   Line %d: %s", err.Line, err.Error))
			}
		}
	}

	// Overall summary
	totalSuccess := 0
	totalFailed := 0
	for _, r := range results {
		totalSuccess += r.SuccessfulStmts
		totalFailed += r.FailedStmts
	}

	if totalFailed == 0 {
		gl.Log("info", fmt.Sprintf("🎉 All migrations completed successfully! (%d statements)", totalSuccess))
	} else {
		gl.Log("warn", fmt.Sprintf("⚠️  Migrations completed with partial success: %d succeeded, %d failed", totalSuccess, totalFailed))
	}

	return results, nil
}

// executeSQLFileWithRecovery executes a single SQL file with statement-level error recovery
func (m *MigrationManager) executeSQLFileWithRecovery(ctx context.Context, db *sql.DB, filename string) MigrationResult {
	start := time.Now()
	result := MigrationResult{
		FileName: filename,
		Errors:   make([]StatementError, 0),
	}

	// Read file content (embed.FS is already rooted at assets/)
	content, err := bootstrap.MigrationFiles.ReadFile(filepath.Join("embedded", filename))
	if err != nil {
		result.Errors = append(result.Errors, StatementError{
			Statement: "",
			Error:     fmt.Sprintf("Failed to read file: %v", err),
			Line:      0,
		})
		result.Duration = time.Since(start)
		return result
	}

	// Parse SQL statements
	statements := m.parseSQL(string(content))
	result.TotalStatements = len(statements)

	gl.Log("debug", fmt.Sprintf("📝 Executing %s (%d statements)...", filename, len(statements)))

	// Execute each statement individually
	for i, stmt := range statements {
		if strings.TrimSpace(stmt.SQL) == "" {
			continue
		}

		// Execute statement with timeout
		stmtCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, err := db.ExecContext(stmtCtx, stmt.SQL)
		cancel()

		if err != nil {
			result.FailedStmts++
			result.Errors = append(result.Errors, StatementError{
				Statement: stmt.SQL,
				Error:     err.Error(),
				Line:      stmt.Line,
			})

			// Log individual statement error (debug level to avoid spam)
			gl.Log("debug", fmt.Sprintf("❌ Statement %d failed: %v", i+1, err))
		} else {
			result.SuccessfulStmts++
			gl.Log("debug", fmt.Sprintf("✅ Statement %d executed", i+1))
		}
	}

	result.Duration = time.Since(start)
	return result
}

// SQLStatement represents a parsed SQL statement with line information
type SQLStatement struct {
	SQL  string
	Line int
}

// parseSQL splits SQL content into individual statements, preserving line numbers
func (m *MigrationManager) parseSQL(content string) []SQLStatement {
	var stmts []SQLStatement
	var b strings.Builder
	line := 1
	stmtStartLine := 1

	// state
	var inLineComment bool
	var inBlockComment bool
	var dollarTag string

	runes := []rune(content)
	i := 0
	for i < len(runes) {
		r := runes[i]

		// track line numbers
		if r == '\n' {
			line++
		}

		// handle end of line comment
		if inLineComment {
			b.WriteRune(r)
			if r == '\n' {
				inLineComment = false
			}
			i++
			continue
		}

		// handle end of block comment
		if inBlockComment {
			b.WriteRune(r)
			if r == '*' && i+1 < len(runes) && runes[i+1] == '/' {
				b.WriteRune(runes[i+1])
				i += 2
				inBlockComment = false
				continue
			}
			i++
			continue
		}

		// handle dollar-quote content
		if dollarTag != "" {
			b.WriteRune(r)
			tagRunes := []rune(dollarTag)
			if r == tagRunes[0] && i+len(tagRunes) <= len(runes) {
				match := true
				for k := 0; k < len(tagRunes); k++ {
					if runes[i+k] != tagRunes[k] {
						match = false
						break
					}
				}
				if match {
					for k := 1; k < len(tagRunes); k++ {
						b.WriteRune(runes[i+k])
					}
					i += len(tagRunes)
					dollarTag = ""
					continue
				}
			}
			i++
			continue
		}

		// detect start of line comment --
		if r == '-' && i+1 < len(runes) && runes[i+1] == '-' {
			inLineComment = true
			b.WriteRune(r)
			b.WriteRune(runes[i+1])
			i += 2
			continue
		}

		// detect start of block comment /*
		if r == '/' && i+1 < len(runes) && runes[i+1] == '*' {
			inBlockComment = true
			b.WriteRune(r)
			b.WriteRune(runes[i+1])
			i += 2
			continue
		}

		// detect dollar quote start $tag$
		if r == '$' {
			j := i + 1
			for j < len(runes) && runes[j] != '$' && ((runes[j] >= 'a' && runes[j] <= 'z') || (runes[j] >= 'A' && runes[j] <= 'Z') || (runes[j] >= '0' && runes[j] <= '9') || runes[j] == '_') {
				j++
			}
			if j < len(runes) && runes[j] == '$' {
				tagRunes := runes[i : j+1]
				dollarTag = string(tagRunes)
				for k := 0; k < len(tagRunes); k++ {
					b.WriteRune(tagRunes[k])
				}
				i = j + 1
				continue
			}
		}

		// single-quote start
		if r == '\'' {
			b.WriteRune(r)
			i++
			for i < len(runes) {
				b.WriteRune(runes[i])
				if runes[i] == '\'' {
					if i+1 < len(runes) && runes[i+1] == '\'' {
						b.WriteRune(runes[i+1])
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
			continue
		}

		// double-quote start (identifiers)
		if r == '"' {
			b.WriteRune(r)
			i++
			for i < len(runes) {
				b.WriteRune(runes[i])
				if runes[i] == '"' {
					if i+1 < len(runes) && runes[i+1] == '"' {
						b.WriteRune(runes[i+1])
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
			continue
		}

		// semicolon at top level -> end statement
		if r == ';' {
			b.WriteRune(r)
			stmt := strings.TrimSpace(b.String())
			if stmt != "" && stmt != ";" {
				stmts = append(stmts, SQLStatement{
					SQL:  stmt,
					Line: stmtStartLine,
				})
			}
			b.Reset()
			i++
			for i < len(runes) && runes[i] == '\n' {
				i++
				line++
			}
			stmtStartLine = line
			continue
		}

		if b.Len() == 0 {
			stmtStartLine = line
		}
		b.WriteRune(r)
		i++
	}

	if strings.TrimSpace(b.String()) != "" {
		stmts = append(stmts, SQLStatement{
			SQL:  strings.TrimSpace(b.String()),
			Line: stmtStartLine,
		})
	}

	return stmts
}

// SchemaExists checks if the required schema is already initialized
func (m *MigrationManager) SchemaExists() (bool, error) {
	db, err := sql.Open("postgres", m.dsn)
	if err != nil {
		return false, err
	}
	defer db.Close()

	// Check for tables in public schema (more reliable than just counting)
	var count int
	query := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
	`

	err = db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, err
	}

	// Also check for extensions (our init script creates extensions)
	var extCount int
	extQuery := `
		SELECT COUNT(*)
		FROM pg_extension
		WHERE extname IN ('uuid-ossp', 'pgcrypto', 'pg_trgm')
	`

	err = db.QueryRow(extQuery).Scan(&extCount)
	if err != nil {
		gl.Log("debug", fmt.Sprintf("Could not check extensions: %v", err))
		extCount = 0
	}

	// Consider schema exists if we have tables OR our extensions
	exists := count > 0 || extCount >= 2

	if exists {
		gl.Log("debug", fmt.Sprintf("Schema check: %d tables, %d extensions found", count, extCount))
	}

	return exists, nil
}

// ValidateConnection tests the database connection
func (m *MigrationManager) ValidateConnection() error {
	db, err := sql.Open("postgres", m.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}
