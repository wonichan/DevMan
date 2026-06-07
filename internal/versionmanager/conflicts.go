package versionmanager

func DetectVersionManagerConflict(env Environment, toolKey string) *VersionManagerConflict {
	if env == nil {
		env = RealEnvironment{}
	}

	for _, check := range conflictChecks(toolKey) {
		if check.envVar != "" {
			if value := env.Getenv(check.envVar); value != "" {
				return &VersionManagerConflict{
					ToolKey:  toolKey,
					Manager:  check.manager,
					Evidence: check.envVar + "=" + value,
					Detected: true,
				}
			}
		}
		if check.command != "" {
			if path := env.LookPath(check.command); path != "" {
				return &VersionManagerConflict{
					ToolKey:  toolKey,
					Manager:  check.manager,
					Evidence: check.command + " at " + path,
					Detected: true,
				}
			}
		}
	}
	return nil
}

type conflictCheck struct {
	manager string
	envVar  string
	command string
}

func conflictChecks(toolKey string) []conflictCheck {
	common := []conflictCheck{
		{manager: "asdf", command: "asdf"},
	}
	switch toolKey {
	case "node":
		return append([]conflictCheck{
			{manager: "nvm", envVar: "NVM_HOME"},
			{manager: "fnm", command: "fnm"},
		}, common...)
	case "go":
		return append([]conflictCheck{
			{manager: "gvm", command: "gvm"},
		}, common...)
	case "flutter":
		return append([]conflictCheck{
			{manager: "fvm", command: "fvm"},
		}, common...)
	default:
		return common
	}
}
