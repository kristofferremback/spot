package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func ReadCache(cacheFileName string, output interface{}) error {
	logrus.Debugf("Attempting to read cache file %s", cacheFileName)

	if _, err := os.Stat(cacheFileName); err != nil {
		// Don't read the cache and ignore the error, cache file doesn't exist
		return nil
	}

	jsonBytes, err := ioutil.ReadFile(cacheFileName)
	if err != nil {
		return fmt.Errorf("Failed to read cache file %s: %v", cacheFileName, err)
	}

	if err := json.Unmarshal(jsonBytes, &output); err != nil {
		return fmt.Errorf("Failed to unmarshal cache file %s: %v", cacheFileName, err)
	}

	logrus.Infof("Successfully read cache file %s", cacheFileName)

	return nil
}

func WriteCache(fileName string, data interface{}) error {
	logrus.Debugf("Attempting to write to the cache file %s", fileName)

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Failed to marshal cache file %s, %v", fileName, err)
	}

	directory := filepath.Dir(fileName)

	if _, err = os.Stat(directory); err != nil {
		logrus.Infof("Creating directory %s to store the cache file in.", directory)

		if err = os.MkdirAll(directory, os.ModePerm); err != nil {
			return fmt.Errorf("Failed to create directory %s for cache file %s: %v", directory, fileName, err)
		}
	}

	if err = ioutil.WriteFile(fileName, jsonBytes, os.ModePerm); err != nil {
		return fmt.Errorf("Failed to write to cache file %s: %v", fileName, err)
	}

	logrus.Infof("Successfully wrote to the cache file %s", fileName)

	return nil
}
