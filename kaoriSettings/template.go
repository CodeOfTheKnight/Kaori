package kaoriSettings

import (
	"errors"
	"gopkg.in/yaml.v2"
	"os"
)

//TemplateConfig è una struttura con le impostazioni dei template.
type TemplateConfig struct {
	Mail  MailTemplateConfig  `yaml:"mail" json:"mail,omitempty"`
	Music MusicTemplateConfig `yaml:"music" json:"music,omitempty"`
	Html  HTMLTemplateConfig `yaml:"html" json:"html,omitempty"`
}

//MusicTemplateConfig è una struttura con le impostazioni per i template relativi alla musica.
type MusicTemplateConfig map[string]string

//MusicTemplateConfig è una struttura con le impostazioni per i template relativi alle mail.
type MailTemplateConfig map[string]MailTemplateField

//MailTemplateField è una struttura con le impostazioni per ogni template della mail.
type MailTemplateField struct {
	File   string `yaml:"file" json:"file,omitempty"`
	Object string `yaml:"object" json:"object,omitempty"`
}

//HTMLTemplateConfig è una struttura con le impostazioni dei template html.
type HTMLTemplateConfig map[string]string


//WriteConf Scrive i settaggi dei template nel relativo file.
func (tempConf *TemplateConfig) WriteConf() error {

	type Write struct {
		Header *TemplateConfig `yaml:"template"`
	}

	w := Write{Header: tempConf}

	data, err := yaml.Marshal(w)
	if err != nil {
		return errors.New("Error to marshal template configuration: "+err.Error())
	}

	err = os.WriteFile("config/template.tmp.yml", data, 0644)
	if err != nil {
		return errors.New("Error to write template configuration file: "+err.Error())
	}

	return nil
}

