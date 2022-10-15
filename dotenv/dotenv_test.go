package dotenv

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestLoad(t *testing.T) {
	defer ClearLoaded()
	err := Load("./testdata", "not-exist", ".env")
	assert.Err(t, err)

	assert.Eq(t, "", os.Getenv("DONT_ENV_TEST"))

	err = Load("./testdata")
	assert.NoErr(t, err)
	assert.Eq(t, "blog", os.Getenv("DONT_ENV_TEST"))
	assert.Eq(t, "blog", Get("DONT_ENV_TEST"))
	_ = os.Unsetenv("DONT_ENV_TEST") // clear

	err = Load("./testdata", "error.ini")
	assert.Err(t, err)

	err = Load("./testdata", "invalid_key.ini")
	assert.Err(t, err)

	assert.Eq(t, "def-val", Get("NOT-EXIST", "def-val"))
}

func TestLoadFiles(t *testing.T) {
	defer Reset()
	assert.Err(t, LoadFiles("./testdata/not-exist"))
	assert.Eq(t, "", os.Getenv("DONT_ENV_TEST"))

	err := LoadFiles("./testdata/.env")

	assert.NoErr(t, err)
	assert.NotEmpty(t, LoadedData())
	assert.Eq(t, "blog", os.Getenv("DONT_ENV_TEST"))
	assert.Eq(t, "blog", Get("DONT_ENV_TEST"))
}

func TestLoadExists(t *testing.T) {
	defer Reset()
	assert.Eq(t, "", os.Getenv("DONT_ENV_TEST"))

	err := LoadExists("./testdata", "not-exist", ".env")

	assert.NoErr(t, err)
	assert.Eq(t, "blog", os.Getenv("DONT_ENV_TEST"))
	assert.Eq(t, "blog", Get("DONT_ENV_TEST"))
}

func TestLoadExistFiles(t *testing.T) {
	defer Reset()
	assert.Eq(t, "", os.Getenv("DONT_ENV_TEST"))

	err := LoadExistFiles("./testdata/not-exist", "./testdata/.env")

	assert.NoErr(t, err)
	assert.Eq(t, "blog", os.Getenv("DONT_ENV_TEST"))
	assert.Eq(t, "blog", Get("DONT_ENV_TEST"))
}

func TestLoadFromMap(t *testing.T) {
	assert.Eq(t, "", os.Getenv("DONT_ENV_TEST"))

	err := LoadFromMap(map[string]string{
		"DONT_ENV_TEST":  "blog",
		"dont_env_test1": "val1",
		"dont_env_test2": "23",
		"dont_env_bool":  "true",
	})

	assert.NoErr(t, err)

	// fmt.Println(os.Environ())
	envStr := fmt.Sprint(os.Environ())
	assert.Contains(t, envStr, "DONT_ENV_TEST=blog")
	assert.Contains(t, envStr, "DONT_ENV_TEST1=val1")

	assert.Eq(t, "blog", Get("DONT_ENV_TEST"))
	assert.Eq(t, "blog", os.Getenv("DONT_ENV_TEST"))
	assert.Eq(t, "val1", Get("DONT_ENV_TEST1"))
	assert.Eq(t, 23, Int("DONT_ENV_TEST2"))
	assert.True(t, Bool("dont_env_bool"))

	assert.Eq(t, "val1", Get("dont_env_test1"))
	assert.Eq(t, 23, Int("dont_env_test2"))

	assert.Eq(t, 20, Int("dont_env_test1", 20))
	assert.Eq(t, 20, Int("dont_env_not_exist", 20))
	assert.False(t, Bool("dont_env_not_exist", false))

	// check cache
	assert.Contains(t, LoadedData(), "DONT_ENV_TEST2")

	// clear
	ClearLoaded()
	assert.Eq(t, "", os.Getenv("DONT_ENV_TEST"))
	assert.Eq(t, "", Get("DONT_ENV_TEST1"))

	err = LoadFromMap(map[string]string{
		"": "val",
	})
	assert.Err(t, err)
}

func TestDontUpperEnvKey(t *testing.T) {
	assert.Eq(t, "", os.Getenv("DONT_ENV_TEST"))

	DontUpperEnvKey()

	err := LoadFromMap(map[string]string{
		"dont_env_test": "val",
	})

	assert.Contains(t, fmt.Sprint(os.Environ()), "dont_env_test=val")
	assert.NoErr(t, err)
	assert.Eq(t, "val", Get("dont_env_test"))

	// on windows, os.Getenv() not case sensitive
	if runtime.GOOS == "windows" {
		assert.Eq(t, "val", Get("DONT_ENV_TEST"))
	} else {
		assert.Eq(t, "", Get("DONT_ENV_TEST"))
	}

	UpperEnvKey = true // revert
	ClearLoaded()
}
