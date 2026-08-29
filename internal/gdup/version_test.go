package gdup

import (
	"testing"
)

func TestQueryBestMatch(t *testing.T) {
	candidates := []string{
		"Godot_v4.3-stable_win64",
		"Godot_v4.3-stable_mono_win64",
		"Godot_v4.3.1-stable_win64",
		"Godot_v4.3.2-stable_win64",
		"Godot_v4.4-dev1_win64",
		"MyStudio_Godot_4.3_custom_win64",
		"Godot_v4.3-stable_macos.universal",
	}

	tests := []struct {
		name     string
		query    string
		platform string
		expected string
	}{
		{
			name:     "Greedy patch matching for 4.3",
			query:    "4.3",
			platform: "win64",
			expected: "Godot_v4.3.2-stable_win64",
		},
		{
			name:     "Mono exclusivity blocks mono from standard query",
			query:    "4.3",
			platform: "win64",
			expected: "Godot_v4.3.2-stable_win64",
		},
		{
			name:     "Mono requested explicitly",
			query:    "4.3_mono",
			platform: "win64",
			expected: "Godot_v4.3-stable_mono_win64",
		},
		{
			name:     "Highest version fallback",
			query:    "4.3",
			platform: "win64",
			// Wait, 4.3 matches 4.3.0, 4.3.1, 4.3.2.
			// MatchesTokens uses strings.Contains(target, "4.3").
			// Both 4.3-stable, 4.3.1-stable, 4.3.2-stable contain "4.3".
			// Highest should be 4.3.2-stable.
			expected: "Godot_v4.3.2-stable_win64",
		},
		{
			name:     "Custom magic matching",
			query:    "4.3_custom",
			platform: "win64",
			expected: "MyStudio_Godot_4.3_custom_win64",
		},
		{
			name:     "Cross platform isolation",
			query:    "4.3",
			platform: "macos",
			expected: "Godot_v4.3-stable_macos.universal",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := QueryBestMatch(candidates, tc.query, tc.platform)
			if result != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
