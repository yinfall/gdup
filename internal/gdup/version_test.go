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

func TestParseInstalledVersion(t *testing.T) {
	tests := []struct {
		input       string
		expectedVer string
		expectedTyp string
		expectedPh  string
	}{
		{"Godot_v4.3-stable_win64", "4.3", "stable", "stable"},
		{"Godot_v4.3-stable_mono_win64", "4.3_mono", "stable", "stable"},
		{"Godot_v4.4-dev2_linux.x86_64", "4.4-dev2", "dev", "dev2"},
		{"Godot_v4.4-dev2_mono_linux.x86_64", "4.4-dev2_mono", "dev", "dev2"},
		{"Godot_v4.3.1-rc1_win64", "4.3.1-rc1", "rc", "rc1"},
		{"4.3.0-stable", "4.3", "stable", "stable"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			ver, typ, ph := parseInstalledVersion(tc.input)
			if ver != tc.expectedVer || typ != tc.expectedTyp || ph != tc.expectedPh {
				t.Errorf("parseInstalledVersion(%q) = (%q, %q, %q), expected (%q, %q, %q)",
					tc.input, ver, typ, ph, tc.expectedVer, tc.expectedTyp, tc.expectedPh)
			}
		})
	}
}

func TestTableWriter(t *testing.T) {
	table := newTable([]string{"VERSION", "TYPE", "COMMAND"})
	table.Append([]string{colorize("stable") + "4.3" + resetColor, colorize("stable") + "STABLE" + resetColor, "install 4.3"})
	table.Append([]string{colorize("dev") + "* 4.0-alpha1_mono" + resetColor, colorize("dev") + "ALPHA1" + resetColor, "install 4.0-alpha1_mono"})
	table.Render()
}
