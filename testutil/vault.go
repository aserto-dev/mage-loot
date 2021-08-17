package testutil

import (
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/tidwall/gjson"
)

var (
	vault      = deps.BinDepOut("vault")
	vaultCache *bigcache.BigCache
)

func init() {
	var err error
	vaultCache, err = bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	if err != nil {
		panic(err)
	}
}

func VaultTokenRefresh() {
	_, err := vault("token", "renew")
	if err != nil {
		panic(err)
	}
}

// VaultValue returns a value from the vault service
// If query has 1 element, the secret is assumed to be named "integration-testing".
// Otherwise, there should be 2 elements, the first a secret, and
// the second a query passed to gjson, to get a value from the json object returned
// from Vault.
// The first parameter can also be a testing.T. In this case, the function won't panic, but use t.Fail.
func VaultValue(params ...interface{}) string {
	var (
		ok     bool
		t      *testing.T
		secret string
		query  string
	)

	fail := func(msg string) {
		if t != nil {
			t.Log(msg)
			t.Fail()
		} else {
			panic(msg)
		}
	}

	if len(params) > 0 {
		t, ok = params[0].(*testing.T)
		if ok {
			params = params[1:]
		}
	}

	if len(params) == 0 {
		fail("must provide at least one value")
	} else if len(params) == 1 {
		secret = "integration-testing"
		query, ok = params[0].(string)
		if !ok {
			fail("expected first query param to be a string")
		}
	} else if len(params) == 2 {
		secret, ok = params[0].(string)
		if !ok {
			fail("expected first query param to be a string")
		}

		query, ok = params[1].(string)
		if !ok {
			fail("expected second query param to be a string")
		}
	}

	entry, err := vaultCache.Get(secret + "#" + query)
	if err == nil {
		return string(entry)
	}
	if err != bigcache.ErrEntryNotFound {
		panic(err)
	}

	out, err := vault("kv", "get", "--format=json", "kv/"+secret)

	if err != nil {
		panic(err)
	}

	result := gjson.Get(out, "data.data."+query)
	if err != nil {
		panic(err)
	}

	err = vaultCache.Set(secret+"#"+query, []byte(result.String()))
	if err != nil {
		panic(err)
	}

	return result.String()
}
