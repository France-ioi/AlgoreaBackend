package cmd

import "os"

func resolveEnvironment(args []string, defaultEnvironment string) string {
	environment := defaultEnvironment
	if env, ok := os.LookupEnv("ALGOREA_ENV"); ok {
		environment = env
	}
	if len(args) > 0 {
		environment = args[0]
	}
	return environment
}
