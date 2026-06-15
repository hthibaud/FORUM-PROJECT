package openenv

import (
	"Forum/pkg/utils"
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var ENV envData

func loadENV(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("impossible d'ouvrir le fichier: %w", err)
	}
	defer file.Close()

	envMap := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			val = strings.Trim(val, `"'`)
			envMap[key] = val
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("erreur de lecture: %w", err)
	}

	v := reflect.ValueOf(&ENV).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		valStr, exists := envMap[tag]
		if !exists {
		}

		switch field.Type.Kind() {
		case reflect.String:
			v.Field(i).SetString(valStr)
		case reflect.Int:
			if valInt, err := strconv.Atoi(valStr); err == nil {
				v.Field(i).SetInt(int64(valInt))
			}
		case reflect.Bool:
			if valBool, err := strconv.ParseBool(valStr); err == nil {
				v.Field(i).SetBool(valBool)
			}
		}
	}

	return nil
}

func Init() {
	utils.Debug("Init env variables...")
	err := loadENV(".env")
	if err != nil {
		utils.LogError("Loading env error", err)
		os.Exit(1)
	}
	utils.Debug("Env variables succefully init")
}
