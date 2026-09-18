package compiler

import (
	"testing"

	harness "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
)

func TestCreateRunStepPrivileged(t *testing.T) {
	tests := []struct {
		name string
		spec *harness.StepRun
		want bool
	}{
		{
			name: "top level privileged",
			spec: &harness.StepRun{
				Privileged: true,
				Container:  &harness.Container{Image: "alpine"},
			},
			want: true,
		},
		{
			name: "container privileged",
			spec: &harness.StepRun{
				Container: &harness.Container{
					Image:      "alpine",
					Privileged: true,
				},
			},
			want: true,
		},
		{
			name: "not privileged",
			spec: &harness.StepRun{
				Container: &harness.Container{Image: "alpine"},
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			step := createRunStep(&harness.Step{Name: "test"}, test.spec)
			if step.Privileged != test.want {
				t.Fatalf("want privileged %v, got %v", test.want, step.Privileged)
			}
		})
	}
}
