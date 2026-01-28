package constant

import "fmt"

const undefined = "undefined"

// ResolverKind represents the kind of resolver.
type ResolverKind string

const (
	// ResolverKindPlugin is a plugin resolver.
	ResolverKindPlugin ResolverKind = "plugin"

	// ResolverKindTemplate is a template resolver.
	ResolverKindTemplate ResolverKind = "template"
)

func ParseResolverKind(r string) (ResolverKind, error) {
	switch r {
	case "plugin":
		return ResolverKindPlugin, nil
	case "template":
		return ResolverKindTemplate, nil
	default:
		return "", fmt.Errorf("unknown resolver kind provided: %s", r)
	}
}

func (r ResolverKind) String() string {
	switch r {
	case ResolverKindPlugin:
		return "plugin"
	case ResolverKindTemplate:
		return "template"
	default:
		return undefined
	}
}

// ResolverType represents the type of resolver.
type ResolverType string

const (
	// ResolverTypeStep is a step level resolver.
	ResolverTypeStep ResolverType = "step"

	// ResolverTypeStage is a stage level resolver.
	ResolverTypeStage ResolverType = "stage"
)

func ParseResolverType(s string) (ResolverType, error) {
	switch s {
	case "step":
		return ResolverTypeStep, nil
	case "stage":
		return ResolverTypeStage, nil
	default:
		return "", fmt.Errorf("unknown template type provided: %s", s)
	}
}

func (t ResolverType) String() string {
	switch t {
	case ResolverTypeStep:
		return "step"
	case ResolverTypeStage:
		return "stage"
	default:
		return undefined
	}
}
