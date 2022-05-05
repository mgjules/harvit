package plan

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/mgjules/harvit/logger"
	"gopkg.in/yaml.v2"
)

// Field types.
const (
	FieldTypeText     = "text"
	FieldTypeNumber   = "number"
	FieldTypeDecimal  = "decimal"
	FieldTypeDateTime = "datetime"
)

// Plan defines the parameters for harvesting.
type Plan struct {
	Source string  `yaml:"source" validate:"required,url"`
	Fields []Field `yaml:",flow" validate:"required,dive"`
}

// SetDefaults sets the default values for the plan.
func (p *Plan) SetDefaults() {
	for i := range p.Fields {
		p.Fields[i].SetDefaults()
	}
}

// Field is a single piece of data.
type Field struct {
	Name     string `yaml:"name" validate:"required,alpha"`
	Type     string `yaml:"type" validate:"required,oneof=text number decimal datetime"`
	Selector string `yaml:"selector" validate:"required"`
	Regex    string `yaml:"regex"`
	Format   string `yaml:"format"`
	Timezone string `yaml:"timezone"`
}

// SetDefaults sets the default values for a field.
func (d *Field) SetDefaults() {
	if d.Type == "" {
		d.Type = "text"
	}
}

// Load loads a plan from a file.
func Load(path string) (*Plan, error) {
	raw, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	var plan Plan
	if err := yaml.Unmarshal(raw, &plan); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan file: %w", err)
	}

	plan.SetDefaults()

	validate := validator.New()
	if err := validate.Struct(plan); err != nil {
		return nil, fmt.Errorf("failed to validate plan: %w", err)
	}

	logger.Log.Debugw("loaded plan", "plan", plan)

	return &plan, nil
}
