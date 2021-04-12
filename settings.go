package main

import (
	"go.uber.org/config"
)

const configFolder string = "config/"

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
		Ssl  struct {
			Certificate string `yaml:"certificate"`
			Key         string `yaml:"key"`
		} `yaml:"ssl"`
		Gui      string `yaml:"gui"`
		Test     string `yaml:"test"`
		Template string `yaml:"template"`
	} `yaml:"server"`

	Logger struct {
		Connection string `yaml:"connection"`
		Server     string `yaml:"server"`
	} `yaml:"logger"`

	Database []struct {
		ProjectId string `yaml:"projectId"`
		Key       string `yaml:"key"`
	} `yaml:"database"`

	Password struct {
		AccessToken  string `yaml:"accessToken"`
		RefreshToken string `yaml:"refreshToken"`
		Cookies      string `yaml:"cookies"`
		Mail         string `yaml:"mail"`
	} `yaml:"password"`

	Mail struct {
		Address    string `yaml:"address"`
		SmtpServer struct {
			Host string `yaml:"host"`
			Port string `yaml:"port"`
		} `yaml:"smtpServer"`
	} `yaml:"mail"`

	Template struct {
		Mail  MailTemplate  `yaml:"mail"`
		Music MusicTemplate `yaml:"music"`
		Html  struct {
			Redirect string `yaml:"redirect"`
		} `yaml:"html"`
	} `yaml:"template"`

	Jwt struct {
		Iss     string `yaml:"iss"`
		Company string `yaml:"company"`
	} `yaml:"jwt"`
}

type MusicTemplate struct {
	Path   string            `yaml:"path"`
	Fields map[string]string `yaml:"fields"`
}

type MailTemplate struct {
	Path   string                       `yaml:"path"`
	Fields map[string]MailTemplateField `yaml:"fields"`
}

type MailTemplateField struct {
	File   string `yaml:"file"`
	Object string `yaml:"object"`
}

func NewConfig() (*Config, error) {

	provider, err := config.NewYAML(
		config.File(configFolder+"server.yml"),
		config.File(configFolder+"logger.yml"),
		config.File(configFolder+"database.yml"),
		config.File(configFolder+"password.yml"),
		config.File(configFolder+"mail.yml"),
		config.File(configFolder+"template.yml"),
		config.File(configFolder+"jwt.yml"),
	)

	if err != nil {
		panic(err) // handle error
	}

	var c Config
	if err := provider.Get("").Populate(&c); err != nil {
		panic(err) // handle error
	}

	return &c, nil
}
