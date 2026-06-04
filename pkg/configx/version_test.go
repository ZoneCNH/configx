package configx

import (
	"os"
	"regexp"
	"testing"
)

func TestVersionMatchesLatestChangelogRelease(t *testing.T) {
	changelog, err := os.ReadFile("../../CHANGELOG.md")
	if err != nil {
		t.Fatalf("read changelog: %v", err)
	}

	latestReleasePattern := regexp.MustCompile(`(?m)^##\s+(v\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?)\b`)
	match := latestReleasePattern.FindStringSubmatch(string(changelog))
	if match == nil {
		t.Fatal("CHANGELOG.md does not contain a release heading")
	}

	if Version != match[1] {
		t.Fatalf("Version = %q, latest CHANGELOG release = %q", Version, match[1])
	}
}

func TestModuleNameContract(t *testing.T) {
	const want = "github.com/ZoneCNH/configx"
	if ModuleName != want {
		t.Fatalf("ModuleName = %q, want %q", ModuleName, want)
	}
}
