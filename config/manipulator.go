package config

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
)

func LoadConfig(filename string) (pc ProgramConfig, err error) {
	var data []byte

	data, err = ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	_, err = toml.Decode(string(data), &pc)

	return
}

func SaveConfig(filename string, pc ProgramConfig) error {
	b := &bytes.Buffer{}
	enc := toml.NewEncoder(b)

	err := enc.Encode(pc)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, b.Bytes(), os.ModePerm)
}
