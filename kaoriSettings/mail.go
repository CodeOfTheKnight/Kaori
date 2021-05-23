package kaoriSettings

import (
	"errors"
	"fmt"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

//MailConfig è una struttura con le impostazioni della mail.
type MailConfig struct {
	Address    string `yaml:"address" json:"address,omitempty"`
	SmtpServer SMTPServerConfig `yaml:"smtpServer" json:"smtpServer,omitempty"`
}

//SMTPServerConfig è una struttura con le impostazioni per il server mail SMTP.
type SMTPServerConfig struct {
	Host string `yaml:"host" json:"host,omitempty"`
	Port string `yaml:"port" json:"port,omitempty"`
}

//WriteConf Scrive i settaggi delle mail nel relativo file.
func (mailConf *MailConfig) WriteConf() error {

	type Write struct {
		Header *MailConfig `yaml:"mail"`
	}

	w := Write{Header: mailConf}

	data, err := yaml.Marshal(w)
	if err != nil {
		return errors.New("Error to marshal mail configuration: "+err.Error())
	}

	err = os.WriteFile("config/mail.tmp.yml", data, 0644)
	if err != nil {
		return errors.New("Error to write mail configuration file: "+err.Error())
	}

	return nil
}


//CheckMail controlla se tutte le impostazioni relative alle mail sono corrette.
func (m *MailConfig) CheckMail() error {

	//Check mail address
	if err := m.CheckMailAddress(); err != nil {
		return err
	}

	//Check mail smtp server
	if err := m.SmtpServer.CheckMailSMTP(); err != nil {
		return err
	}

	return nil
}

//CheckMailAddress controlla che l'indirizzo della mail impostato nei file di configurazione sia valido.
func (m *MailConfig) CheckMailAddress() error {
	if m.Address == "" || strings.Contains(m.Address, "@") || len(strings.Trim(m.Address, "@")) > 3 {
		return errors.New("Email address not valid.")
	}
	return nil
}

//CheckMailSMTP controlla che tutte le impostazioni del server smtp siano corrette.
func (sm *SMTPServerConfig) CheckMailSMTP() error {

	//Check host
	if err := sm.CheckMailSMTPHost(); err != nil {
		return err
	}

	//Check port
	if err := sm.CheckMailSMTPPort(); err != nil {
		return err
	}

	return nil
}

//CheckMailSMTPHost Controlla che l'impostazione "host" del server SMTP sia corretta.
func (sm *SMTPServerConfig) CheckMailSMTPHost() error {
	if sm.Host == "" {
		return errors.New("Hostname server SMTP not valid")
	}
	return nil
}

//CheckMailSMTPPort Controlla che l'impostazione "port" del server SMTP sia corretta.
func (sm *SMTPServerConfig) CheckMailSMTPPort() error {
	if err := kaoriUtils.PortValid(sm.Port); err != nil {
		return err
	}
	return nil
}

//CheckTemplate controlla tutte le impostazioni di tutti i template.
func (tmpl *TemplateConfig) CheckTemplate() error {

	//Check mail template
	if err := tmpl.Mail.CheckMailTemplate(); err != nil {
		return err
	}

	//Check music template
	if err := tmpl.Music.CheckMusicTemplate(); err != nil {
		return err
	}

	//Check html template
	if err := tmpl.Html.CheckHTMLTemplate(); err != nil {
		return err
	}

	return nil
}

//CheckMailTemplate controlla tutte le impostazioni dei mail template
func (tmplm MailTemplateConfig) CheckMailTemplate() error {

	//Check templates
	for _, field := range tmplm {
		if err := field.CheckAll(); err != nil {
			return err
		}
	}

	return nil
}

//CheckAll controlla tutte le impostazioni di un template mail.
func (fm *MailTemplateField) CheckAll() error {

	//Check file
	if err := fm.CheckFile(); err != nil {
		return err
	}

	//Check object
	if err := fm.CheckObject(); err != nil {
		return err
	}

	return nil
}

//CheckFile controlla l'impostazione "file" di un template mail.
func (fm *MailTemplateField) CheckFile() error {
	if _, err := os.Stat(fm.File); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`Template path "%s", not exist. Modify config file "server.yml!"`, fm.File))
	}
	return nil
}

//CheckObject controlla l'oggetto della mail del template mail.
func (fm *MailTemplateField) CheckObject() error {
	if fm.Object == "" {
		return errors.New("Object of template mail not valid.")
	}
	return nil
}

//CheckMusicTemplate controlla tutte le impostazioni dei music template.
func (tmplmu MusicTemplateConfig) CheckMusicTemplate() error {
	for _, field := range tmplmu {
		if _, err := os.Stat(field); os.IsNotExist(err) {
			return errors.New(fmt.Sprintf(`Template path "%s", not exist."`, field))
		}
	}
	return nil
}

//CheckHTMLTemplate controlla tutte le impostazioni dei template html.
func (tmplh HTMLTemplateConfig) CheckHTMLTemplate() error {
	for _, field := range tmplh {
		if _, err := os.Stat(field); os.IsNotExist(err) {
			return errors.New(fmt.Sprintf(`Template path "%s", not exist."`, field))
		}
	}
	return nil
}