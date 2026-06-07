package versionmanager

import (
	"fmt"
	"os"
)

type fakeEnvironment struct {
	vars            map[string]string
	paths           map[string]string
	dirs            map[string]bool
	files           map[string]bool
	writes          map[string][]byte
	writePerms      map[string]os.FileMode
	mkdirs          map[string]os.FileMode
	exeDir          string
	runOutput       string
	runErr          error
	runCommands     []fakeRunCommand
	userPathEntries []string
	userEnvSets     map[string]string
}

type fakeRunCommand struct {
	command string
	args    []string
}

func newFakeEnvironment() *fakeEnvironment {
	return &fakeEnvironment{
		vars:        map[string]string{},
		paths:       map[string]string{},
		dirs:        map[string]bool{},
		files:       map[string]bool{},
		writes:      map[string][]byte{},
		writePerms:  map[string]os.FileMode{},
		mkdirs:      map[string]os.FileMode{},
		userEnvSets: map[string]string{},
	}
}

func (f *fakeEnvironment) Getenv(key string) string {
	return f.vars[key]
}

func (f *fakeEnvironment) LookPath(command string) string {
	return f.paths[command]
}

func (f *fakeEnvironment) DirExists(path string) bool {
	return f.dirs[path]
}

func (f *fakeEnvironment) FileExists(path string) bool {
	return f.files[path]
}

func (f *fakeEnvironment) ExecutableDir() string {
	return f.exeDir
}

func (f *fakeEnvironment) WriteFile(path string, data []byte, perm os.FileMode) error {
	f.writes[path] = append([]byte(nil), data...)
	f.writePerms[path] = perm
	return nil
}

func (f *fakeEnvironment) MkdirAll(path string, perm os.FileMode) error {
	f.mkdirs[path] = perm
	f.dirs[path] = true
	return nil
}

func (f *fakeEnvironment) SetUserEnv(key string, value string) error {
	f.vars[key] = value
	f.userEnvSets[key] = value
	return nil
}

func (f *fakeEnvironment) EnsureUserPathEntry(entry string) error {
	f.userPathEntries = append(f.userPathEntries, entry)
	return nil
}

func (f *fakeEnvironment) Run(command string, args ...string) (string, error) {
	copiedArgs := append([]string(nil), args...)
	f.runCommands = append(f.runCommands, fakeRunCommand{command: command, args: copiedArgs})
	return f.runOutput, f.runErr
}

func (f *fakeEnvironment) assertNoMutation(t testingT) {
	t.Helper()
	if len(f.writes) != 0 {
		t.Fatalf("writes occurred before validation completed: %#v", f.writes)
	}
	if len(f.mkdirs) != 0 {
		t.Fatalf("mkdirs occurred before validation completed: %#v", f.mkdirs)
	}
	if len(f.userPathEntries) != 0 {
		t.Fatalf("PATH entries occurred before validation completed: %#v", f.userPathEntries)
	}
	if len(f.userEnvSets) != 0 {
		t.Fatalf("user env writes occurred before validation completed: %#v", f.userEnvSets)
	}
}

type testingT interface {
	Helper()
	Fatalf(format string, args ...any)
}

func errFakeRunFailed() error {
	return fmt.Errorf("verification failed")
}

type fakeVersionRegistry struct {
	versions []ManagedVersion
	saved    []ManagedVersion
}

func newFakeVersionRegistry(versions []ManagedVersion) *fakeVersionRegistry {
	copied := append([]ManagedVersion(nil), versions...)
	return &fakeVersionRegistry{versions: copied}
}

func (f *fakeVersionRegistry) ListToolVersions(toolKey string) ([]ManagedVersion, error) {
	var versions []ManagedVersion
	for _, version := range f.versions {
		if version.ToolKey == toolKey {
			versions = append(versions, version)
		}
	}
	return versions, nil
}

func (f *fakeVersionRegistry) SaveToolVersion(v *ManagedVersion) error {
	copied := *v
	f.saved = append(f.saved, copied)
	for i := range f.versions {
		if f.versions[i].ID == v.ID {
			f.versions[i] = copied
			return nil
		}
	}
	f.versions = append(f.versions, copied)
	return nil
}

func (f *fakeVersionRegistry) GetInstallStrategy(toolKey string) (*InstallStrategy, error) {
	return nil, nil
}

func (f *fakeVersionRegistry) SaveInstallStrategy(strategy InstallStrategy) error {
	return nil
}

func (f *fakeVersionRegistry) savedByID(id int64) *ManagedVersion {
	for i := range f.saved {
		if f.saved[i].ID == id {
			return &f.saved[i]
		}
	}
	return nil
}
