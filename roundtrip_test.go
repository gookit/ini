package ini_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/ini/v2"
)

// WriteTo emitted values verbatim while LoadStrings trims whitespace and strips
// one matching quote pair, so passwords/paths and quoted or space-padded values
// were silently corrupted on reload. Check the public round-trip is exact.
func TestWriteToLoadStringsRoundTrip(t *testing.T) {
	is := assert.New(t)

	vals := map[string]string{
		"quoted_dq":   `"quoted"`,
		"quoted_sq":   "'quoted'",
		"lead_space":  "  leading",
		"trail_space": "trailing  ",
		"password":    "  s3cr3t ",
		"win_path":    `C:\dir\`,
		"json_val":    `{"a":1}`,
		"url":         "http://x/y?a=b#frag",
		"plain":       "hello",
	}

	c := ini.New()
	is.NoErr(c.SetSection("sec", vals))

	var buf bytes.Buffer
	_, err := c.WriteTo(&buf)
	is.NoErr(err)

	c2 := ini.New()
	is.NoErr(c2.LoadStrings(buf.String()))

	for k, orig := range vals {
		is.Eq(orig, c2.Get("sec."+k), "value %q corrupted on round-trip", k)
	}
}

func TestWriteToAfterOverridingEnvironmentValue(t *testing.T) {
	t.Setenv("INI_RAW_ORIGINAL", "original")
	t.Setenv("INI_RAW_KEEP", "neighbor")

	mutations := []struct {
		name           string
		replaceSection bool
		apply          func(*testing.T, *ini.Ini, string, string, string) error
	}{
		{
			name: "LoadStrings",
			apply: func(t *testing.T, c *ini.Ini, section, key, value string) error {
				return c.LoadStrings(replacementINI(c, section, key, value))
			},
		},
		{
			name: "LoadFiles",
			apply: func(t *testing.T, c *ini.Ini, section, key, value string) error {
				file := filepath.Join(t.TempDir(), "override.ini")
				if err := os.WriteFile(file, []byte(replacementINI(c, section, key, value)), 0600); err != nil {
					return err
				}
				return c.LoadFiles(file)
			},
		},
		{
			name: "Set",
			apply: func(t *testing.T, c *ini.Ini, section, key, value string) error {
				return c.Set(key, value, section)
			},
		},
		{
			name: "SetSection",
			apply: func(t *testing.T, c *ini.Ini, section, key, value string) error {
				return c.SetSection(section, map[string]string{key: value})
			},
		},
		{
			name:           "NewSection",
			replaceSection: true,
			apply: func(t *testing.T, c *ini.Ini, section, key, value string) error {
				return c.NewSection(section, map[string]string{key: value})
			},
		},
		{
			name: "LoadData",
			apply: func(t *testing.T, c *ini.Ini, section, key, value string) error {
				return c.LoadData(map[string]ini.Section{section: {key: value}})
			},
		},
	}

	for _, mutation := range mutations {
		for _, named := range []bool{false, true} {
			name := mutation.name + "/default"
			if named {
				name = mutation.name + "/named-ignore-case"
			}
			t.Run(name, func(t *testing.T) {
				c := ini.New()
				section, key, keyPath, keepPath := c.DefSection(), "value", "value", "keep"
				if named {
					c.WithOptions(ini.IgnoreCase)
					section, key, keyPath, keepPath = "SERVICE", "VALUE", "service.value", "service.keep"
				}
				input := replacementINI(c, section, key, "${INI_RAW_ORIGINAL}") +
					"keep = ${INI_RAW_KEEP}\n[untouched]\nneighbor = ${INI_RAW_KEEP}\n"
				if err := c.LoadStrings(input); err != nil {
					t.Fatal(err)
				}
				if err := mutation.apply(t, c, section, key, "replacement"); err != nil {
					t.Fatal(err)
				}
				if got := c.Get(keyPath); got != "replacement" {
					t.Fatalf("current value = %q, want replacement", got)
				}

				output := exportINI(t, c)
				if strings.Contains(output, "${INI_RAW_ORIGINAL}") {
					t.Fatalf("export restored the previous expression:\n%s", output)
				}
				if !strings.Contains(output, "neighbor = ${INI_RAW_KEEP}") {
					t.Fatalf("export lost the untouched expression:\n%s", output)
				}
				keptExpressions := 2
				if mutation.replaceSection {
					keptExpressions = 1
				}
				if got := strings.Count(output, "${INI_RAW_KEEP}"); got != keptExpressions {
					t.Fatalf("export kept %d neighbor expressions, want %d:\n%s", got, keptExpressions, output)
				}
				reloaded := ini.New()
				if err := reloaded.LoadStrings(output); err != nil {
					t.Fatal(err)
				}
				if got := reloaded.Get(keyPath); got != "replacement" {
					t.Fatalf("reloaded value = %q, want replacement", got)
				}
				if mutation.replaceSection {
					if reloaded.HasKey(keepPath) {
						t.Fatal("section replacement retained a removed key")
					}
				} else if got := reloaded.Get(keepPath); got != "neighbor" {
					t.Fatalf("untouched section value = %q, want neighbor", got)
				}
			})
		}
	}
}

func TestWriteToAfterDeletingEnvironmentValue(t *testing.T) {
	t.Setenv("INI_RAW_ORIGINAL", "original")

	t.Run("Delete-Set", func(t *testing.T) {
		c := ini.New()
		if err := c.LoadStrings("value = ${INI_RAW_ORIGINAL}"); err != nil {
			t.Fatal(err)
		}
		if !c.Delete("value") {
			t.Fatal("Delete did not remove the value")
		}
		if err := c.Set("value", "replacement"); err != nil {
			t.Fatal(err)
		}
		output := exportINI(t, c)
		if strings.Contains(output, "${INI_RAW_ORIGINAL}") || !strings.Contains(output, "value = replacement") {
			t.Fatalf("export did not retain the replacement for the deleted value:\n%s", output)
		}
	})

	t.Run("DelSection-LoadData", func(t *testing.T) {
		c := ini.NewWithOptions(ini.IgnoreCase)
		if err := c.LoadStrings("[service]\nvalue = ${INI_RAW_ORIGINAL}"); err != nil {
			t.Fatal(err)
		}
		if !c.DelSection("SERVICE") || !c.IsEmpty() {
			t.Fatal("DelSection did not empty the config")
		}
		if err := c.LoadData(map[string]ini.Section{"service": {"value": "replacement"}}); err != nil {
			t.Fatal(err)
		}
		output := exportINI(t, c)
		if strings.Contains(output, "${INI_RAW_ORIGINAL}") || !strings.Contains(output, "value = replacement") {
			t.Fatalf("export did not retain the replacement section value:\n%s", output)
		}
	})
}

func TestWriteToAfterRejectedReadonlyOverride(t *testing.T) {
	t.Setenv("INI_RAW_ORIGINAL", "original")
	mutations := []struct {
		name  string
		apply func(*ini.Ini) error
	}{
		{"Set", func(c *ini.Ini) error { return c.Set("value", "replacement", "service") }},
		{"SetSection", func(c *ini.Ini) error {
			return c.SetSection("service", map[string]string{"value": "replacement"})
		}},
		{"NewSection", func(c *ini.Ini) error {
			return c.NewSection("service", map[string]string{"value": "replacement"})
		}},
		{"LoadData", func(c *ini.Ini) error {
			return c.LoadData(map[string]ini.Section{"service": {"value": "replacement"}})
		}},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			c := ini.NewWithOptions(ini.Readonly)
			if err := c.LoadStrings("[service]\nvalue = ${INI_RAW_ORIGINAL}"); err != nil {
				t.Fatal(err)
			}
			if err := mutation.apply(c); err == nil {
				t.Fatal("Readonly override succeeded")
			}
			if got := c.Get("service.value"); got != "original" {
				t.Fatalf("rejected write changed the value to %q", got)
			}
			if output := exportINI(t, c); !strings.Contains(output, "value = ${INI_RAW_ORIGINAL}") {
				t.Fatalf("rejected write lost the original expression:\n%s", output)
			}
		})
	}
}

func TestWriteToAfterUpdatingEnvironmentExpression(t *testing.T) {
	t.Setenv("INI_RAW_ORIGINAL", "original")
	t.Setenv("INI_RAW_UPDATED", "updated")
	c := ini.New()
	if err := c.LoadStrings("value = ${INI_RAW_ORIGINAL}", "value = ${INI_RAW_UPDATED}"); err != nil {
		t.Fatal(err)
	}
	if got := c.Get("value"); got != "updated" {
		t.Fatalf("current value = %q, want updated", got)
	}
	output := exportINI(t, c)
	if strings.Contains(output, "${INI_RAW_ORIGINAL}") || !strings.Contains(output, "value = ${INI_RAW_UPDATED}") {
		t.Fatalf("export did not preserve the updated expression:\n%s", output)
	}
}

func TestWriteToWithUnderscoresInSectionAndKey(t *testing.T) {
	t.Setenv("INI_RAW_ORIGINAL", "original")
	t.Setenv("INI_RAW_UPDATED", "updated")

	mutations := []struct {
		name  string
		apply func(*testing.T, *ini.Ini)
		gone  bool
	}{
		{"unchanged", func(t *testing.T, c *ini.Ini) {}, false},
		{"LoadStrings", func(t *testing.T, c *ini.Ini) {
			if err := c.LoadStrings("[a]\nb_c = replacement"); err != nil {
				t.Fatal(err)
			}
		}, false},
		{"Set", func(t *testing.T, c *ini.Ini) {
			if err := c.Set("b_c", "replacement", "a"); err != nil {
				t.Fatal(err)
			}
		}, false},
		{"SetSection", func(t *testing.T, c *ini.Ini) {
			if err := c.SetSection("a", map[string]string{"b_c": "replacement"}); err != nil {
				t.Fatal(err)
			}
		}, false},
		{"NewSection", func(t *testing.T, c *ini.Ini) {
			if err := c.NewSection("a", map[string]string{"b_c": "replacement"}); err != nil {
				t.Fatal(err)
			}
		}, false},
		{"Delete", func(t *testing.T, c *ini.Ini) {
			if !c.Delete("a.b_c") {
				t.Fatal("Delete did not remove the value")
			}
		}, true},
		{"DelSection", func(t *testing.T, c *ini.Ini) {
			if !c.DelSection("a") {
				t.Fatal("DelSection did not remove the section")
			}
		}, true},
	}

	for _, reverse := range []bool{false, true} {
		for _, expression := range []bool{false, true} {
			name := "section-first/literal"
			if reverse {
				name = "key-first/literal"
			}
			value, currentValue := "literal", "literal"
			if expression {
				name = strings.TrimSuffix(name, "literal") + "expression"
				value, currentValue = "${INI_RAW_UPDATED}", "updated"
			}
			for _, mutation := range mutations {
				t.Run(name+"/"+mutation.name, func(t *testing.T) {
					left := "[a_b]\nc = ${INI_RAW_ORIGINAL}\n"
					right := "[a]\nb_c = " + value + "\n"
					input := left + right
					if reverse {
						input = right + left
					}
					c := ini.New()
					if err := c.LoadStrings(input); err != nil {
						t.Fatal(err)
					}
					mutation.apply(t, c)
					output := exportINI(t, c)
					if got := c.Get("a_b.c"); got != "original" {
						t.Fatalf("WriteTo changed the current value to %q", got)
					}
					reloaded := ini.NewWithOptions(func(o *ini.Options) { o.ParseEnv = false })
					if err := reloaded.LoadStrings(output); err != nil {
						t.Fatal(err)
					}
					if got := reloaded.Get("a_b.c"); got != "${INI_RAW_ORIGINAL}" {
						t.Fatalf("untouched expression = %q, want ${INI_RAW_ORIGINAL}:\n%s", got, output)
					}
					if mutation.gone {
						if reloaded.HasKey("a.b_c") {
							t.Fatalf("export retained the removed value:\n%s", output)
						}
						return
					}
					want, wantCurrent := "replacement", "replacement"
					if mutation.name == "unchanged" {
						want, wantCurrent = value, currentValue
					}
					if got := reloaded.Get("a.b_c"); got != want {
						t.Fatalf("other export value = %q, want %q:\n%s", got, want, output)
					}
					if got := c.Get("a.b_c"); got != wantCurrent {
						t.Fatalf("WriteTo changed the other current value to %q, want %q", got, wantCurrent)
					}
				})
			}
		}
	}
}

func replacementINI(c *ini.Ini, section, key, value string) string {
	input := key + " = " + value + "\n"
	if section != c.DefSection() {
		input = "[" + section + "]\n" + input
	}
	return input
}

func exportINI(t *testing.T, c *ini.Ini) string {
	t.Helper()
	var output bytes.Buffer
	if _, err := c.WriteTo(&output); err != nil {
		t.Fatal(err)
	}
	return output.String()
}
