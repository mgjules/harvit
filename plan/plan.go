package plan

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/go-playground/validator"
	"github.com/mgjules/harvit/logger"
	"gopkg.in/yaml.v2"
)

// Datum types.
const (
	DatumTypeText    = "text"
	DatumTypeNumber  = "number"
	DatumTypeDecimal = "decimal"
)

// Plan defines the parameters for harvesting.
type Plan struct {
	Scheme string  `yaml:"scheme" validate:"required,alpha"`
	Domain string  `yaml:"domain" validate:"required,fqdn"`
	Path   string  `yaml:"path" validate:"required"`
	Data   []Datum `yaml:",flow" validate:"required,dive"`
}

// SetDefaults sets the default values for the plan.
func (p *Plan) SetDefaults() {
	if p.Scheme == "" {
		p.Scheme = "http"
	}

	for i := range p.Data {
		p.Data[i].SetDefaults()
	}
}

// Datum is a single piece of data.
type Datum struct {
	Name     string `yaml:"name" validate:"required,alphanum"`
	Type     string `yaml:"type" validate:"required,oneof=text number decimal"`
	Selector string `yaml:"selector" validate:"required"`
}

// SetDefaults sets the default values for a datum.
func (d *Datum) SetDefaults() {
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
