package ini_test

import (
	"fmt"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/testutil"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/ini/v2"
)

// https://github.com/gookit/ini/issues/88 MapStruct does not parse env, when mapping all data #88
func TestIssues_88(t *testing.T) {
	type MongoDb struct {
		Uri string
	}
	type Config struct {
		MongoDb
	}

	defer ini.ResetStd()

	testutil.MockOsEnv(map[string]string{
		"MONGO_URI": "mongodb://localhost:27017",
	}, func() {
		err := ini.LoadStrings(`
[mongodb]
uri = ${MONGO_URI}
`)
		assert.NoErr(t, err)
	})

	dump.P(ini.Data())
	assert.Eq(t, "mongodb://localhost:27017", ini.String("mongodb.uri"))

	// mapping by key
	mCfg := MongoDb{}
	err := ini.MapStruct("mongodb", &mCfg)
	assert.NoErr(t, err)
	assert.Eq(t, "mongodb://localhost:27017", mCfg.Uri)

	// mapping all data
	cfg := Config{}
	err = ini.MapStruct("", &cfg)
	assert.NoErr(t, err)
	fmt.Println(cfg.MongoDb.Uri)
}
