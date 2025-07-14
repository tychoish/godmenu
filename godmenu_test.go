package godmenu

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"
)

func TestDmenu(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		t.Run("CheckNilSafe", func(t *testing.T) {
			var st *set
			if st != nil {
				t.Fatal("expected nil")
			}

			if st.check("key") {
				t.Fatal("should report false for everything")
			}
		})
		t.Run("ExistanceChecks", func(t *testing.T) {
			st := set{}
			st.init([]string{"one", "two"})
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
			st.init([]string{"one", "one", "two"})
			t.Log(st.Len(), st.items, st.set)
			if st.Len() != 2 {
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
				s := set{}
				s.init([]string{"abc def"})

				out, err := s.processOutput([]byte(" abc def \n"), nil)
				t.Log("out", out, "err", err)
				if err != nil || out != "abc def" {
					t.Error(err)
				}
			})
			t.Run("Passthrough", func(t *testing.T) {
				s := set{}
				s.init([]string{"abc def"})

				out, err := s.processOutput([]byte(" abc def \n"), context.Canceled)
				t.Log("out", out, "err", err)
				if !errors.Is(err, context.Canceled) || out != "" {
					t.Error(err)
				}
			})
		})
		t.Run("Selections", func(t *testing.T) {
			t.Run("Sorted", func(t *testing.T) {
				stdin := string(newset([]string{"abc", "def", "111", "999"}).rendered(true))
				t.Log(stdin)
				if stdin != "111\n999\nabc\ndef" {
					t.Fail()
				}
			})
			t.Run("InOrder", func(t *testing.T) {
				stdin := string(newset([]string{"abc", "def", "111", "999\n"}).rendered(false))
				t.Log(stdin)
				if stdin != "abc\ndef\n111\n999" {
					t.Fail()
				}
			})
			t.Run("Trim", func(t *testing.T) {
				stdin := string(newset([]string{" abc\n", " def \n", "\n111 ", "999"}).rendered(true))
				t.Log(stdin)
				if stdin != "111\n999\nabc\ndef" {
					t.Fail()
				}

			})

		})

	})
	t.Run("ProcessOutput", func(t *testing.T) {
		t.Run("PermissiveMode", func(t *testing.T) {
			var st *set

			t.Run("WithoutError", func(t *testing.T) {
				st = newset([]string{"hello"})
				out, err := st.processOutput([]byte("hello"), nil)

				t.Logf("out=%q, err=%q", out, err)

				if err != nil && out != "hello" {
					t.Fail()
				}

			})
			t.Run("Error", func(t *testing.T) {
				out, err := st.processOutput([]byte("hello"), context.Canceled)
				t.Logf("out=%q, err=%q", out, err)

				if !errors.Is(err, context.Canceled) && out != "" {
					t.Fail()
				}
			})
			t.Run("EmptyOutput", func(t *testing.T) {
				out, err := st.processOutput([]byte(""), context.Canceled)
				t.Logf("out=%q, err=%q", out, err)

				if !errors.Is(err, ErrSelectionMissing) && !errors.Is(err, context.Canceled) && out != "" {
					t.Fail()
				}
			})
			t.Run("NilOutput", func(t *testing.T) {
				out, err := st.processOutput(nil, context.Canceled)
				t.Logf("out=%q, err=%q", out, err)

				if !errors.Is(err, ErrSelectionMissing) && !errors.Is(err, context.Canceled) && out != "" {
					t.Fail()
				}
			})
		})
		t.Run("StrictMode", func(t *testing.T) {
			st := set{}
			st.init([]string{"one"})

			t.Run("SelectionExists", func(t *testing.T) {
				out, err := st.processOutput([]byte("one"), nil)
				t.Logf("out=%q, err=%q", out, err)
				if err != nil && out != "" {
					t.Fail()
				}
			})
			t.Run("NoOutput", func(t *testing.T) {
				out, err := st.processOutput([]byte(""), nil)
				t.Logf("out=%q, err=%q", out, err)

				if !errors.Is(err, ErrSelectionMissing) && out != "" {
					t.Fail()
				}
			})
			t.Run("NilOutput", func(t *testing.T) {
				out, err := st.processOutput(nil, nil)
				t.Logf("out=%q, err=%q", out, err)

				if !errors.Is(err, ErrSelectionMissing) && out != "" {
					t.Fail()
				}
			})
			t.Run("SelectionMissing", func(t *testing.T) {
				out, err := st.processOutput([]byte("two"), nil)
				t.Logf("out=%q, err=%q", out, err)

				if !errors.Is(err, ErrSelectionUnknown) && out != "" {
					t.Fail()
				}
			})
		})
	})
	t.Run("Integration", func(t *testing.T) {
		t.Run("Strict", func(t *testing.T) {
			t.Run("Empty", func(t *testing.T) {
				out, err := Do(t.Context(), Options{})
				t.Logf("out=%q, err=%q", out, err)
				if err == nil || !errors.Is(err, ErrConfigurationInvalid) || out != "" {
					t.Fail()
				}
			})
			t.Run("Duplicates", func(t *testing.T) {
				out, err := Do(t.Context(), Options{Selections: []string{"a", "a", "b"}})
				t.Logf("out=%q, err=%q", out, err)
				if err == nil || !errors.Is(err, ErrConfigurationInvalid) || out != "" {
					t.Fail()
				}
			})
			t.Run("ZeroString", func(t *testing.T) {
				out, err := Do(t.Context(), Options{Selections: []string{"", "a", "b"}})
				t.Logf("out=%q, err=%q", out, err)
				if err == nil || !errors.Is(err, ErrConfigurationInvalid) || out != "" {
					t.Fail()
				}
			})
		})
		t.Run("NonStrict", func(t *testing.T) {
			t.Run("Empty", func(t *testing.T) {
				out, err := Do(t.Context(), Options{})
				t.Logf("out=%q, err=%q", out, err)
				if !errors.Is(err, ErrConfigurationInvalid) || out != "" {
					t.Fail()
				}
			})
			t.Run("Duplicates", func(t *testing.T) {
				out, err := Do(t.Context(), Options{Selections: []string{"a", "a", "b"}})
				t.Logf("out=%q, err=%q", out, err)
				if !errors.Is(err, ErrConfigurationInvalid) || out != "" {
					t.Fail()
				}
			})
			t.Run("ZeroString", func(t *testing.T) {
				out, err := Do(t.Context(), Options{Selections: []string{"", "a", "b"}})
				t.Logf("out=%q, err=%q", out, err)
				if !errors.Is(err, ErrConfigurationInvalid) || out != "" {
					t.Fail()
				}
			})

		})

	})
	t.Run("FunctionaArugments ", func(t *testing.T) {
		t.Run("Duplicates", func(t *testing.T) {
			out, err := Run(t.Context(), Selections("a", "a", "a", "a"), Sorted(), DMenuPrompt("godmenu =>>"))
			t.Logf("out=%q, err=%q", out, err)
			if !errors.Is(err, ErrConfigurationInvalid) || out != "" {
				t.Fail()
			}

		})
		t.Run("NoConfiguration", func(t *testing.T) {
			out, err := Run(t.Context())
			t.Logf("out=%q, err=%q", out, err)
			if !errors.Is(err, ErrConfigurationInvalid) || out != "" {
				t.Fail()
			}
		})

		t.Run("NoConfiguration", func(t *testing.T) {
			out, err := Run(t.Context())
			t.Logf("out=%q, err=%q", out, err)
			if !errors.Is(err, ErrConfigurationInvalid) || out != "" {
				t.Fail()
			}
		})
		t.Run("Working", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
			t.Cleanup(cancel)

			out, err := Run(ctx, Items("a", "b", "c"))
			t.Logf("out=%q, err=%q", out, err)

			exerr := &exec.ExitError{}
			if !errors.As(err, &exerr) {
				t.Error(err)
			}
			// if it was signaled this is -1
			if exerr.ExitCode() != -1 {
				t.Error(exerr.ExitCode())
			}

			if out != "" {
				t.Error(out)
			}
		})

	})
}
