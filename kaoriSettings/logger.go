package kaoriSettings

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

//LoggerConfig è una struttura con le impostazioni del logger.
type LoggerConfig struct {
	Connection string `yaml:"connection" json:"connection,omitempty"`
	Server     string `yaml:"server" json:"server,omitempty"`
}

//WriteConf Scrive i settaggi del logger nel relativo file.
func (logConf *LoggerConfig) WriteConf() error {

	type Write struct {
		Header *LoggerConfig `yaml:"logger"`
	}

	w := Write{Header: logConf}

	data, err := yaml.Marshal(w)
	if err != nil {
		return errors.New("Error to marshal logger configuration: "+err.Error())
	}

	err = os.WriteFile("config/logger.tmp.yml", data, 0644)
	if err != nil {
		return errors.New("Error to write logger configuration file: "+err.Error())
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