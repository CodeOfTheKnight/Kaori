package kaoriSettings

//DatabaseConfig è una struttura con le impostazioni del database.
type DatabaseConfig struct {
	Relational []DBRelational `yaml:"relational" json:"relational"`
	NonRelational []DBNonRealtional `yaml:"nonRelational" json:"nonRelational"`
}

type DBRelational struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Host string `yaml:"host" json:"host"`
	Port string `yaml:"port" json:"port"`
	Db string `yaml:"db" json:"db"`
}

type DBNonRealtional struct {
	ProjectId string `yaml:"projectId" json:"projectId,omitempty"`
	Key       string `yaml:"key" json:"key,omitempty"`
}


//TODO: Modify checkDatabaseProjectId
/*

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

*/
