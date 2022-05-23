package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/joho/godotenv"
	"gonum.org/v1/gonum/stat/combin"
)

type NonZeroExitError struct {
	ExitCode int
}

type LinuxDistroInfo struct {
	Name         string
	PrettyName   string
	VersionId    string
	Version      string
	Id           string
	IdLike       string
	HomeUrl      string
	SupportUrl   string
	BugReportUrl string
}

var cachedLinuxDistroInfo LinuxDistroInfo = LinuxDistroInfo{}

func ReadLinuxDistroInfo() LinuxDistroInfo {
	if cachedLinuxDistroInfo.Name == "" {
		if _, err := os.Stat("/etc/os-release"); err != nil {
			log.Fatal("/etc/os-release not found.")
		}
		osRelease, err := godotenv.Read("/etc/os-release")
		if err != nil {
			log.Fatal("Unable to read /etc/os-release.")
		}
		cachedLinuxDistroInfo = LinuxDistroInfo{
			Name:         osRelease["NAME"],
			PrettyName:   osRelease["PRETTY_NAME"],
			VersionId:    osRelease["VERSION_ID"],
			Version:      osRelease["VERSION"],
			Id:           osRelease["ID"],
			IdLike:       osRelease["ID_LIKE"],
			HomeUrl:      osRelease["HOME_URL"],
			SupportUrl:   osRelease["SUPPORT_URL"],
			BugReportUrl: osRelease["BUG_REPORT_URL"],
		}
	}
	return cachedLinuxDistroInfo
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

func MergeProperties(existingVal interface{}, inVal interface{}) interface{} {
	if existingVal == nil {
		return inVal
	}

	if inVal == nil {
		return existingVal
	}

	typ := reflect.TypeOf(inVal).Kind()
	existingTyp := reflect.TypeOf(existingVal).Kind()
	if typ != existingTyp {
		log.Fatal("Failed to merge properties due to type mismatch. Existing: ", existingTyp, ", input:", typ)
	}
	if typ == reflect.Slice || typ == reflect.Array {
		outVal := make([]interface{}, 0)
		rExVal := reflect.ValueOf(existingVal)
		for i := 0; i < rExVal.Len(); i++ {
			outVal = append(outVal, rExVal.Index(i).Interface())
		}
		rInVal := reflect.ValueOf(inVal)
		for i := 0; i < rInVal.Len(); i++ {
			outVal = append(outVal, rInVal.Index(i).Interface())
		}
		return outVal

	} else if typ == reflect.Map {
		outVal := make(map[string]interface{})
		rExVal := reflect.ValueOf(existingVal)
		rInVal := reflect.ValueOf(inVal)
		exItr := rExVal.MapRange()
		for exItr.Next() {
			outVal[exItr.Key().String()] = exItr.Value().Interface()
		}
		inItr := rInVal.MapRange()
		for inItr.Next() {
			rExistingMapVal := rExVal.MapIndex(inItr.Key())
			if rExistingMapVal.Kind() == reflect.Invalid {
				outVal[inItr.Key().String()] = MergeProperties(nil, inItr.Value().Interface())
			} else {
				outVal[inItr.Key().String()] = MergeProperties(rExistingMapVal.Interface(), inItr.Value().Interface())
			}
		}
		return outVal
	}
	return inVal
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

func InterfaceToStringSlice(inter interface{}) []string {
	value := reflect.ValueOf(inter)
	newSlice := make([]string, value.Len())
	for i := 0; i < value.Len(); i++ {
		newSlice[i] = fmt.Sprint(value.Index(i))
	}
	return newSlice
}

func ToJsonRawMessage(value interface{}) json.RawMessage {
	var err error
	var bytes json.RawMessage
	if bytes, err = json.Marshal(value); err != nil {
		log.Fatal("Failed to convert to json.RawMessage:", err)
	}
	return bytes
}

func ExecCmd(workingDir string, captureOutput bool, command string, args ...string) []byte {
	var outputBytes bytes.Buffer
	var errorOutput bytes.Buffer

	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()
	if captureOutput {
		cmd.Stdout = &outputBytes
		cmd.Stderr = &errorOutput
	} else {
		writer := log.Writer()
		cmd.Stdout = writer
		cmd.Stderr = writer
	}
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	if err := cmd.Run(); err != nil {
		log.Fatal("Command", command, fmt.Sprint(args), " failed. ", err)
	} else if cmd.ProcessState.ExitCode() != 0 {
		log.Fatal("Command", command, fmt.Sprint(args), " failed with exit code ", cmd.ProcessState.ExitCode())
		if captureOutput {
			log.Fatal("Command output:", outputBytes.String())
		}
	}
	return outputBytes.Bytes()
}

func UntarBytes(tarBytes []byte, destination string, strip int) {
	Untar(bytes.NewReader(tarBytes), destination, strip)
}

func Untar(reader io.Reader, destination string, strip int) {
	var err error
	if destination, err = filepath.Abs(destination); err != nil {
		log.Fatal("Failed to convert path to absolute path. ", err)
	}

	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		log.Fatal("Unable to create gzip reader. ", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		// Iterate through each entry in the file
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal("Error reading tar file. ", err)
		}

		// Strip out specified number of folders from target path
		typeflag := rune(header.Typeflag)
		targetRelPath := header.Name
		if strip > 0 {
			tarFilePathParts := strings.Split(header.Name, string(os.PathSeparator))
			if len(tarFilePathParts) <= strip {
				continue
			}
			basePath := filepath.Join(tarFilePathParts[:strip]...)
			targetRelPath, _ = filepath.Rel(basePath, targetRelPath)
		}
		linkRelPath := header.Linkname
		if linkRelPath != "" {
			linkRelPath = filepath.Join(filepath.Dir(targetRelPath), header.Linkname)
		}

		// Convert to absolute paths
		var targetPath, linkPath string
		if targetPath, err = filepath.Abs(filepath.Join(destination, targetRelPath)); err != nil {
			log.Fatal("Failed to convert path to absolute path. ", err)
		}
		if !strings.HasPrefix(targetPath, destination) {
			continue
		}
		if typeflag == tar.TypeLink || typeflag == tar.TypeSymlink {
			if linkPath, err = filepath.Abs(filepath.Join(destination, linkRelPath)); err != nil {
				log.Fatal("Failed to convert path to absolute path. ", err)
			}
			if !strings.HasPrefix(linkPath, destination) {
				continue
			}
			log.Println("Typeflag:", header.Typeflag, "targetRelPath:", targetRelPath, "linkRelPath:", linkRelPath)
		}

		// Process contents
		switch typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(targetPath); err != nil {
				if err := os.MkdirAll(targetPath, fs.FileMode(header.Mode)); err != nil {
					log.Fatal("Failed to create directory. ", err)
				}
			}
		case tar.TypeSymlink:
			if err := os.Symlink(linkPath, targetPath); err != nil {
				log.Fatal("Failed to create symlink. ", err)
			}
		case tar.TypeLink:
			if err := os.Link(linkPath, targetPath); err != nil {
				log.Fatal("Failed to create link. ", err)
			}
		case tar.TypeReg:
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.FileMode(header.Mode))
			if err != nil {
				log.Fatal("Failed to open file. ", err)
			}
			defer file.Close()
			if _, err = io.Copy(file, tarReader); err != nil {
				log.Fatal("Failed to copy file. ", err)
			}
		}
	}
	return
}

func NewSemverRange(version string) semver.Range {
	// Convert node shorthands to semver.Range string
	requestedVersion := strings.ReplaceAll(version, "*", "x")
	// 18.1.2 - 18.3.2 is >=18.1.2 <=18.3.2
	exp := regexp.MustCompile(`[0-9x]+(\.[^ ]+)? - [0-9x]+`)
	rangeLocs := exp.FindAllStringIndex(requestedVersion, -1)
	if rangeLocs != nil {
		for _, loc := range rangeLocs {
			requestedVersion = requestedVersion[:loc[0]] + ">=" + requestedVersion[loc[0]:]
		}
		requestedVersion = strings.ReplaceAll(requestedVersion, " - ", " <=")
	}
	// Handle ^ and ~
	hasCarrotOrTilde, _ := regexp.MatchString("(^|~)", requestedVersion)
	if hasCarrotOrTilde {
		semverRange := ""
		rangeParts := strings.Split(requestedVersion, " ")
		for _, part := range rangeParts {
			if part[0] == '~' {
				// ~18.1.2 is >=18.1.2 <18.2.0
				semverRange += ">=" + part[1:]
				tempVersion, err := semver.ParseTolerant(part[1:])
				if err != nil {
					log.Fatal(err)
				}
				tempVersion.IncrementMinor()
				tempVersion.Patch = 0
				semverRange += " <" + tempVersion.FinalizeVersion() + " "
			} else if part[0] == '^' {
				// ^18.1.2 is >=18.1.2 <19.0.0
				semverRange += ">=" + part[1:] + " "
				tempVersion, err := semver.ParseTolerant(part[1:])
				if err != nil {
					log.Fatal(err)
				}
				tempVersion.IncrementMajor()
				tempVersion.Minor = 0
				tempVersion.Patch = 0
				semverRange += " <" + tempVersion.FinalizeVersion() + " "
			} else {
				semverRange += part + " "
			}
		}
		requestedVersion = semverRange
	}
	// 18 is 18.x.x, 18.1 is 18.1.x
	expX := regexp.MustCompile(`[!=>< ][0-9x]+(\.[0-9x]+)? `)
	requestedVersion = " " + requestedVersion + " "
	for {
		loc := expX.FindStringIndex(requestedVersion)
		if loc == nil {
			break
		}
		version := requestedVersion[loc[0]+1 : loc[1]-1]
		if strings.Contains(version, ".") {
			version = version + ".x"
		} else {
			version = version + ".x.x"
		}
		requestedVersion = requestedVersion[:loc[0]+1] + version + requestedVersion[loc[1]-1:]
	}

	return semver.MustParseRange(requestedVersion)
}

func DownloadBytesFromUrl(dlUrl string) []byte {
	response, err := http.Get(dlUrl)
	if err != nil {
		log.Fatal(err)
	}
	if response.StatusCode != 200 {
		log.Fatal("Got status code ", response.StatusCode, " for ", dlUrl)
	}
	defer response.Body.Close()
	outBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Failed to read response body. ", err)
	}
	return outBytes
}
