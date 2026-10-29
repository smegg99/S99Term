// version_test.go

package s99term

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestDeclaredComesFromTheVersionFile(t *testing.T) {
	body, err := os.ReadFile("VERSION")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSpace(string(body))

	if want == "" {
		t.Fatal("VERSION is empty")
	}
	if got := Declared(); got != want {
		t.Fatalf("Declared() = %q, want %q", got, want)
	}
	if strings.Count(want, ".") != 2 {
		t.Fatalf("VERSION = %q, want MAJOR.MINOR.PATCH", want)
	}
}

// A tagged commit whose VERSION disagrees ships a module that misreports itself.
func TestTagAgreesWithVersionFile(t *testing.T) {
	tag, ok := exactTag()
	if !ok {
		t.Skip("HEAD carries no tag")
	}

	if want := "v" + Declared(); tag != want {
		t.Fatalf("HEAD is tagged %s but VERSION says %s; one of them is wrong",
			tag, Declared())
	}
}

// exactTag returns the tag on HEAD, or false outside a worktree or on an untagged commit.
func exactTag() (string, bool) {
	git, err := exec.LookPath("git")
	if err != nil {
		return "", false
	}
	if err := exec.Command(git, "rev-parse", "--git-dir").Run(); err != nil {
		return "", false
	}

	out, err := exec.Command(git, "describe", "--tags", "--exact-match").Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}
