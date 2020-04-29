package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
)

type APISpec struct {
	//APIVersion string `json:"apiVersion"`
	Kind     string `json:"kind"`
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
}

func main() {
	path := flag.String("path", ".", "path to search for duplicated files in")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "This tool will is used in our gitops environment. We have one folder for each flux deployment so we use the folder also when checking for uniqueness.")
		fmt.Fprintf(flag.CommandLine.Output(), "\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	list := []string{}
	filenames := make(map[string]string)

	err := filepath.Walk(*path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.HasSuffix(path, ".yml") {
				return nil
			}
			line, err := fetchFile(path)
			if err != nil {
				return err
			}
			v := fmt.Sprintf("%s-%s", filepath.Dir(path), line)
			list = append(list, v)
			filenames[v] += path + ","

			return nil
		})
	if err != nil {
		log.Fatal(err)
	}

	sort.Strings(list)

	last := ""
	foundDuplicated := false
	for _, v := range list {
		if v == last {
			fmt.Printf("we found duplicate %s in files %s\n", v, filenames[v])
			foundDuplicated = true
		}
		last = v
	}

	if foundDuplicated {
		os.Exit(1)
	}
}

func fetchFile(file string) (string, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	spec := &APISpec{}
	err = yaml.Unmarshal(content, spec)
	if err != nil {
		return "", err
	}

	if spec.Metadata.Namespace == "" {
		spec.Metadata.Namespace = "default"
	}

	return fmt.Sprintf("%s-%s-%s", spec.Metadata.Name, spec.Metadata.Namespace, spec.Kind), nil
}
