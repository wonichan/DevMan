package versionmanager

import "os"

type fakeEnvironment struct {
	vars            map[string]string
	paths           map[string]string
	dirs            map[string]bool
	writes          map[string][]byte
	writePerms      map[string]os.FileMode
	mkdirs          map[string]os.FileMode
	exeDir          string
	runOutput       string
	runCommands     []fakeRunCommand
	userPathEntries []string
}

type fakeRunCommand struct {
	command string
	args    []string
}

func newFakeEnvironment() *fakeEnvironment {
	return &fakeEnvironment{
		vars:       map[string]string{},
		paths:      map[string]string{},
		dirs:       map[string]bool{},
		writes:     map[string][]byte{},
		writePerms: map[string]os.FileMode{},
		mkdirs:     map[string]os.FileMode{},
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
	return nil
}

func (f *fakeEnvironment) EnsureUserPathEntry(entry string) error {
	f.userPathEntries = append(f.userPathEntries, entry)
	return nil
}

func (f *fakeEnvironment) Run(command string, args ...string) (string, error) {
	copiedArgs := append([]string(nil), args...)
	f.runCommands = append(f.runCommands, fakeRunCommand{command: command, args: copiedArgs})
	return f.runOutput, nil
}
