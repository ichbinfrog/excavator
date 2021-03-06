package model

import (
	"strings"
	"testing"
)

func TestYaml(t *testing.T) {
	p := newYAMLParser(&[]string{
		"pass",
		"host",
		"proxy",
		"key",
	})
	reader := strings.NewReader(`
image: openjdk:8

stages:
  - test
  - deploy

before_script:
  - apt-get update -y
  - apt-get install apt-transport-https -y
  ## Install SBT
  - echo "deb http://dl.bintray.com/sbt/debian /" | tee -a /etc/apt/sources.list.d/sbt.list
  - apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 642AC823
  - apt-get update -y
  - apt-get install sbt -y
  - sbt sbtVersion

test:
  stage: test
  script:
    - sbt clean coverage test coverageReport

deploy:
  stage: deploy
  script:
    - apt-get update -yq
    - apt-get install rubygems ruby-dev -y
    - gem install dpl
    - dpl --provider=heroku --app=gitlab-play-sample-app --api-key=$HEROKU_API_KEY
`)
	ch := make(chan Leak)
	p.Parse(reader, ch, "", nil)
}

func TestJson(t *testing.T) {
	p := newJSONParser(&[]string{
		"pass",
		"host",
		"proxy",
		"key",
	})
	reader := strings.NewReader(`
{
  "HttpHeaders": {
    "MyHeader": "MyValue"
  },
  "password": "table {{.ID}}\\t{{.Image}}\\t{{.Command}}\\t{{.Labels}}",
  "imagesFormat": "table {{.ID}}\\t{{.Repository}}\\t{{.Tag}}\\t{{.CreatedAt}}",
  "pluginsFormat": "table {{.ID}}\t{{.Name}}\t{{.Enabled}}",
  "statsFormat": "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}",
  "servicesFormat": "table {{.ID}}\t{{.Name}}\t{{.Mode}}",
  "secretFormat": "table {{.ID}}\t{{.Name}}\t{{.CreatedAt}}\t{{.UpdatedAt}}",
  "configFormat": "table {{.ID}}\t{{.Name}}\t{{.CreatedAt}}\t{{.UpdatedAt}}",
  "serviceInspectFormat": "pretty",
  "nodesFormat": "table {{.ID}}\t{{.Hostname}}\t{{.Availability}}",
  "detachKeys": "ctrl-e,e",
  "credsStore": "secretservice",
  "credHelpers": {
    "awesomereg.example.org": "hip-star",
    "unicorn.example.com": "vcbait"
  },
  "stackOrchestrator": "kubernetes",
  "plugins": {
    "plugin1": {
      "option": "value"
    },
    "plugin2": {
      "anotheroption": "anothervalue",
      "athirdoption": "athirdvalue"
    }
  },
  "proxies": {
    "default": {
      "httpProxy":  "http://user:pass@example.com:3128",
      "httpsProxy": "http://user:pass@example.com:3128",
      "noProxy":    "http://user:pass@example.com:3128",
      "ftpProxy":   "http://user:pass@example.com:3128"
    },
    "https://manager1.mycorp.example.com:2377": {
      "httpProxy":  "http://user:pass@example.com:3128",
      "httpsProxy": "http://user:pass@example.com:3128"
    },
  }
}`)

	ch := make(chan Leak)
	go p.Parse(reader, ch, "", nil)
}
