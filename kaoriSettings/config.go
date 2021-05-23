package kaoriSettings

import (
	"fmt"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	"go.uber.org/config"
	"os"
	"strings"
)

//configFolder è una costante che definisce la path dove risiedono i file di configurazione.
const configFolder string = "../config/"

type error interface {
	Error() string
}

//Config è una struttura con le impostazioni del software.
type Config struct {
	Server ServerConfig `yaml:"server" json:"server,omitempty"`
	Logger LoggerConfig `yaml:"logger" json:"logger,omitempty"`
	Database DatabaseConfig `yaml:"database" json:"database,omitempty"`
	Password PasswordConfig `yaml:"password" json:"password,omitempty"`
	Mail MailConfig `yaml:"mail" json:"mail,omitempty"`
	Template TemplateConfig `yaml:"template" json:"template,omitempty"`
	Jwt JWTConfig `yaml:"jwt" json:"jwt,omitempty"`
}

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
		return nil, err
	}

	var c Config
	if err := provider.Get("").Populate(&c); err != nil {
		panic(err) // handle error
	}

	return &c, nil
}

//WriteConfig Scrive tutti i settaggi nei relativi file.
func (conf *Config) WriteConfig() (wdone []string, err error) {

	//Write server configuration
	if err = conf.Server.WriteConf(); err != nil {
		return
	}

	wdone = append(wdone, `[DONE] Write server configuration in the file.`)

	//Write logger configuration
	if err = conf.Logger.WriteConf(); err != nil {
		return
	}

	wdone = append(wdone, `[DONE] Write logger configuration in the file.`)

	//TODO: Add
	//Write database configuration
	/*if err = conf.WriteDatabaseConf(); err != nil {
		return nil, err
	}*/

	wdone = append(wdone, `[DONE] Write database configuration in the file.`)

	//Write password configuration
	if err = conf.Password.WriteConf(); err != nil {
		return
	}

	wdone = append(wdone, `[DONE] Write password configuration in the file.`)

	//Write password configuration
	if err = conf.Mail.WriteConf(); err != nil {
		return
	}

	wdone = append(wdone, `[DONE] Write mail configuration in the file.`)

	//Write template configuration
	if err = conf.Template.WriteConf(); err != nil {
		return
	}

	wdone = append(wdone, `[DONE] Write template configuration in the file.`)

	//Write JWT configuration
	if err = conf.Jwt.WriteConf(); err != nil {
		return
	}

	wdone = append(wdone, `[DONE] Write JWT configuration in the file.`)

	return
}

/*
//TODO: Modify database configuration
//WriteDatabaseConf Scrive i settaggi dei database nel relativo file.
func (conf *Config) WriteDatabaseConf() error {

	type Write struct {
		Header DatabaseConfig `yaml:"database"`
	}

	w := Write{}
	w.Header = conf.Database

	data, err := yaml.Marshal(w)
	if err != nil {
		return errors.New("Error to marshal database configuration: "+err.Error())
	}

	err = os.WriteFile("config/database.tmp.yml", data, 0644)
	if err != nil {
		return errors.New("Error to write database configuration file: "+err.Error())
	}

	return nil
}

*/


//TODO: Modify database config

/*
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
*/

func CheckPrecedentConfig() error {

	//Get files name
	files, err := kaoriUtils.Ls(configFolder)
	if err != nil {
		return err
	}

	fmt.Println(files)

	for _, file := range files {
		if strings.Contains(file, ".tmp") {
			realName := strings.Replace(file, ".tmp", "", -1)

			//Remove old file
			err = os.Remove(realName)
			if err != nil {
				return err
			}

			//Rename new file
			err = os.Rename(file, realName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

