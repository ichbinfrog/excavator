# Excavator

[![Go Report Card](https://goreportcard.com/badge/ichbinfrog/excavator)](https://goreportcard.com/report/github.com/ichbinfrog/excavator)  [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/ichbinfrog/excavator/blob/master/LICENSE)

Excavator is a lightweight pure Golang git leak scanning tool based on [SAP's credential digger](https://github.com/SAP/credential-digger) (without machine learning models) which attempts to improve on performance by parallelising commit iteration.

## CLI usage
Download a binary ![here](https://github.com/ichbinfrog/excavator/releases).

```sh
excavator scan [flags]
```

### Flags

- `-h` , `--help` : display help
- `-c` , `--concurrent <int>` : number of concurrent executions (defaults to 1), any integer given below 0 is considered as a single routine execution
- `-p` , `--path <string>` : temporary local path to store the git repository (only applies to remote repository) (default *.*)
- `-r` , `--rules <string>` : location of the rule declaration (default */home/ichbinfrog/.excavator/rules.yaml*)
- `-f` , `--format <string>` : format of output result (default *html*) (currently supports `yaml`, `html`)

### Global Flags

- `-v` , `--verbosity <int>` : logging verbosity:
  - 0: Fatal 
  - 1: Error
  - 2: Warning 
  - 3: Info (default) 
  - 4: Debug 
  - 5: Trace

Scanning a repository without backend
```sh
excavator scan {repository}
```

## Include in code

```golang
import (
  "github.com/ichbinfrog/excavator/pkg/scan"
)

func main() {
  c := &scan.Scanner{}

  // Directory in which to store the cloned repository
  directory := ...
  
  // URL / local path of git repository
  // for private repositories the url can be set as
  // https://user:pass@host/repo.git
  repo := ...
  
  // path to rule file
  rule := ...

  // Number of concurrent go routines 
  concurrent := ...

  // Whether or not to show progress bar
  progressBar := ...

  // Output interface
  // Can be either
  //  - &YamlReport{}
  //  - &HTMLReport{}
  report := ...
  c.New(repo, directory, rule, report, progressBar)
}
```

## Declaring rules

```yaml
# rules.yaml
#
apiVersion: v1
rules:
  - # regex of rule
    definition: EAACEdEose0cBA[0-9A-Za-z]+
    # category of rule
    category: token
    # description (optional)
    description: facebook access token rule

# list of regexes of file to ignore
black_list:
  - '.*sample.*'
```
