package grdep

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"

	"github.com/berquerant/execx"
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

type Validatable interface {
	Validate() error
}

var (
	_ Validatable = &Config{}
	_ Validatable = &Matcher{}
	_ Validatable = &NamedMatcher{}
	_ Validatable = &CSelector{}
	_ Validatable = &NSelector{}
	_ Validatable = &Normalizers{}
)

type Config struct {
	// Ignore files with matching paths.
	Ignores []Matcher `yaml:"ignore,omitempty" json:"ignore,omitempty"`
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
	Regex    *Regexp  `yaml:"r,omitempty" json:"r,omitempty"`
	Not      *Regexp  `yaml:"not,omitempty" json:"not,omitempty"`
	Shell    string   `yaml:"sh,omitempty" json:"sh,omitempty"`
	Template string   `yaml:"tmpl,omitempty" json:"tmpl,omitempty"`
	Value    []string `yaml:"val,omitempty" json:"val,omitempty"`

	shellScript *execx.Script `yaml:"-" json:"-"`
	mux         sync.Mutex    `yaml:"-" json:"-"`
}

func (m Matcher) countSettings() int {
	var c int
	if m.Regex != nil {
		c++
	}
	if m.Not != nil {
		c++
	}
	if m.Shell != "" {
		c++
	}
	if m.Template != "" {
		c++
	}
	if len(m.Value) > 0 {
		c++
	}
	return c
}

func (m Matcher) Validate() error {
	switch m.countSettings() {
	case 0:
		return fmt.Errorf("%w: empty matcher", ErrInvalidConfig)
	case 1:
		if m.Template != "" {
			return fmt.Errorf("%w: tmpl requires r", ErrInvalidConfig)
		}
		return nil
	case 2:
		switch {
		case m.Regex != nil:
			if m.Template != "" || len(m.Value) > 0 {
				return nil
			}
		case m.Not != nil:
			if len(m.Value) > 0 {
				return nil
			}
		}
	}

	return fmt.Errorf("%w: only (r, tmpl), (r, val), (not, val) can be specified at the same time", ErrInvalidConfig)
}

type NamedMatcher struct {
	Name    string    `yaml:"name,omitempty" json:"name,omitempty"`
	Matcher []Matcher `yaml:"matcher" json:"matcher"`
}

func (m NamedMatcher) Validate() error {
	for i, x := range m.Matcher {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("%w: %s matcher[%d]", err, m.Name, i)
		}
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
	Name     string    `yaml:"name,omitempty" json:"name,omitempty"`
	Filename []Matcher `yaml:"filename,omitempty" json:"filename,omitempty"`
	Text     []Matcher `yaml:"text,omitempty" json:"text,omitempty"`
}

func (s CSelector) Validate() error {
	if !XOR(len(s.Filename) > 0, len(s.Text) > 0) {
		return fmt.Errorf("%w: category(%s) should have only either filename or text", ErrInvalidConfig, s.Name)
	}

	for i, x := range s.Filename {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("%w: category(%s) filename[%d]", err, s.Name, i)
		}
	}

	for i, x := range s.Text {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("%w: category(%s) text[%d]", err, s.Name, i)
		}
	}

	return nil
}

type NSelector struct {
	Name     string    `yaml:"name,omitempty" json:"name,omitempty"`
	Category Regexp    `yaml:"category" json:"category"`
	Matcher  []Matcher `yaml:"matcher" json:"matcher"`
}

func (s NSelector) Validate() error {
	for i, x := range s.Matcher {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("%w: node(%s) selector[%d]", err, s.Name, i)
		}
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
		return fmt.Errorf("%w: value is %s", err, value.Value)
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
	return json.Marshal(r.Unwrap().String())
}
