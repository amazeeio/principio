//go:build generate

package main

//go:generate ${CONTROLLER_GEN} object:headerFile=./hack/boilerplate.go.txt paths="./..."
//go:generate ${CONTROLLER_GEN} ${CRD_OPTIONS} rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

// Run this file itself
//go:generate go run generate.go config/crd/bases/init.amazee.io_initconfigs.yaml

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// controller-gen 0.3 creates CRDs with apiextensions.k8s.io/v1beta1, but some generated properties aren't valid for that version
// So we have to patch the CRD in post-generation.
// See https://github.com/kubernetes/kubernetes/issues/91395
func main() {
	workdir, _ := os.Getwd()
	log.Println("Running post-generate in " + workdir)
	file := os.Args[1]
	patchFile(file)
}

func patchFile(fileName string) {
	log.Println(fmt.Sprintf("Reading file %s", fileName))
	lines, err := readLines(fileName)
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	var result []string
	result = patchV1(lines, result)

	log.Println(fmt.Sprintf("Writing new file to %s", fileName))
	if err := writeLines(result, fileName); err != nil {
		log.Fatalf("writeLines: %s", err)
	}
}

func patchV1(lines []string, result []string) []string {
	for i, line := range lines {
		switch line {
		case "                  type: object":
			result = append(result, line)
			affectsSyncItems := strings.Contains(lines[i-4], "description: InitConfig")

			if affectsSyncItems {
				hasEmbeddedResource := strings.Contains(lines[i+2], "x-kubernetes-embedded-resource")
				if hasEmbeddedResource {
					result = append(result, "                  x-kubernetes-embedded-resource: true")
					log.Println(fmt.Sprintf("Added  'x-kubernetes-embedded-resource' after line %d", i))
				}
				hasPreserveUnknownFields := strings.Contains(lines[i+3], "x-kubernetes-preserve-unknown-fields")
				if hasPreserveUnknownFields {
					result = append(result, "                  x-kubernetes-preserve-unknown-fields: true")
					log.Println(fmt.Sprintf("Added  'x-kubernetes-preserve-unknown-fields' after line %d", i))
				}
			}
		case "                x-kubernetes-embedded-resource: true":
			log.Println(fmt.Sprintf("Removed 'x-kubernetes-embedded-resource' in line %d", i))
		case "                x-kubernetes-preserve-unknown-fields: true":
			log.Println(fmt.Sprintf("Removed 'x-kubernetes-preserve-unknown-fields' in line %d", i))
		default:
			result = append(result, line)
		}
	}
	return result
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
