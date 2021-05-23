package kaoriSettings

import (
	"errors"
	"fmt"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	"gopkg.in/yaml.v2"
	"os"
)

//PasswordConfig è una struttura dove sono impostate le password del software.
type PasswordConfig struct {
	AccessToken  string `yaml:"accessToken" json:"accessToken,omitempty"`
	RefreshToken string `yaml:"refreshToken" json:"refreshToken,omitempty"`
	Cookies      string `yaml:"cookies" json:"cookies,omitempty"`
	Mail         string `yaml:"mail" json:"mail,omitempty"`
}

//WriteConf Scrive i settaggi delle password nel relativo file.
func (passConf *PasswordConfig) WriteConf() error {

	type Write struct {
		Header *PasswordConfig `yaml:"password"`
	}

	w := Write{Header: passConf}

	data, err := yaml.Marshal(w)
	if err != nil {
		return errors.New("Error to marshal password configuration: "+err.Error())
	}

	err = os.WriteFile("config/password.tmp.yml", data, 0644)
	if err != nil {
		return errors.New("Error to write password configuration file: "+err.Error())
	}

	return nil
}

//CheckPassword controlla se tutte le password nel database sono corrette.
func (psw *PasswordConfig) CheckPassword() error {

	//Check cookies password
	if err := kaoriUtils.PasswordValid(psw.Cookies); err != nil {
		return fmt.Errorf("Cookie " + err.Error())
	}

	//Check mail password
	if psw.Mail == "" {
		return errors.New("Mail password not valid.")
	}

	//Check Refresh Token password
	if err := kaoriUtils.PasswordValid(psw.RefreshToken); err != nil {
		return fmt.Errorf("Refresh Token " + err.Error())
	}

	//Check Access Token password
	if err := kaoriUtils.PasswordValid(psw.AccessToken); err != nil {
		return fmt.Errorf("Access Token " + err.Error())
	}

	return nil
}
