package data

type SearchResult struct {
	Language string
	Repo     string
	Filename string
	Snippet  string
	LineNum  int
}

var Results = []SearchResult{
	{
		Language: "Go",
		Repo:     "example/repo1",
		Filename: "main.go",
		Snippet:  "func main() {\n    fmt.Println(\"Hello, World!\")\n    // This is a multi-line example\n    for i := 0; i < 10; i++ {\n        fmt.Printf(\"%d \", i)\n    }\n}",
		LineNum:  1,
	},
	{
		Language: "Go",
		Repo:     "example/repo2",
		Filename: "utils.go",
		Snippet:  "func complexCalculation(x, y int) int {\n    result := 0\n    for i := 0; i < x; i++ {\n        result += y * i\n    }\n    return result\n}",
		LineNum:  15,
	},
	{
		Language: "Go",
		Repo:     "example/repo3",
		Filename: "api.go",
		Snippet:  "func HandleRequest(w http.ResponseWriter, r *http.Request) {\n    params := r.URL.Query()\n    id := params.Get(\"id\")\n    if id == \"\" {\n        http.Error(w, \"Missing id parameter\", http.StatusBadRequest)\n        return\n    }\n    // Process the request...\n}",
		LineNum:  42,
	},
	{
		Language: "Go",
		Repo:     "example/repo4",
		Filename: "model.go",
		Snippet:  "type User struct {\n    ID        int64  `json:\"id\"`\n    Name      string `json:\"name\"`\n    Email     string `json:\"email\"`\n    CreatedAt time.Time `json:\"created_at\"`\n}",
		LineNum:  10,
	},
	{
		Language: "Go",
		Repo:     "example/repo5",
		Filename: "database.go",
		Snippet:  "func Connect() (*sql.DB, error) {\n    connStr := fmt.Sprintf(\"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable\",\n        config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)\n    return sql.Open(\"postgres\", connStr)\n}",
		LineNum:  25,
	},
	{
		Language: "Go",
		Repo:     "example/repo5",
		Filename: "database.go",
		Snippet:  "func Connect() (*sql.DB, error) {\n    connStr := fmt.Sprintf(\"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable\",\n        config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)\n    return sql.Open(\"postgres\", connStr)\n}",
		LineNum:  25,
	},
	{
		Language: "Go",
		Repo:     "example/repo5",
		Filename: "database.go",
		Snippet:  "func Connect() (*sql.DB, error) {\n    connStr := fmt.Sprintf(\"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable\",\n        config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)\n    return sql.Open(\"postgres\", connStr)\n}",
		LineNum:  25,
	},
	{
		Language: "Go",
		Repo:     "example/repo7",
		Filename: "database.go",
		Snippet:  "func Connect() (*sql.DB, error) {\n    connStr := fmt.Sprintf(\"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable\",\n        config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)\n    return sql.Open(\"postgres\", connStr)\n}",
		LineNum:  25,
	},
	{
		Language: "Go",
		Repo:     "example/repo3",
		Filename: "database.go",
		Snippet:  "func Connect() (*sql.DB, error) {\n    connStr := fmt.Sprintf(\"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable\",\n        config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)\n    return sql.Open(\"postgres\", connStr)\n}",
		LineNum:  25,
	},
	{
		Language: "Go",
		Repo:     "example/repo2",
		Filename: "database.go",
		Snippet:  "func Connect() (*sql.DB, error) {\n    connStr := fmt.Sprintf(\"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable\",\n        config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)\n    return sql.Open(\"postgres\", connStr)\n}",
		LineNum:  25,
	},
}
