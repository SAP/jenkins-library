package npm

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	fileName = ".npmrc"
)

//     //registry.npmjs.org/:_authToken=${NPM_TOKEN}
func (rc NPMRC) SetAuth(registryUrl, token string) error {
	//TODO: handle urls correctly, cut of protocoll
	rc.Set(fmt.Sprintf("//%s:_authToken", registryUrl), token)
	return nil
}

func (rc NPMRC) Set(key, value string) {
	if rc.values == nil {
		rc.values = make(map[string]string)
	}
	rc.values[key] = value
}

func NewNPMRC() NPMRC {
	return NPMRC{values: make(map[string]string)}
}

type NPMRC struct {
	values map[string]string
}

func (rc NPMRC) Create() error {
	// err := ioutil.WriteFile(fileName, []byte(""), 0644)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func (rc NPMRC) content() string {
	var lines []string

	for key, value := range rc.values {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(lines[:], "\n")
}

func (rc NPMRC) Write() error {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	_, err = file.WriteString("\n" + rc.content() + "\n")
	if err != nil {
		return err
	}
	return nil
}

func (rc NPMRC) Read() (string, error) {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
