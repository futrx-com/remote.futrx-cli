package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var ApplicationID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
var environmentKey = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

func ValidateMetadata(manifest Manifest) error {
	if !ApplicationID.MatchString(manifest.ID) {
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
	for _, variable := range manifest.Env {
		if !environmentKey.MatchString(variable.Key) {
			return fmt.Errorf("application.json: invalid env key %q", variable.Key)
		}
		if variable.Generate != "" && variable.Generate != "password" {
			return fmt.Errorf("application.json: unsupported generator %q", variable.Generate)
		}
	}
	return nil
}

func ValidateRuntime(manifest Manifest) error {
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
