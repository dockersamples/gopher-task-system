package types

// TaskDefinition represents a task definition document.
type TaskDefinition struct {
	Version string `yaml:"version,omitempty"`
	Tasks   []Task `yaml:"tasks,omitempty"`
}

// Task provides a task definition for gopher.
type Task struct {
	Name    string   `yaml:"name,omitempty"`
	Runner  string   `yaml:"runner,omitempty"`
	Command []string `yaml:"command,omitempty"`
	Cleanup bool     `yaml:"cleanup,omitempty"`
}
