package sa

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// CoolingSchedule represents the type of cooling schedule
type CoolingSchedule string

const (
	CoolingLinear      CoolingSchedule = "linear"
	CoolingExponential CoolingSchedule = "exponential"
	CoolingPolynomial  CoolingSchedule = "polynomial"
)

// Config holds configuration parameters for simulated annealing
type Config struct {
	Tmax           float64         `yaml:"Tmax"`
	Tmin           float64         `yaml:"Tmin"`
	NSteps         int             `yaml:"nsteps"`
	NStepsPerT     int             `yaml:"nsteps_per_T"`
	Cooling        CoolingSchedule `yaml:"cooling"`
	Alpha          float64         `yaml:"alpha"`
	N              float64         `yaml:"n"` // Polynomial exponent
	PositionDelta  float64         `yaml:"position_delta"`
	AngleDelta     float64         `yaml:"angle_delta"`
	RandomSeed     int64           `yaml:"random_state"`
	LogFreq        int             `yaml:"log_freq"`
	OverlapPenalty float64         `yaml:"overlap_penalty"` // λ multiplier for penalty-based SA
}

// LoadConfig loads SA configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse wrapper structure (config.yaml has nested "params" key)
	var wrapper struct {
		Params Config `yaml:"params"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		// Try parsing directly as Config
		var config Config
		if err2 := yaml.Unmarshal(data, &config); err2 != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
		return &config, nil
	}

	return &wrapper.Params, nil
}

// DefaultConfig returns a default SA configuration
func DefaultConfig() *Config {
	return &Config{
		Tmax:           20.0, // Higher temperature to allow initial overlaps
		Tmin:           1e-6, // Lower minimum temperature for fine-tuning
		NSteps:         500,  // More temperature steps
		NStepsPerT:     100,  // More iterations per temperature (Total: 1M steps)
		Cooling:        CoolingExponential,
		Alpha:          0.99,
		N:              4,
		PositionDelta:  0.05, // Slightly larger initial moves
		AngleDelta:     15.0,
		RandomSeed:     0,
		LogFreq:        10000, // Logging frequency
		OverlapPenalty: 50.0,  // Stronger penalty to enforce valid solutions eventually
	}
}
