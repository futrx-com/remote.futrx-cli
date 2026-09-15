package application

import (
	"fmt"
	"regexp"
	"strings"
)

var appID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
var envKey = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

func validateManifestMetadata(manifest Manifest) error {
	if !appID.MatchString(manifest.ID) {
		return fmt.Errorf("application id %q must use lowercase letters, numbers, and hyphens", manifest.ID)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return fmt.Errorf("application.json: name is required")
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return fmt.Errorf("application.json: version is required")
	}
	if len(manifest.Scopes) == 0 {
		return fmt.Errorf("application.json: at least one scope is required")
	}
	for _, scope := range manifest.Scopes {
		if scope != "global" && scope != "project" {
			return fmt.Errorf("application.json: invalid scope %q", scope)
		}
	}
	for _, environmentVariable := range manifest.Env {
		if !envKey.MatchString(environmentVariable.Key) {
			return fmt.Errorf("application.json: invalid env key %q", environmentVariable.Key)
		}
		if environmentVariable.Generate != "" && environmentVariable.Generate != "password" {
			return fmt.Errorf("application.json: unsupported generator %q", environmentVariable.Generate)
		}
	}
	return nil
}

func validateManifestRuntime(manifest Manifest) error {
	if manifest.Port.Internal < 0 || manifest.Port.Internal > 65535 || manifest.Port.DefaultExternal < 0 || manifest.Port.DefaultExternal > 65535 {
		return fmt.Errorf("application.json: ports must be between 1 and 65535")
	}
	if manifest.Port.Internal == 0 && (manifest.Port.DefaultExternal != 0 || manifest.Healthcheck.Command != "") {
		return fmt.Errorf("defaultExternal and healthcheck require port.internal")
	}
	if manifest.Port.Protocol != "" && manifest.Port.Protocol != "tcp" && manifest.Port.Protocol != "udp" {
		return fmt.Errorf("application.json: protocol must be tcp or udp")
	}
	return nil
}
