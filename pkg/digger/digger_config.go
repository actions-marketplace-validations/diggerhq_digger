package digger

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type WorkflowConfiguration struct {
	OnPullRequestPushed []string `yaml:"on_pull_request_pushed"`
	OnPullRequestClosed []string `yaml:"on_pull_request_closed"`
	OnCommitToDefault   []string `yaml:"on_commit_to_default"`
}

type DiggerConfig struct {
	Projects []Project `yaml:"projects"`
}

type Project struct {
	Name                  string                `yaml:"name"`
	Dir                   string                `yaml:"dir"`
	WorkflowConfiguration WorkflowConfiguration `yaml:"workflow_configuration"`
}

func (p *Project) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawProject Project
	raw := rawProject{
		WorkflowConfiguration: WorkflowConfiguration{
			OnPullRequestPushed: []string{"digger plan"},
			OnPullRequestClosed: []string{"digger unlock"},
			OnCommitToDefault:   []string{"digger apply"},
		},
	}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	*p = Project(raw)
	return nil

}

func NewDiggerConfig(workingDir string) (*DiggerConfig, error) {
	config := &DiggerConfig{}
	var fileName string
	if workingDir == "" {
		fileName = "digger.yml"
	} else {
		fileName = workingDir + "/digger.yml"
	}
	if data, err := os.ReadFile(fileName); err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("error parsing digger.yml: %v", err)
		}
	} else {
		config.Projects = make([]Project, 1)
		config.Projects[0] = Project{Name: "default", Dir: ".", WorkflowConfiguration: WorkflowConfiguration{
			OnPullRequestPushed: []string{"digger plan"},
			OnPullRequestClosed: []string{"digger unlock"},
			OnCommitToDefault:   []string{"digger apply"},
		}}
		return config, nil
	}
	return config, nil
}

func (c *DiggerConfig) GetProject(projectName string) *Project {
	for _, project := range c.Projects {
		if projectName == project.Name {
			return &project
		}
	}
	return nil
}

func (c *DiggerConfig) GetProjects(projectName string) []Project {
	if projectName == "" {
		return c.Projects
	}
	project := c.GetProject(projectName)
	if project == nil {
		return nil
	}
	return []Project{*project}
}

func (c *DiggerConfig) GetModifiedProjects(changedFiles []string) []Project {
	var result []Project
	for _, project := range c.Projects {
		for _, file := range changedFiles {
			absoluteFile, _ := filepath.Abs(path.Join("/", file))
			absoluteDir, _ := filepath.Abs(path.Join("/", project.Dir))
			if strings.HasPrefix(absoluteFile, absoluteDir) {
				result = append(result, project)
				break
			}
		}
	}
	return result
}

func (c *DiggerConfig) GetDirectory(projectName string) string {
	project := c.GetProject(projectName)
	if project == nil {
		return ""
	}
	return project.Dir
}

func (c *DiggerConfig) GetWorkflowConfiguration(projectName string) WorkflowConfiguration {
	project := c.GetProject(projectName)
	if project == nil {
		return WorkflowConfiguration{}
	}
	return project.WorkflowConfiguration
}

type File struct {
	Filename string
}
