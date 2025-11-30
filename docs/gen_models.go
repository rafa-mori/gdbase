package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	tp "github.com/kubex-ecosystem/gdbase/internal/types"
	gl "github.com/kubex-ecosystem/logz"
	l "github.com/kubex-ecosystem/logz"

	_ "gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Column struct {
	Name string
	Type string
}

// Main GenerateModels generates user models from database
func main() {
	// Initialize database
	_, dbSQL, err := initDB()
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error initializing database: %v", err))
		return
	}
	defer dbSQL.Close()

	// Query to get table structure
	rows, err := dbSQL.Query(`
        SELECT table_name, column_name, data_type
        FROM information_schema.columns
        WHERE table_schema = 'public';
    `)

	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error executing query: %v", err))
		return
	}

	defer rows.Close()

	tables := make(map[string][]Column)

	for rows.Next() {
		var tableName, columnName, dataType string
		if err := rows.Scan(&tableName, &columnName, &dataType); err != nil {
			gl.Log("fatal", fmt.Sprintf("Error scanning row: %v", err))
			return
		}
		tables[tableName] = append(tables[tableName], Column{Name: titleCase(columnName), Type: mapSQLType(dataType)})
	}

	// Generate Go code from structure
	generateGoModels(tables)
}

// Function to title case field names
func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

func initDB() (*gorm.DB, *sql.DB, error) {
	configPath := os.Getenv("GDBASE_CONFIGFILE")
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			gl.Log("fatal", fmt.Sprintf("Error getting user home directory: %v", err))
			return nil, nil, err
		}
		configPath = fmt.Sprintf("%s/.kubex/gdbase/config/config.json", homeDir)
	}
	var dbConfig *svc.DBConfig
	// Load configuration
	if _, err := os.Stat(configPath); err != nil && !os.IsNotExist(err) {
		gl.Log("fatal", fmt.Sprintf("Config file not found at %s", configPath))
		return nil, nil, fmt.Errorf("config file not found at %s", configPath)
	} else if os.IsNotExist(err) {
		dbConfig = svc.NewDBConfig(&svc.DBConfig{})
	} else {
		dbConfig = svc.NewDBConfigWithFilePath("GoBE-DB", configPath)
	}

	if dbConfig == nil {
		gl.Log("fatal", "Error loading database configuration")
		return nil, nil, fmt.Errorf("error loading database configuration")
	}
	if len(os.Args) > 1 {
		database := &tp.Database{
			Enabled:   true,
			IsDefault: true,
			Type:      os.Args[1],
		}
		if strings.Contains(database.Type, "://") {
			database.Dsn = database.Type
			dsnArray := strings.Split(database.Dsn, "/")
			database.Type = strings.TrimSuffix(dsnArray[0], ":")
			database.Name = dsnArray[len(dsnArray)-1]
		} else {
			// Set your DB config here if needed
			switch database.Type {
			case "mysql", "mariadb":
				database.Type = "mysql"
			case "postgres", "postgresql":
				database.Type = "postgresql"
			case "sqlserver", "mssql":
				database.Type = "sqlserver"
			default:
				database.Type = "sqlite"
			}
		}
		dbConfig.Databases = map[string]*tp.Database{
			"GoBE-DB": database,
		}
	}

	// Initialize database
	// Create database service
	dbService, err := svc.NewDatabaseService(context.Background(), dbConfig, l.GetLogger("gen_models"))
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error creating database service: %v", err))
		return nil, nil, err
	}
	// Initialize database service
	err = dbService.Initialize(context.Background())
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error initializing database service: %v", err))
		return nil, nil, err
	}
	db, err := svc.GetDB(context.Background(), dbService.(*svc.DBServiceImpl))
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error getting database instance: %v", err))
		return nil, nil, err
	}
	// Database connection
	dbSQL, err := db.DB()
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error getting database connection: %v", err))
		return nil, nil, err
	}
	if err := dbSQL.Ping(); err != nil {
		gl.Log("fatal", fmt.Sprintf("Error connecting to database: %v", err))
		return nil, nil, err
	}
	fmt.Println("Database connection established successfully!")

	return db, dbSQL, nil
}

// Generate Go structs dynamically
func generateGoModels(tables map[string][]Column) {
	modelTemplate := `package main

{{range $table, $columns := .}}
type {{$table | title}} struct {
	{{range $columns}}
		{{.Name}} {{.Type}} ` + "`" + `json:"{{.Name}}" yaml:"{{.Name}}" xml:"{{.Name}}"` + "`" + `{{end}}
}
{{end}}
`

	file, err := os.Create("models.go")
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error creating file: %v", err))
		return
	}
	defer file.Close()

	tmpl, err := template.New("models").
		Funcs(template.FuncMap{"title": titleCase}).
		Parse(modelTemplate)
	// Option("missingkey=zero") to avoid missing key error
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error creating template: %v", err))
		return
	}

	writer := io.Writer(file)
	if err = tmpl.Execute(writer, tables); err != nil {
		gl.Log("fatal", fmt.Sprintf("Error executing template: %v", err))
		return
	}
	file.Sync()
	fmt.Println("models.go file generated successfully!")
}

// Maps SQL types to Go types
func mapSQLType(sqlType string) string {
	switch strings.ToLower(sqlType) {
	case "integer":
		return "int"
	case "numeric":
		return "float64"
	case "text", "varchar", "character varying":
		return "string"
	case "timestamp", "timestamp without time zone":
		return "time.Time"
	case "boolean":
		return "bool"
	default:
		return "any"
	}
}
