package application

import "encoding/json"

type Manifest struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Category    string          `json:"category,omitempty"`
	Version     string          `json:"version"`
	Icon        string          `json:"icon,omitempty"`
	Scopes      []string        `json:"scopes"`
	Port        Port            `json:"port,omitempty"`
	Env         []Env           `json:"env,omitempty"`
	Service     string          `json:"service,omitempty"`
	Install     string          `json:"install,omitempty"`
	Healthcheck Healthcheck     `json:"healthcheck,omitempty"`
	Connection  json.RawMessage `json:"connection,omitempty"`
	UI          json.RawMessage `json:"ui,omitempty"`
	Backend     json.RawMessage `json:"backend,omitempty"`
	Base        string          `json:"base,omitempty"`
	HostTools   json.RawMessage `json:"hostTools,omitempty"`
}

type Port struct {
	Internal        int    `json:"internal,omitempty"`
	DefaultExternal int    `json:"defaultExternal,omitempty"`
	Protocol        string `json:"protocol,omitempty"`
	BindAddress     string `json:"bindAddress,omitempty"`
}

type Env struct {
	Key      string `json:"key"`
	Label    string `json:"label,omitempty"`
	Required bool   `json:"required,omitempty"`
	Secret   bool   `json:"secret,omitempty"`
	Default  string `json:"default,omitempty"`
	Generate string `json:"generate,omitempty"`
}

type Healthcheck struct {
	Command string `json:"command,omitempty"`
}
