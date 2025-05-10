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
			expected: 3,
		},
		{
			name:     "root and var",
			key:      []byte("/"),
			current:  []byte("/var"),
			expected: 1,
		},
		{
			name:     "root and var/log",
			key:      []byte("/var/"),
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
			expected: 1,
		},
		{
			name:     "longer path with many components",
			key:      []byte("/usr"),
			current:  []byte("/usr/local/bin/python"),
			expected: 3,
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
