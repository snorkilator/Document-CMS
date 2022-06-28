package readConfig

import (
	"os"

	"gopkg.in/yaml.v2"
)

type CFG struct {
	Host struct {
		IP   string
		Port string
		Path string
	}
	Db struct {
		Name     string
		Port     string
		Password string
		User     string
		Dbtype   string
	}
}

//GetConfig unmarshalls yaml config file stored in current directory and returns config as a struct.
func GetConfig() (CFG, error) {
	b, err := os.ReadFile("./serverConfig.yaml")
	if err != nil {
		panic(err)
	}
	var i CFG
	err = yaml.Unmarshal(b, &i)
	if err != nil {
		return CFG{}, err
	}
	return i, nil
}
