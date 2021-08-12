package xfile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestReadPropertiesAsMap(t *testing.T) {

	wd, _ := os.Getwd()
	path := wd + "/application-test.properties"

	kvs, err := ReadPropertiesAsMap(path)

	fmt.Println("Error: ", err)
	fmt.Println(kvs)

}

func TestReadYamlAsMap(t *testing.T) {

	wd, _ := os.Getwd()
	path := wd + "/application-test.yml"

	kvs, err := ReadYamlAsMap(path)

	fmt.Println("Error: ", err)
	fmt.Println(kvs)

}

func TestDirTest(t *testing.T) {
	wd, _ := os.Getwd()
	fmt.Println(filepath.Abs(wd + "/../../xver"))
}
