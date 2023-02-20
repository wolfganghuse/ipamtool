package main

import (
	"strconv"
	"reflect"
    "os"
	"fmt"
)
// string to bool.
func stob(s string) bool {
    i, err := strconv.ParseBool(s)
    if err != nil {
        panic(err)
    }
    return i
}

func setConfig() (err error) {
    // ValueOf returns a Value representing the run-time data
    v := reflect.ValueOf(serviceConfig)
    for i := 0; i < v.NumField(); i++ {
        // Get the field tag value
        tag := v.Type().Field(i).Tag.Get(ENV)
        defaultTag := v.Type().Field(i).Tag.Get(DEFAULT)

        // Skip if tag is not defined or ignored
        if tag == "" || tag == "-" {
            continue
        }
        //a := reflect.Indirect(reflect.ValueOf(serviceConfig))
        EnvVar, Info := loadFromEnv(tag, defaultTag)
        if Info != "" {
            if  v.Type().Field(i).Tag.Get(DEFAULT)!="" {
                fmt.Println("Missing environment configuration for '" + v.Type().Field(i).Name + "', please set ENV Variable " +  v.Type().Field(i).Tag.Get(ENV) + ". Using default value " +  v.Type().Field(i).Tag.Get(DEFAULT))
            } else {
                fmt.Println("Missing environment configuration for '" + v.Type().Field(i).Name + "', please set ENV Variable " +  v.Type().Field(i).Tag.Get(ENV))
                err=fmt.Errorf("not all ENV variables set")
            }
        }
        /* Set the value in the environment variable to the respective struct field */
        reflect.ValueOf(&serviceConfig).Elem().Field(i).SetString(EnvVar)

    }
    return err
}

func loadFromEnv(tag string, defaultTag string) (string, string) {
    /* Check if the tag is defined in the environment or else replace with default value */
    envVar := os.Getenv(tag)
    if envVar == "" {
        envVar = defaultTag
        /* '1' is used to indicate that default value is being loaded */
        return envVar, "1"
    }
    return envVar, ""
}

/*GetConfiguration :Exported function to return a copy of the configuration instance */
func GetConfiguration() Configuration {
    return serviceConfig
}