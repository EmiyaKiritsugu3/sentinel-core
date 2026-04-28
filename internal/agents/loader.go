package agents

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// Loader handles parsing of Smart Agent Artifacts (.md files with YAML frontmatter).
type Loader struct {
	validate *validator.Validate
}

// NewLoader initializes a new agent loader.
func NewLoader() *Loader {
	return &Loader{
		validate: validator.New(),
	}
}

// LoadAgent reads an agent definition from a file using Standard #01 (Buffered Reads).
func (l *Loader) LoadAgent(path string) (*AgentDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open agent file: %w", err)
	}
	defer file.Close()

	var yamlBuilder strings.Builder
	var mdBuilder strings.Builder
	scanner := bufio.NewScanner(file)
	
	// Increase buffer size for large prompts (Standard #01)
	const maxCapacity = 1 * 1024 * 1024
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	inYAML := false
	frontmatterFound := false
	yamlClosed := false

	for scanner.Scan() {
		line := scanner.Text()
		
		if line == "---" {
			if !frontmatterFound {
				inYAML = true
				frontmatterFound = true
				continue
			} else if !yamlClosed {
				inYAML = false
				yamlClosed = true
				continue
			}
		}

		if inYAML {
			yamlBuilder.WriteString(line + "\n")
		} else {
			mdBuilder.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading agent file: %w", err)
	}

	var def AgentDefinition
	if frontmatterFound {
		if err := yaml.Unmarshal([]byte(yamlBuilder.String()), &def); err != nil {
			return nil, fmt.Errorf("failed to parse agent frontmatter: %w", err)
		}
	} else {
		return nil, fmt.Errorf("agent definition must contain YAML frontmatter delimited by ---")
	}

	// Validate configuration
	if err := l.validate.Struct(&def); err != nil {
		return nil, fmt.Errorf("invalid agent configuration: %w", err)
	}

	// Attach Markdown body as System Prompt
	def.SystemPrompt = strings.TrimSpace(mdBuilder.String())

	return &def, nil
}
