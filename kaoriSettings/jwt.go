package kaoriSettings

import (
	"errors"
	"gopkg.in/yaml.v2"
	"os"
)

//JWTConfig è una struttura con le impostazioni del JWT.
type JWTConfig struct {
	Iss     string `yaml:"iss" json:"iss,omitempty"`
	Company string `yaml:"company" json:"company,omitempty"`
	Exp JWTExpConfig `yaml:"exp" json:"exp,omitempty"`
}

//JWTExpConfig è una struttura con le impostazioni di scadenza dei JWT.
type JWTExpConfig struct {
	AccessToken int `yaml:"accessToken" json:"accessToken,omitempty"`
	RefreshToken int `yaml:"refreshToken" json:"refreshToken,omitempty"`
}

//WriteConf Scrive i settaggi del JWT nel relativo file.
func (jwtConf *JWTConfig) WriteConf() error {

	type Write struct {
		Header *JWTConfig `yaml:"jwt"`
	}

	w := Write{Header: jwtConf}

	data, err := yaml.Marshal(w)
	if err != nil {
		return errors.New("Error to marshal JWT configuration: "+err.Error())
	}

	err = os.WriteFile("config/jwt.tmp.yml", data, 0644)
	if err != nil {
		return errors.New("Error to write JWT configuration file: "+err.Error())
	}

	return nil
}
