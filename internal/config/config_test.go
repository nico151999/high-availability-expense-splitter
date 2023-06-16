package config

import (
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

type testConfig struct {
	Item testItem
}

type testItem struct {
	Foo string
	Bar int
}

func TestLoadConfig(t *testing.T) {
	testData := testConfig{
		Item: testItem{
			Foo: "Hello world",
			Bar: 5,
		},
	}
	const configFilePath = "/tmp/config_test.yaml"
	deleteTempFile := mustCreateYAMLFile(testData, configFilePath, t)
	defer deleteTempFile()
	cfg, err := LoadConfig[testConfig](configFilePath)
	if err != nil {
		t.Errorf("failed loading config: %+v", err)
		return
	}
	if !reflect.DeepEqual(*cfg, testData) {
		t.Errorf("the loaded config contains data different to the ones saved to the FS")
	}
}

func mustCreateYAMLFile[T any](data T, path string, t *testing.T) func() {
	var marshalled []byte
	{
		var err error
		marshalled, err = yaml.Marshal(data)
		if err != nil {
			t.Errorf("failed marshalling config: %+v", err)
		}
	}
	if err := os.WriteFile(path, marshalled, 0644); err != nil {
		t.Errorf("failed writing config file: %+v", err)
	}
	return func() {
		err := os.Remove(path)
		if err != nil {
			t.Errorf("failed cleaning up temporary config file")
		}
	}
}
