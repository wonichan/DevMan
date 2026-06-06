package versionmanager

type fakeEnvironment struct {
	vars  map[string]string
	paths map[string]string
	dirs  map[string]bool
	files map[string]bool
}

func newFakeEnvironment() *fakeEnvironment {
	return &fakeEnvironment{
		vars:  map[string]string{},
		paths: map[string]string{},
		dirs:  map[string]bool{},
		files: map[string]bool{},
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
