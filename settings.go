package main

import (
	"errors"
	"fmt"
	"go.uber.org/config"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//configFolder è una costante che definisce la path dove risiedono i file di configurazione.
const configFolder string = "config/"

//Config è una struttura con le impostazioni del software.
type Config struct {
	Server ServerConfig `yaml:"server" json:"server,omitempty"`
	Logger LoggerConfig `yaml:"logger" json:"logger,omitempty"`
	Database []DatabaseConfig `yaml:"database" json:"database,omitempty"`
	Password PasswordConfig `yaml:"password" json:"password,omitempty"`
	Mail MailConfig `yaml:"mail" json:"mail,omitempty"`
	Template TemplateConfig `yaml:"template" json:"template,omitempty"`
	Jwt JWTConfig `yaml:"jwt" json:"jwt,omitempty"`
}

//ServerConfig è una struttura con le impostazioni del server.
type ServerConfig struct {
	Host string `yaml:"host" json:"host,omitempty"`
	Port string `yaml:"port" json:"port,omitempty"`
	Ssl  SSLConfig `yaml:"ssl" json:"ssl,omitempty"`
	Limiter int `yaml:"limiter" json:"limiter,omitempty"`
	Gui      string `yaml:"gui" json:"gui,omitempty"`
	Test     string `yaml:"test" json:"test,omitempty"`
	Template string `yaml:"template" json:"template,omitempty"`
}

//LoggerConfig è una struttura con le impostazioni del logger.
type LoggerConfig struct {
	Connection string `yaml:"connection" json:"connection,omitempty"`
	Server     string `yaml:"server" json:"server,omitempty"`
}

//DatabaseConfig è una struttura con le impostazioni del database.
type DatabaseConfig struct {
	ProjectId string `yaml:"projectId" json:"projectId,omitempty"`
	Key       string `yaml:"key" json:"key,omitempty"`
}

//PasswordConfig è una struttura dove sono impostate le password del software.
type PasswordConfig struct {
	AccessToken  string `yaml:"accessToken" json:"accessToken,omitempty"`
	RefreshToken string `yaml:"refreshToken" json:"refreshToken,omitempty"`
	Cookies      string `yaml:"cookies" json:"cookies,omitempty"`
	Mail         string `yaml:"mail" json:"mail,omitempty"`
}

//MailConfig è una struttura con le impostazioni della mail.
type MailConfig struct {
	Address    string `yaml:"address" json:"address,omitempty"`
	SmtpServer SMTPServerConfig `yaml:"smtpServer" json:"smtpServer,omitempty"`
}

//TemplateConfig è una struttura con le impostazioni dei template.
type TemplateConfig struct {
	Mail  MailTemplateConfig  `yaml:"mail" json:"mail,omitempty"`
	Music MusicTemplateConfig `yaml:"music" json:"music,omitempty"`
	Html  HTMLTemplateConfig `yaml:"html" json:"html,omitempty"`
}

//JWTConfig è una struttura con le impostazioni del JWT.
type JWTConfig struct {
	Iss     string `yaml:"iss" json:"iss,omitempty"`
	Company string `yaml:"company" json:"company,omitempty"`
	Exp JWTExpConfig `yaml:"exp" json:"exp,omitempty"`
}

//JWTExpConfig è una struttura con le impostazioni di scadenza dei JWT.
type JWTExpConfig struct {
	AccessToken string `yaml:"accessToken" json:"accessToken,omitempty"`
	RefreshToken string `yaml:"refreshToken" json:"refreshToken,omitempty"`
}

//SSLConfig è una struttura con le impostazioni dei certificati SSL.
type SSLConfig struct {
	Certificate string `yaml:"certificate" json:"certificate,omitempty"`
	Key         string `yaml:"key" json:"key,omitempty"`
}

//SMTPServerConfig è una struttura con le impostazioni per il server mail SMTP.
type SMTPServerConfig struct {
	Host string `yaml:"host" json:"host,omitempty"`
	Port string `yaml:"port" json:"port,omitempty"`
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

//NewConfig è un costruttore dell'oggetto config.
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

//CheckConfig controlla la validità di tutte le configurazioni.
func (cfg *Config) CheckConfig() error {

	//Check server settings
	if err := cfg.Server.CheckServer(); err != nil {
		return err
	}

	//Check logger settings
	if err := cfg.Logger.CheckLogger(); err != nil {
		return err
	}

	//Check databases settings
	for _, db := range cfg.Database {
		if err := db.CheckDatabase(); err != nil {
			return err
		}
	}

	//Check password settings
	if err := cfg.Password.CheckPassword(); err != nil {
		return err
	}

	//Check template settings
	if err := cfg.Template.CheckTemplate(); err != nil {
		return err
	}

	return nil
}

//CheckServer controlla la validità di tutte le impostazioni relative al server.
func (srv *ServerConfig) CheckServer() error {

	//Check host validity
	if err := srv.CheckHost(); err != nil {
		return err
	}

	//Check port validity and disponibility
	if err := srv.CheckPort(); err != nil {
		return err
	}

	//Check Limiter validity
	if err := srv.CheckLimiter(); err != nil {
		return err
	}

	//Check SSL config
	if err := cfg.Server.Ssl.CheckSSL(); err != nil {
		return err
	}

	//Check template path
	if err := srv.CheckTemplate(); err != nil {
		return err
	}

	//Check gui path
	if err := srv.CheckGui(); err != nil {
		return err
	}

	//Check test path
	if err := srv.CheckTest(); err != nil {
		return err
	}

	return nil
}

//CheckHost controlla l'impostazione host del server.
func (srv *ServerConfig) CheckHost() error {
	if srv.Host == "" {
		return errors.New("Server host not valid.")
	}
	return nil
}

//CheckPort controlla l'impostazione porta del server.
func (srv *ServerConfig) CheckPort() error {
	portInt, err := strconv.Atoi(strings.Trim(srv.Port, ":"))
	if err != nil {
		return errors.New("Invalid Port: Conversion of port to int not valid")
	}

	if portInt < 1024 || portInt > 49151 {
		return errors.New("Port not valid. [1024-49151]")
	}
	return nil
}

//CheckTemplate controlla la path template del server.
func (srv *ServerConfig) CheckTemplate() error {
	if _, err := os.Stat(srv.Template); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`Template path "%s", not exist. Modify config file "server.yml!"`, srv.Template))
	}
	return nil
}

//CheckGui controlla la path gui del server.
func (srv *ServerConfig) CheckGui() error {
	if _, err := os.Stat(srv.Gui); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`Gui path "%s", not exist. Modify config file "server.yml!"`, srv.Gui))
	}
	return nil
}

//CheckTest controlla la path test del server
func (srv *ServerConfig) CheckTest() error {
	if _, err := os.Stat(cfg.Server.Test); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`Test path "%s", not exist. Modify config file "server.yml!"`, cfg.Server.Test))
	}
	return nil
}

//CheckSSL controlla tutte le impostazioni SSL del server.
func (ssl *SSLConfig) CheckSSL() error {

	//Check key SSL
	if _, err := os.Stat(ssl.Key); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`"%s": Not exist! Modify config file "server.yml!"`, ssl.Key))
	}

	//Check certificate SSL
	if _, err := os.Stat(ssl.Certificate); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`"%s": Not exist! Modify config file "server.yml!"`, ssl.Certificate))
	}

	return nil
}

//CheckLimiter controlla tutte le impostazioni del limiter del server.
func (srv *ServerConfig) CheckLimiter() error {

	if srv.Limiter < 1 {
		return errors.New("Limiter not valid.")
	}

	return nil
}

//CheckLogger controlla se esistono tutte le path di log impostate nel file di configurazione.
func (lo *LoggerConfig) CheckLogger() error {

	//Check Server
	if err := lo.CheckLoggerServer(); err != nil {
		return err
	}

	//Check Connection
	if err := lo.CheckLoggerConnection(); err != nil {
		return err
	}

	return nil
}

//CheckLoggerServer controlla se esiste la path di log del server impostata nel file di configurazione.
func (lo *LoggerConfig) CheckLoggerServer() error {

	//Check Server
	dirname := filepath.Dir(lo.Server)
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`Server logger path "%s", not exist. Modify config file "server.yml!"`, lo.Server))
	}

	return nil
}

//CheckLoggerConnection controlla se esiste la path di log delle connessioni impostata nel file di configurazione.
func (lo *LoggerConfig) CheckLoggerConnection() error {
	dirname := filepath.Dir(lo.Connection)
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`Connection logger path "%s", not exist. Modify config file "server.yml!"`, lo.Connection))
	}

	return nil
}

//CheckDatabase controlla che tutte le impostazioni del database siano corrette.
func (db *DatabaseConfig) CheckDatabase() error {

	//Check projectId
	if err := db.CheckDatabaseProjectId(); err != nil {
		return err
	}

	//Check Key
	if err := db.CheckDatabaseKey(); err != nil {
		return err
	}

	return nil
}

//CheckDatabase controlla che projectId sia corretto.
func (db *DatabaseConfig) CheckDatabaseProjectId() error {
	if db.ProjectId == "" {
		return errors.New("ProjectId not valid.")
	}
	return nil
}

//CheckDatabase controlla che esista il file che contiene la key del database.
func (db *DatabaseConfig) CheckDatabaseKey() error {
	if _, err := os.Stat(db.Key); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`Key database file path "%s", not exist. Modify config file "server.yml!"`, db.Key))
	}
	return nil
}

//CheckPassword controlla se tutte le password nel database sono corrette.
func (psw *PasswordConfig) CheckPassword() error {

	//Check cookies password
	if err := passwordValid(psw.Cookies); err != nil {
		return fmt.Errorf("Cookie " + err.Error())
	}

	//Check mail password
	if psw.Mail == "" {
		return errors.New("Mail password not valid.")
	}

	//Check Refresh Token password
	if err := passwordValid(psw.RefreshToken); err != nil {
		return fmt.Errorf("Refresh Token " + err.Error())
	}

	//Check Access Token password
	if err := passwordValid(psw.AccessToken); err != nil {
		return fmt.Errorf("Access Token " + err.Error())
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
	if err := portValid(sm.Port); err != nil {
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