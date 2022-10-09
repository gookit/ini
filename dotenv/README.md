# Dotenv

Package `dotenv` that supports importing data from files (eg `.env`) to ENV

## Usage

### Load Env

```go
err := dotenv.Load("./", ".env")
// Or use
// err := dotenv.LoadExists("./", ".env")
```

Load from string-map:

```go
err := dotenv.LoadFromMap(map[string]string{
	"ENV_KEY": "value",
	"LOG_LEVEL": "info",
})
```

### Read Env

```go
val := dotenv.Get("ENV_KEY")
// Or use 
// val := os.Getenv("ENV_KEY")

// get int value
intVal := dotenv.Int("LOG_LEVEL")

// get bool value
blVal := dotenv.Bool("OPEN_DEBUG")

// with default value
val := dotenv.Get("ENV_KEY", "default value")
```
