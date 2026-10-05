package route

import (
	"reflect"
	"testing"
)

func TestMatch(t *testing.T) {
	cases := []struct {
		pattern, path string
		ok            bool
		values        []string
	}{
		{"/api/v1/auth/login", "/api/v1/auth/login", true, nil},
		{"/api/v1/auth/login", "/api/v1/auth/logout", false, nil},
		{"/users/:id", "/users/42", true, []string{"42"}},
		{"/users/:id", "/users", false, nil},
		{"/users/:id", "/users/42/x", false, nil},
		{"/files/*", "/files/a/b", true, []string{"a/b"}},
	}
	for _, c := range cases {
		_, values, ok := match(splitPath(c.pattern), splitPath(c.path))
		if ok != c.ok || (ok && !reflect.DeepEqual(values, c.values)) {
			t.Errorf("match(%q,%q) = %v %v, want %v %v", c.pattern, c.path, values, ok, c.values, c.ok)
		}
	}
}
