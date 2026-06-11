package scanner

import (
	"devman/internal/models"
	"devman/internal/registry"
	"devman/internal/versionmanager"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Scanner interface {
	Name() string
	Detect() ([]models.EnvInstance, []models.EnvPath, error)
}

type ScanOptions struct {
	CustomScanPaths []string
}

type Engine struct {
	reg      *registry.Registry
	scanners []Scanner
}

func NewEngine(reg *registry.Registry) *Engine {
	return &Engine{
		reg: reg,
		scanners: []Scanner{
			&NodeScanner{},
			&PythonScanner{},
			&JavaScanner{},
			&GoScanner{},
			&FlutterScanner{},
			&RustScanner{},
			&DockerScanner{},
			&PnpmScanner{},
			&YarnScanner{},
			&BunScanner{},
		},
	}
}

func (e *Engine) Register(s Scanner) {
	e.scanners = append(e.scanners, s)
}

func (e *Engine) ScanAll() ([]models.EnvSummary, error) {
	return e.ScanAllWithOptions(ScanOptions{})
}

func (e *Engine) ScanAllWithOptions(opts ScanOptions) ([]models.EnvSummary, error) {
	start := time.Now()
	logrus.WithFields(logrus.Fields{"scanner_count": len(e.scanners), "custom_scan_paths": len(opts.CustomScanPaths)}).Info("environment scan started")
	var summaries []models.EnvSummary
	scannerEnvKeys := make(map[string]struct{}, len(e.scanners))
	detectedKeys := make(map[string]struct{}, len(e.scanners))
	for _, s := range e.scanners {
		scannerStart := time.Now()
		entry := logrus.WithField("scanner", s.Name())
		entry.Info("scanner started")
		desc := descriptorForScanner(s)
		env := desc.Env
		instances, paths, err := s.Detect()
		if err != nil {
			entry.WithError(err).Error("scanner failed")
			continue
		}
		scannerEnvKeys[env.Key] = struct{}{}

		if len(opts.CustomScanPaths) > 0 {
			customInst, customPaths := detectCustomPaths(desc, opts.CustomScanPaths)
			instances = append(instances, customInst...)
			paths = append(paths, customPaths...)
		}

		if len(instances) > 0 && desc.SyncedToolKey != "" {
			if err := e.reg.SyncScannedToolVersions(desc.SyncedToolKey, buildScannedToolVersions(desc.SyncedToolKey, instances)); err != nil {
				entry.WithError(err).WithField("tool_key", desc.SyncedToolKey).Warn("failed to sync scanned tool versions")
			}
		}

		if len(instances) == 0 {
			entry.WithField("duration_ms", time.Since(scannerStart).Milliseconds()).Info("scanner completed with no instances")
			continue
		}
		detectedKeys[env.Key] = struct{}{}

		// Save env metadata
		if err := e.reg.SaveEnv(&env); err != nil {
			entry.WithError(err).WithField("env_key", env.Key).Error("failed to save scanned environment")
			continue
		}

		// Clear old data
		if err := e.reg.ClearInstances(env.ID); err != nil {
			entry.WithError(err).WithField("env_id", env.ID).Warn("failed to clear old instances")
		}
		if err := e.reg.ClearPaths(env.ID); err != nil {
			entry.WithError(err).WithField("env_id", env.ID).Warn("failed to clear old paths")
		}

		// Save instances
		for i := range instances {
			instances[i].EnvID = env.ID
			if err := e.reg.SaveInstance(&instances[i]); err != nil {
				entry.WithError(err).WithField("install_path", instances[i].InstallPath).Warn("failed to save scanned instance")
			}
		}

		// Save paths and compute sizes
		for i := range paths {
			paths[i].EnvID = env.ID
			if paths[i].SizeBytes == 0 {
				paths[i].SizeBytes = DirSize(paths[i].Path)
			}
			if err := e.reg.SavePath(&paths[i]); err != nil {
				entry.WithError(err).WithField("path", paths[i].Path).Warn("failed to save scanned path")
			}
		}

		totalSize := int64(0)
		for _, p := range paths {
			totalSize += p.SizeBytes
		}

		health := models.HealthHealthy
		if totalSize > 5*1024*1024*1024 {
			health = models.HealthWarning
		}

		summaries = append(summaries, models.EnvSummary{
			Env:       env,
			Instances: instances,
			Paths:     paths,
			TotalSize: totalSize,
			Health:    health,
		})
		entry.WithFields(logrus.Fields{"instances": len(instances), "paths": len(paths), "total_size": totalSize, "duration_ms": time.Since(scannerStart).Milliseconds()}).Info("scanner completed")
	}
	e.pruneStaleEnvs(scannerEnvKeys, detectedKeys)
	logrus.WithFields(logrus.Fields{"summary_count": len(summaries), "duration_ms": time.Since(start).Milliseconds()}).Info("environment scan completed")
	return summaries, nil
}

func (e *Engine) pruneStaleEnvs(scannerEnvKeys, detectedKeys map[string]struct{}) {
	for key := range scannerEnvKeys {
		if _, detected := detectedKeys[key]; detected {
			continue
		}
		existing, err := e.reg.GetEnvByKey(key)
		if err != nil {
			logrus.WithError(err).WithField("env_key", key).Warn("failed to look up env for stale cleanup")
			continue
		}
		if existing == nil {
			continue
		}
		if existing.IsManaged {
			logrus.WithField("env_key", key).Debug("preserving managed env without current detection")
			continue
		}
		if err := e.reg.DeleteEnv(key); err != nil {
			logrus.WithError(err).WithField("env_key", key).Warn("failed to delete stale unmanaged env")
			continue
		}
		logrus.WithField("env_key", key).Info("removed stale unmanaged env not detected by current scan")
	}
}

func detectCustomPaths(desc ScannerDescriptor, customPaths []string) ([]models.EnvInstance, []models.EnvPath) {
	var instances []models.EnvInstance
	var paths []models.EnvPath

	if len(desc.CustomExeNames) == 0 {
		return instances, paths
	}

	for _, cp := range customPaths {
		cp = strings.TrimSpace(cp)
		if cp == "" {
			continue
		}
		fullPath := filepath.Join(cp, "bin")
		for _, name := range desc.CustomExeNames {
			candidate := filepath.Join(fullPath, name)
			if PathExists(candidate) {
				instances = append(instances, models.EnvInstance{
					Version:     "custom path",
					InstallPath: fullPath,
					IsDefault:   false,
					IsActive:    true,
					Source:      "custom",
				})
				paths = append(paths, models.EnvPath{
					Type:      models.PathInstall,
					Path:      fullPath,
					IsMovable: true,
				})
				break
			}
			candidate = filepath.Join(cp, name)
			if PathExists(candidate) {
				instances = append(instances, models.EnvInstance{
					Version:     "custom path",
					InstallPath: cp,
					IsDefault:   false,
					IsActive:    true,
					Source:      "custom",
				})
				paths = append(paths, models.EnvPath{
					Type:      models.PathInstall,
					Path:      cp,
					IsMovable: true,
				})
				break
			}
		}
	}

	return instances, paths
}

func buildScannedToolVersions(toolKey string, instances []models.EnvInstance) []versionmanager.ManagedVersion {
	versions := make([]versionmanager.ManagedVersion, 0, len(instances))
	for _, instance := range instances {
		installPath := strings.TrimSpace(instance.InstallPath)
		if installPath == "" {
			continue
		}
		versions = append(versions, versionmanager.ManagedVersion{
			ToolKey:      toolKey,
			Version:      instance.Version,
			InstallPath:  installPath,
			BinPath:      scannedBinPath(toolKey, installPath),
			Source:       versionmanager.SourceExternal,
			IsDefault:    instance.IsDefault,
			IsActive:     instance.IsActive,
			CanDelete:    false,
			DeletePolicy: versionmanager.DeletePolicyForceRequired,
			DetectedAt:   detectedAtForInstance(instance),
		})
	}
	return versions
}

func scannedBinPath(toolKey string, installPath string) string {
	tool, ok := versionmanager.ToolByKey(toolKey)
	if !ok {
		return ""
	}
	targets, err := versionmanager.ShimTargets(toolKey, installPath)
	if err != nil {
		return ""
	}
	primaryShim := strings.TrimSuffix(tool.PrimaryExe, filepath.Ext(tool.PrimaryExe)) + ".cmd"
	return targets[primaryShim]
}

func detectedAtForInstance(instance models.EnvInstance) time.Time {
	if instance.DetectedAt.IsZero() {
		return time.Now()
	}
	return instance.DetectedAt
}
