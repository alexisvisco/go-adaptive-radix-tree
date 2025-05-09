package main

import (
	"testing"
)

func TestNumberOfSlashesFromPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		key      []byte
		current  []byte
		expected interface{}
	}{
		{
			name:     "etc and etc/hosts",
			key:      []byte("/etc"),
			current:  []byte("/etc/hosts"),
			expected: 1,
		},
		{
			name:     "var and var/log/syslog",
			key:      []byte("/var"),
			current:  []byte("/var/log/syslog"),
			expected: 2,
		},
		{
			name:     "root and var/log/syslog",
			key:      []byte("/"),
			current:  []byte("/var/log/syslog"),
			expected: 2,
		},
		{
			name:     "root and var",
			key:      []byte("/"),
			current:  []byte("/var"),
			expected: 0,
		},
		{
			name:     "root and var/log",
			key:      []byte("/"),
			current:  []byte("/var/log"),
			expected: 1,
		},
		{
			name:     "identical paths",
			key:      []byte("/var/log/syslog"),
			current:  []byte("/var/log/syslog"),
			expected: 0,
		},
		{
			name:     "key with trailing slash",
			key:      []byte("/home/"),
			current:  []byte("/home/user"),
			expected: 0,
		},
		{
			name:     "longer path with many components",
			key:      []byte("/usr"),
			current:  []byte("/usr/local/bin/python"),
			expected: 3,
		},
		{
			name:     "empty key",
			key:      []byte(""),
			current:  []byte("/var/log"),
			expected: 2, // Count all slashes
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := numberOfSlashesFromPrefix(tc.key, tc.current)
			if result != tc.expected {
				t.Errorf("numberOfSlashesFromPrefix(%q, %q) = %v; want %v",
					string(tc.key), string(tc.current), result, tc.expected)
			}
		})
	}
}

// Test using table-driven subtests with more edge cases
func TestNumberOfSlashesFromPrefixEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		key      []byte
		current  []byte
		expected interface{}
	}{
		{
			name:     "both empty",
			key:      []byte(""),
			current:  []byte(""),
			expected: 0, // Same path
		},
		{
			name:     "key with only slash",
			key:      []byte("/"),
			current:  []byte("/"),
			expected: 0, // Same path
		},
		{
			name:     "path with consecutive slashes",
			key:      []byte("/var"),
			current:  []byte("/var//log///syslog"),
			expected: 5, // Counts all slashes
		},
		{
			name:     "non-root paths",
			key:      []byte("home"),
			current:  []byte("home/user/docs"),
			expected: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := numberOfSlashesFromPrefix(tc.key, tc.current)
			if result != tc.expected {
				t.Errorf("numberOfSlashesFromPrefix(%q, %q) = %v; want %v",
					string(tc.key), string(tc.current), result, tc.expected)
			}
		})
	}
}
