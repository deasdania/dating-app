package config

import (
	"strings"
)

type AppMetadata struct {
	Namespace string
	Env       Env
	AppEnv    AppEnv
	App       App
}

type Env uint8

const (
	Env_Unknown Env = iota
	Env_Development
	Env_Staging
	Env_Production
)

func (e *Env) String() string {
	if e == nil {
		return ""
	}
	switch *e {
	case Env_Staging:
		return "staging"
	case Env_Production:
		return "production"
	case Env_Development:
		return "development"
	}
	return ""
}

type AppEnv uint8

const (
	AppEnv_Unknown AppEnv = iota
	AppEnv_Sandbox
	AppEnv_Live
)

func (e *AppEnv) String() string {
	if e == nil {
		return ""
	}
	switch *e {
	case AppEnv_Sandbox:
		return "sandbox"
	case AppEnv_Live:
		return "live"
	}
	return ""
}

func (e *AppEnv) Title() string {
	if e == nil {
		return ""
	}
	switch *e {
	case AppEnv_Sandbox:
		return "Sandbox"
	case AppEnv_Live:
		return "Live"
	}
	return ""
}

func ParseAppEnv(appEnv string) AppEnv {
	switch strings.ToLower(appEnv) {
	case "sandbox":
		return AppEnv_Sandbox
	case "live":
		return AppEnv_Live
	default:
		return AppEnv_Unknown
	}
}

type App uint8

const (
	App_Unknown App = iota
	App_Direct
	App_Tap
	App_FraudDetection
	App_AccountOpening
)

func (m *AppMetadata) IsSandbox() bool {
	return m.AppEnv == AppEnv_Sandbox
}

func (m *AppMetadata) IsLive() bool {
	return m.AppEnv == AppEnv_Live
}
