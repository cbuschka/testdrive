package config

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestLoadConfigJson(t *testing.T) {
	reader := strings.NewReader(`{"version":"v1",
					"services":{
					  "db":{"image":"db-image"},
					  "appserver":{"image":"appserver-image",
					    "depends_on": ["db"]
					  }
					},"tasks":{
					  "integrationtests":{"image":"integrationtests-image",
					    "depends_on": ["db"]
					  }
					}
				     }`)

	config, err := LoadConfig(reader)
	if err != nil {
		t.Fatal(err)
		return
	}

	assert.Equal(t, "v1", config.Version)
	assert.Equal(t, "db-image", config.Services["db"].Image)
	assert.Equal(t, "appserver-image", config.Services["appserver"].Image)
	assert.Equal(t, []string{"db"}, config.Services["appserver"].Dependencies)
	assert.Equal(t, "integrationtests-image", config.Tasks["integrationtests"].Image)
	assert.Equal(t, []string{"db"}, config.Tasks["integrationtests"].Dependencies)
}

func TestLoadConfigYaml(t *testing.T) {
	reader := strings.NewReader(
		`
version: v1
services:
  db:
    image: db-image

  appserver:
    image: appserver-image
    depends_on:
      - db
tasks:
  integrationtests:
    image: integrationtests-image
    depends_on:
      - db
`)

	config, err := LoadConfig(reader)
	if err != nil {
		t.Fatal(err)
		return
	}

	assert.Equal(t, "v1", config.Version)
	assert.Equal(t, "db-image", config.Services["db"].Image)
	assert.Equal(t, "appserver-image", config.Services["appserver"].Image)
	assert.Equal(t, []string{"db"}, config.Services["appserver"].Dependencies)
	assert.Equal(t, "integrationtests-image", config.Tasks["integrationtests"].Image)
	assert.Equal(t, []string{"db"}, config.Tasks["integrationtests"].Dependencies)
}
