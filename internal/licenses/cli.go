package licenses

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func RunInteractiveLicenseCLI() {
	if len(ThirdPartyLicenses) == 0 && ProjectLicense == "" {
		fmt.Println("No licenses found.")
		return
	}

	entries := make([]LicenseEntry, len(ThirdPartyLicenses))
	copy(entries, ThirdPartyLicenses)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LibraryName < entries[j].LibraryName
	})

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n=== GDUP - Godot Version Manager ===")
		fmt.Println("Author: yinfall")
		fmt.Println("License: MIT License")
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("%3d. %s\n", 0, "View full GDUP license")
		
		fmt.Println("\nThird-party open source libraries used by GDUP:")
		fmt.Println(strings.Repeat("-", 50))
		for i, entry := range entries {
			fmt.Printf("%3d. %s\n", i+1, entry.LibraryName)
		}
		fmt.Println(strings.Repeat("-", 50))
		fmt.Print("Enter number to view full license, or 'q' to quit: ")

		input, err := reader.ReadString('\n')
		if err != nil && strings.TrimSpace(input) == "" {
			// Prevent infinite loop if stdin is closed (e.g., EOF)
			break
		}
		input = strings.TrimSpace(input)

		if input == "q" || input == "quit" || input == "exit" {
			break
		}

		var choice int
		_, err = fmt.Sscanf(input, "%d", &choice)
		if err != nil || choice < 0 || choice > len(entries) {
			fmt.Println("\n[!] Invalid input. Please enter a valid number or 'q' to quit.")
			continue
		}

		if choice == 0 {
			fmt.Printf("\n=== License for GDUP ===\n\n")
			if ProjectLicense != "" {
				fmt.Println(ProjectLicense)
			} else {
				fmt.Println("License information not found.")
			}
		} else {
			selected := entries[choice-1]
			fmt.Printf("\n=== License for %s ===\n\n", selected.LibraryName)
			fmt.Println(selected.Content)
		}
		
		fmt.Println(strings.Repeat("=", 50))
		fmt.Print("\nPress Enter to return to the list...")
		_, err = reader.ReadString('\n')
		if err != nil {
			break
		}
	}
}
