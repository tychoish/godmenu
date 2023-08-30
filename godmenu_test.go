package godmenu

import (
	"context"
	"errors"
	"testing"
)

func TestDmenu(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		t.Run("CheckNilSafe", func(t *testing.T) {
			var st set
			if st != nil {
				t.Fatal("expected nil")
			}

			if st.check("key") {
				t.Fatal("should report false for everything")
			}
		})
		t.Run("ExistanceChecks", func(t *testing.T) {
			st := set{}
			st.add("one")
			st.add("two")
			t.Log(st)
			if !st.check("one") {
				t.Fail()
			}
			if st.check("three") {
				t.Fail()
			}
		})
		t.Run("Extend", func(t *testing.T) {
			st := set{}
			st.extend([]string{"one", "one", "two"})
			t.Log(len(st), st)
			if len(st) != 2 {
				t.Fail()
			}
			if !st.check("one") || !st.check("two") {
				t.Fail()
			}
		})
	})
	t.Run("Rendering", func(t *testing.T) {
		t.Run("CommandOutput", func(t *testing.T) {
			t.Run("Trim", func(t *testing.T) {
				out, err := renderCommandOutput([]byte(" abc def \n"), nil)
				t.Log(out, err)
				if err != nil || out != "abc def" {
					t.Fail()
				}
			})
			t.Run("Passthrough", func(t *testing.T) {
				out, err := renderCommandOutput([]byte(" abc def \n"), context.Canceled)
				t.Log(out, err)
				if err != context.Canceled || out != "abc def" {
					t.Fail()
				}
			})
		})
		t.Run("Selections", func(t *testing.T) {
			t.Run("Sorted", func(t *testing.T) {
				stdin := renderSelections(true, []string{"abc", "def", "111", "999"})
				t.Log(stdin)
				if stdin != "111\n999\nabc\ndef" {
					t.Fail()
				}
			})
			t.Run("InOrder", func(t *testing.T) {
				stdin := renderSelections(false, []string{"abc", "def", "111", "999\n"})
				t.Log(stdin)
				if stdin != "abc\ndef\n111\n999" {
					t.Fail()
				}
			})
			t.Run("Trim", func(t *testing.T) {
				stdin := renderSelections(false, []string{" abc\n", " def \n", "\n111 ", "999"})
				t.Log(stdin)
				if stdin != "abc\ndef\n111\n999" {
					t.Fail()
				}

			})

		})

	})
	t.Run("ProcessOutput", func(t *testing.T) {
		t.Run("PermissiveMode", func(t *testing.T) {
			var st set

			if st != nil {
				t.Fatal("expected nil")
			}

			t.Run("WithoutError", func(t *testing.T) {
				out, err := st.processOutput("hello", nil)
				t.Log(out, err)

				if err != nil && out != "hello" {
					t.Fail()
				}

			})
			t.Run("Error", func(t *testing.T) {
				out, err := st.processOutput("hello", context.Canceled)
				t.Log(out, err)

				if !errors.Is(err, context.Canceled) && out != "" {
					t.Fail()
				}
			})
			t.Run("EmptyOutput", func(t *testing.T) {
				out, err := st.processOutput("", context.Canceled)
				t.Log(out, err)

				if !errors.Is(err, ErrSelectionMissing) && !errors.Is(err, context.Canceled) && out != "" {
					t.Fail()
				}
			})
		})
		t.Run("StrictMode", func(t *testing.T) {
			st := set{}
			st.add("one")

			t.Run("SelectionExists", func(t *testing.T) {
				out, err := st.processOutput("one", nil)
				t.Log(out, err)
				if err != nil && out != "" {
					t.Fail()
				}
			})
			t.Run("NoOutput", func(t *testing.T) {
				out, err := st.processOutput("", nil)
				t.Log(out, err)

				if !errors.Is(err, ErrSelectionMissing) && out != "" {
					t.Fail()
				}
			})
			t.Run("SelectionMissing", func(t *testing.T) {
				out, err := st.processOutput("two", nil)
				t.Log(out, err)

				if !errors.Is(err, ErrSelectionUnknown) && out != "" {
					t.Fail()
				}
			})
		})
	})
	t.Run("Integration", func(t *testing.T) {
		t.Run("StrictEmpty", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			out, err := RunDMenu(ctx, Options{Strict: true})
			t.Log(out, err)
			if !errors.Is(err, ErrConfigurationInvalid) || out != "" {
				t.Fail()
			}
		})
	})
}
