package common

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"

	"gonum.org/v1/gonum/stat/combin"
)

type NonZeroExitError struct {
	ExitCode int
}

func (err NonZeroExitError) Error() string {
	return "Non-zero exit code: " + strconv.FormatInt(int64(err.ExitCode), 10)
}

func PropertyMapToInterface(propertyMap map[string]interface{}, typ reflect.Type) interface{} {
	objValue := reflect.New(typ)
	objValueElem := objValue.Elem()
	for key, value := range propertyMap {
		field := objValueElem.FieldByName(key)
		SetFieldValue(field, value)
	}
	return objValue.Interface()
}

func SetFieldValue(field reflect.Value, value interface{}) {
	reflectValue := reflect.ValueOf(value)
	reflectValueType := reflectValue.Type().Kind().String()
	switch reflectValueType {
	case "slice":
		convertedSliceValue := reflect.MakeSlice(field.Type(), 0, reflectValue.Len())
		for _, sliceItem := range value.([]interface{}) {
			convertedSliceValue = reflect.Append(convertedSliceValue, reflect.ValueOf(sliceItem))
		}
		field.Set(convertedSliceValue)
	case "map":
		convertedMapValue := reflect.MakeMap(field.Type())
		for key, value := range value.(map[string]interface{}) {
			convertedMapValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		}
		field.Set(convertedMapValue)
	default:
		field.Set(reflectValue.Convert(field.Type()))
	}
}

func CpR(sourcePath string, targetFolderPath string) {
	sourceFileInfo, err := os.Stat(sourcePath)
	if err != nil {
		// Return if source path doesn't exist so we can use this with optional files
		return
	}
	// Handle if source is file
	if !sourceFileInfo.IsDir() {
		Cp(sourcePath, targetFolderPath)
		return
	}

	// Otherwise create the directory and scan contents
	toFolderPath := filepath.Join(targetFolderPath, sourceFileInfo.Name())
	os.MkdirAll(toFolderPath, sourceFileInfo.Mode())
	fileInfos, err := ioutil.ReadDir(sourcePath)
	if err != nil {
		log.Fatal(err)
	}
	for _, fileInfo := range fileInfos {
		fromPath := filepath.Join(sourcePath, fileInfo.Name())
		if fileInfo.IsDir() {
			CpR(fromPath, toFolderPath)
		} else {
			Cp(fromPath, toFolderPath)
		}
	}
}

func Cp(sourceFilePath string, targetFolderPath string) {
	sourceFileInfo, err := os.Stat(sourceFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Make target file
	targetFilePath := filepath.Join(targetFolderPath, sourceFileInfo.Name())
	targetFile, err := os.Create(targetFilePath)
	if err != nil {
		log.Fatal(err)
	}
	// Sync source and target file mode and ownership
	targetFile.Chmod(sourceFileInfo.Mode())
	SyncUIDGID(targetFile, sourceFileInfo)

	// Execute copy
	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		log.Fatal(err)
	}
	targetFile.Close()
	sourceFile.Close()
}

func WriteFile(filename string, fileBytes []byte) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if _, err = file.Write(fileBytes); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return nil
}

func GetAllCombinations(arraySize int) [][]int {
	combinationList := [][]int{}
	for i := 1; i <= arraySize; i++ {
		combinationList = append(combinationList, combin.Combinations(arraySize, i)...)
	}
	return combinationList
}

func AddToSliceIfUnique(slice []string, value string) []string {
	if SliceContainsString(slice, value) {
		return slice
	}
	return append(slice, value)
}

func SliceContainsString(slice []string, item string) bool {
	for _, sliceItem := range slice {
		if sliceItem == item {
			return true
		}
	}
	return false
}

func SliceUnion(slice1 []string, slice2 []string) []string {
	union := slice1[0:]
	for _, sliceItem := range slice2 {
		union = AddToSliceIfUnique(union, sliceItem)
	}
	return union
}

func ToJsonRawMessage(value interface{}) json.RawMessage {
	var err error
	var bytes json.RawMessage
	if bytes, err = json.Marshal(value); err != nil {
		log.Fatal("Failed to convert to json.RawMessage:", err)
	}
	return bytes
}

func DockerCli(workingDir string, captureOutput bool, args ...string) []byte {
	var outputBytes bytes.Buffer
	var errorOutput bytes.Buffer

	dockerCommand := exec.Command("docker", args...)
	dockerCommand.Env = os.Environ()
	if captureOutput {
		dockerCommand.Stdout = &outputBytes
		dockerCommand.Stderr = &errorOutput
	} else {
		writer := log.Writer()
		dockerCommand.Stdout = writer
		dockerCommand.Stderr = writer
	}
	if workingDir != "" {
		dockerCommand.Dir = workingDir
	}
	commandErr := dockerCommand.Run()
	if commandErr != nil || dockerCommand.ProcessState.ExitCode() != 0 || errorOutput.Len() != 0 {
		log.Fatal("Docker command failed: " + errorOutput.String() + commandErr.Error())
	}
	return outputBytes.Bytes()
}

func UntarBytes(tarBytes []byte, targetFolder string) error {
	return Untar(bytes.NewReader(tarBytes), targetFolder)
}

func Untar(reader io.Reader, targetFolder string) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		// Iterate through each entry in the file
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		targetPath := filepath.Join(targetFolder, header.Name)

		headerFileInfo := header.FileInfo()
		// If header says entry is a folder, create it
		if headerFileInfo.IsDir() {
			if _, err := os.Stat(targetPath); err != nil {
				if err := os.MkdirAll(targetPath, headerFileInfo.Mode()); err != nil {
					return err
				}
			}
		} else {
			// Otherwise create a file
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, headerFileInfo.Mode())
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err = io.Copy(file, tarReader); err != nil {
				return err
			}
		}
	}
	return nil
}
