package grdep

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

func ParseConfig(fileOrText string) (*Config, error) {
	c, fErr := parseConfigFile(fileOrText)
	if fErr == nil {
		return c, nil
	}
	c, err := parseConfigText(fileOrText)
	if err == nil {
		return c, nil
	}
	return nil, errors.Join(err, fErr)
}

func parseConfigFile(configFile string) (*Config, error) {
	fp, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return NewConfigParser().Parse(fp)
}

func parseConfigText(configText string) (*Config, error) {
	return NewConfigParser().Parse(bytes.NewBufferString(configText))
}

func NewConfigParser() *ConfigParser {
	return &ConfigParser{}
}

type ConfigParser struct{}

func (p ConfigParser) Parse(r io.Reader) (*Config, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	c, err := p.parse(b)
	if err != nil {
		return nil, err
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (p ConfigParser) parse(b []byte) (*Config, error) {
	c, yErr := p.parseYAML(b)
	if yErr == nil {
		return c, nil
	}
	c, err := p.parseJSON(b)
	if err == nil {
		return c, nil
	}
	return nil, errors.Join(err, yErr)
}

func (ConfigParser) parseYAML(b []byte) (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (ConfigParser) parseJSON(b []byte) (*Config, error) {
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

type Config struct {
	// Ignore files with matching paths.
	Ignores []Regexp `yaml:"ignore,omitempty" json:"ignore,omitempty"`
	// Select file category.
	Categories []CSelector `yaml:"category" json:"category"`
	// Find nodes corresponding to categories.
	Nodes []NSelector `yaml:"node" json:"node"`
	// Normalize categories and nodes.
	Normalizers Normalizers `yaml:"normalizer,omitempty" json:"normalizer,omitempty"`
}

func (c Config) Validate() error {
	for i, x := range c.Categories {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("%w: category[%d]", err, i)
		}
	}

	for i, x := range c.Nodes {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("%w: node[%d]", err, i)
		}
	}

	if err := c.Normalizers.Validate(); err != nil {
		return err
	}

	return nil
}

func (c Config) Add(other Config) Config {
	return Config{
		Ignores:    append(c.Ignores, other.Ignores...),
		Categories: append(c.Categories, other.Categories...),
		Nodes:      append(c.Nodes, other.Nodes...),
		Normalizers: Normalizers{
			Categories: append(c.Normalizers.Categories, other.Normalizers.Categories...),
			Nodes:      append(c.Normalizers.Nodes, other.Normalizers.Nodes...),
		},
	}
}

var (
	ErrInvalidConfig = errors.New("InvalidConfig")
)

type Matcher struct {
	Regex    Regexp   `yaml:"regex" json:"regex"`
	Template string   `yaml:"template,omitempty" json:"template,omitempty"`
	Value    []string `yaml:"value,omitempty" json:"value,omitempty"`
}

func (m Matcher) Validate() error {
	return nil
}

type NamedMatcher struct {
	Name    string   `yaml:"name,omitempty" json:"name,omitempty"`
	Matcher *Matcher `yaml:"matcher" json:"matcher"`
}

func (m NamedMatcher) Validate() error {
	if err := m.Matcher.Validate(); err != nil {
		return fmt.Errorf("%w: %s", err, m.Name)
	}
	return nil
}

type Normalizers struct {
	Categories []NamedMatcher `yaml:"category,omitempty" json:"category,omitempty"`
	Nodes      []NamedMatcher `yaml:"node,omitempty" json:"node,omitempty"`
}

func (n Normalizers) Validate() error {
	for i, x := range n.Categories {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("%w: category normalizer[%d]", err, i)
		}
	}

	for i, x := range n.Nodes {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("%w: node normalizer[%d]", err, i)
		}
	}

	return nil
}

type CSelector struct {
	Name     string   `yaml:"name,omitempty" json:"name,omitempty"`
	Filename *Matcher `yaml:"filename,omitempty" json:"filename,omitempty"`
	Text     *Matcher `yaml:"text,omitempty" json:"text,omitempty"`
}

func (s CSelector) Validate() error {
	if !XOR(s.Filename == nil, s.Text == nil) {
		return fmt.Errorf("%w: category(%s) should have only either filename or text", ErrInvalidConfig, s.Name)
	}

	if s.Filename != nil {
		if err := s.Filename.Validate(); err != nil {
			return fmt.Errorf("%w: category(%s)", err, s.Name)
		}
		return nil
	}

	if err := s.Text.Validate(); err != nil {
		return fmt.Errorf("%w: category(%s)", err, s.Name)
	}
	return nil
}

type NSelector struct {
	Name     string   `yaml:"name,omitempty" json:"name,omitempty"`
	Category Regexp   `yaml:"category" json:"category"`
	Selector *Matcher `yaml:"selector" json:"selector"`
}

func (s NSelector) Validate() error {
	if s.Selector == nil {
		return fmt.Errorf("%w: node(%s) should have selector", ErrInvalidConfig, s.Name)
	}
	if err := s.Selector.Validate(); err != nil {
		return fmt.Errorf("%w: node(%s)", err, s.Name)
	}
	return nil
}

type Regexp regexp.Regexp

func NewRegexp(pattern string) Regexp {
	return Regexp(*regexp.MustCompile(pattern))
}

func (r Regexp) Unwrap() *regexp.Regexp {
	x := regexp.Regexp(r)
	return &x
}

var (
	ErrNotScalarNode = errors.New("NotScalarNode")
)

func (r *Regexp) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return ErrNotScalarNode
	}

	v, err := regexp.Compile(value.Value)
	if err != nil {
		return err
	}
	*r = Regexp(*v)
	return nil
}

func (r Regexp) MarshalYAML() (any, error) {
	return r.Unwrap().String(), nil
}

func (r *Regexp) UnmarshalJSON(b []byte) error {
	v, err := regexp.Compile(string(b))
	if err != nil {
		return err
	}
	*r = Regexp(*v)
	return nil
}

func (r Regexp) MarshalJSON() ([]byte, error) {
	return []byte(r.Unwrap().String()), nil
}
