package godmenu

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type set struct {
	set   map[string]int
	items []string
}

func newset(in []string) *set { s := &set{}; return s.init(in) }
func (s *set) Len() int       { return len(s.set) }

func (s *set) init(in []string) *set {
	s.set = make(map[string]int, len(in))
	s.items = in
	for idx := range in {
		k := strings.TrimSpace(s.items[idx])
		if k == "" {
			continue
		}
		s.set[k] = idx
		s.items[idx] = k
	}
	return s
}

func (s *set) validate() error {
	if len(s.items) == 0 {
		return fmt.Errorf("must define selections: %w", ErrConfigurationInvalid)
	}
	if diff := len(s.items) - len(s.set); diff != 0 {
		return fmt.Errorf("found %d duplicate selections: %w", diff, ErrConfigurationInvalid)
	}
	return nil
}

func (s *set) check(key string) bool {
	if s == nil || s.set == nil || key == "" {
		return false
	}
	_, ok := s.set[key]
	return ok
}

func (s *set) rendered(shouldSort bool) []byte {
	out := s.selections()
	if shouldSort {
		sort.Strings(out)
	}
	return []byte(strings.Join(out, "\n"))
}

func (s set) selections() []string { return append(make([]string, 0, len(s.set)), s.items...) }

func (s set) processOutput(data []byte, err error) (string, error) {
	switch {
	case err != nil && len(data) != 0:
		return "", fmt.Errorf("dmenu failed [%s]: %w", string(data), err)
	case err != nil && len(data) == 0:
		return "", fmt.Errorf("dmenu error [%w]: %w ", err, ErrSelectionMissing)
	case len(data) == 0:
		return "", ErrSelectionMissing
	}

	out := string(bytes.TrimSpace(data))

	if out == "" {
		return "", ErrSelectionMissing
	}

	if !s.check(out) {
		return "", fmt.Errorf("value %q was not provided: %w", out, ErrSelectionUnknown)
	}

	return out, nil
}
