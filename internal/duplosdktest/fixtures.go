package duplosdktest

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"strings"
)

type fixtureCache map[string][]byte

type FixtureWriter func()

var (
	fc   = fixtureCache{}
	fdir string
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fdir = path.Join(dir, "../internal/duplosdktest/fixtures")
}

func ResetFixtures() {
	fc = fixtureCache{}
}

func PatchFixture(location string, target interface{}, writer FixtureWriter) {
	body := GetFixture(location)
	err := json.Unmarshal(body, target)
	if err != nil {
		log.Panicf("json.Unmarshal: %s: %s", location, err)
	}

	writer()

	PostFixture(location, target)
}

func PostFixture(location string, source interface{}) {
	body, err := json.Marshal(source)
	if err != nil {
		log.Panicf("json.Marshal: %s: %s", location, err)
	}

	log.Printf("[TRACE] %s: (over)wrote new fixture", location)
	fc[location] = body

	// Invalidate any cached list
	delete(fc, path.Dir(location))
}

func DeleteFixture(location string) bool {
	file := path.Join(fdir, location) + ".json"

	if _, err := os.Stat(file); err == nil {
		log.Panicf("fixtureDelete: %s: cannot delete a permanent fixture", location)
	}

	_, ok := fc[location]
	delete(fc, location)

	// Invalidate any cached list
	delete(fc, path.Dir(location))
	return ok
}

func ListFixtures(location string) []byte {
	// Return the data if it is cached
	if buff, ok := fc[location]; ok {
		return buff
	}

	dir := path.Join(fdir, location)

	// List files
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Panicf("readdir: %s: %s", location, err)
	}

	// Cache all non-present elements.
	for i := range files {
		name := files[i].Name()
		if strings.HasSuffix(name, ".json") {
			GetFixture(path.Join(location, strings.TrimSuffix(name, ".json")))
		}
	}

	// Collect all elements, and calculate our final size.
	prefix := location + "/"
	size := 2 // 2 brackets
	locations := make([]string, 0, len(files)*2)
	for key, buff := range fc {

		// Only matching keys.
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		name := strings.TrimPrefix(key, prefix)

		// Only direct descendants.
		if strings.Contains(name, "/") {
			continue
		}

		if size != 2 {
			size += 1 // leading comma
		}
		size += len(buff)
		locations = append(locations, key)
	}

	// Fill the buffer
	buff := make([]byte, 0, size)
	buff = append(buff, []byte("[")...)
	for i := range locations {
		if i > 0 {
			buff = append(buff, []byte(",")...)
		}
		buff = append(buff, fc[locations[i]]...)
	}
	buff = append(buff, []byte("]")...)

	// cache the result and return it
	fc[location] = buff
	return buff
}

// nolint
func GetFixture(location string) []byte {
	if buff, ok := fc[location]; ok {
		return buff
	}

	path := path.Join(fdir, location) + ".json"
	buff, err := os.ReadFile(path)
	if err != nil {
		log.Panicf("readfile: %s: %s", path, err)
	}

	// cache the result and return it
	fc[location] = buff
	return buff
}

func GetResource(location string, target interface{}) error {
	bytes := GetFixture(location)
	return json.Unmarshal(bytes, target)
}
