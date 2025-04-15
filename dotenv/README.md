# Dotenv

Package `dotenv` that supports importing data from files (eg `.env`) to ENV

- filename support simple glob pattern. eg: ".env.*", "*.env"

## Install

```bash
go get github.com/gookit/ini/v2/dotenv
```

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

## Functions API

```go
func DontUpperEnvKey()
// get env value
func Bool(name string, defVal ...bool) (val bool)
func Get(name string, defVal ...string) (val string)
func Int(name string, defVal ...int) (val int)
// load env files/data
func Load(dir string, filenames ...string) (err error)
func LoadExistFiles(filePaths ...string) error
func LoadExists(dir string, filenames ...string) error
func LoadFiles(filePaths ...string) (err error)
func LoadFromMap(kv map[string]string) (err error)
// extra methods
func ClearLoaded()
func LoadedFiles() []string
func LoadedData() map[string]string
func Reset()
```

## License

**MIT**
