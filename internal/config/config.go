package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Nexus         NexusConfig `yaml:"nexus"`
	Rules         []Rule      `yaml:"rules"`
	ProtectedTags []string    `yaml:"protected_tags"`
	Schedule      string      `yaml:"schedule"`
	DryRun        bool        `yaml:"dry_run"`
	LogFile       string      `yaml:"log_file"`
}

type NexusConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Timeout  int    `yaml:"timeout"`
}

type Rule struct {
	Name         string `yaml:"name"`
	Regex        string `yaml:"regex"`
	Keep         int    `yaml:"keep"`
	compiledRegex *regexp.Regexp
}

func (r *Rule) Matches(imageName string) bool {
	if r.compiledRegex == nil {
		return false
	}
	return r.compiledRegex.MatchString(imageName)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Compile regex patterns
	for i := range cfg.Rules {
		compiled, err := regexp.Compile(cfg.Rules[i].Regex)
		if err != nil {
			return nil, fmt.Errorf("invalid regex in rule '%s': %w", cfg.Rules[i].Name, err)
		}
		cfg.Rules[i].compiledRegex = compiled
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Nexus.URL == "" {
		return fmt.Errorf("nexus.url is required")
	}
	if c.Nexus.Username == "" {
		return fmt.Errorf("nexus.username is required")
	}
	if c.Nexus.Password == "" {
		return fmt.Errorf("nexus.password is required")
	}
	if len(c.Rules) == 0 {
		return fmt.Errorf("at least one rule is required")
	}
	for _, rule := range c.Rules {
		if rule.Keep < 1 {
			return fmt.Errorf("rule '%s': keep must be at least 1", rule.Name)
		}
	}
	if c.LogFile == "" {
		c.LogFile = "deletion_log.csv"
	}
	return nil
}

func (c *Config) IsProtected(tag string) bool {
	for _, protected := range c.ProtectedTags {
		if protected == tag {
			return true
		}
	}
	return false
}

func (c *Config) GetKeepCount(imageName string) (int, string, bool) {
	for _, rule := range c.Rules {
		if rule.Matches(imageName) {
			return rule.Keep, rule.Name, true
		}
	}
	return 0, "", false
}
