package cli

import (
	"fmt"
	"strconv"
)

type httpInitOptions struct {
	providerName         string
	presetName           string
	providerURL          string
	providerModel        string
	apiKeyEnv            string
	agent                string
	timeoutMS            int
	maxOutputTokens      int
	responseFormat       string
	inputCostPerMillion  float64
	outputCostPerMillion float64
}

func (opts *options) startHTTPProvider(name string) {
	opts.finishHTTPInit()
	opts.httpInit = httpInitOptions{providerName: name}
}

func (opts *options) finishHTTPInit() {
	if opts.httpInit.IsZero() {
		return
	}
	opts.httpInits = append(opts.httpInits, opts.httpInit)
	opts.httpInit = httpInitOptions{}
}

func (init httpInitOptions) IsZero() bool {
	return init.providerName == "" &&
		init.presetName == "" &&
		init.providerURL == "" &&
		init.providerModel == "" &&
		init.apiKeyEnv == "" &&
		init.agent == "" &&
		init.timeoutMS == 0 &&
		init.maxOutputTokens == 0 &&
		init.responseFormat == "" &&
		init.inputCostPerMillion == 0 &&
		init.outputCostPerMillion == 0
}

func parseHTTPInitFlag(args []string, index int, opts *options) (bool, int, error) {
	switch args[index] {
	case "--http-provider":
		value, err := parseNextValue(args, index, "--http-provider requires a name")
		if err != nil {
			return true, index, err
		}
		opts.startHTTPProvider(value)
	case "--http-preset":
		value, err := parseNextValue(args, index, "--http-preset requires a preset name")
		if err != nil {
			return true, index, err
		}
		opts.httpInit.presetName = value
	case "--http-url":
		value, err := parseNextValue(args, index, "--http-url requires a URL")
		if err != nil {
			return true, index, err
		}
		opts.httpInit.providerURL = value
	case "--http-model":
		value, err := parseNextValue(args, index, "--http-model requires a model")
		if err != nil {
			return true, index, err
		}
		opts.httpInit.providerModel = value
	case "--http-api-key-env":
		value, err := parseNextValue(args, index, "--http-api-key-env requires an env var name")
		if err != nil {
			return true, index, err
		}
		opts.httpInit.apiKeyEnv = value
	case "--http-agent":
		value, err := parseNextValue(args, index, "--http-agent requires an agent name")
		if err != nil {
			return true, index, err
		}
		opts.httpInit.agent = value
	case "--http-timeout-ms":
		value, err := parseNonNegativeIntFlag(args, index, "--http-timeout-ms")
		if err != nil {
			return true, index, err
		}
		opts.httpInit.timeoutMS = value
	case "--http-max-output-tokens":
		value, err := parseNonNegativeIntFlag(args, index, "--http-max-output-tokens")
		if err != nil {
			return true, index, err
		}
		opts.httpInit.maxOutputTokens = value
	case "--http-response-format":
		value, err := parseNextValue(args, index, "--http-response-format requires a value")
		if err != nil {
			return true, index, err
		}
		opts.httpInit.responseFormat = value
	case "--http-input-cost-per-million":
		value, err := parseNonNegativeFloatFlag(args, index, "--http-input-cost-per-million")
		if err != nil {
			return true, index, err
		}
		opts.httpInit.inputCostPerMillion = value
	case "--http-output-cost-per-million":
		value, err := parseNonNegativeFloatFlag(args, index, "--http-output-cost-per-million")
		if err != nil {
			return true, index, err
		}
		opts.httpInit.outputCostPerMillion = value
	default:
		return false, index, nil
	}
	return true, index + 1, nil
}

func parseNextValue(args []string, index int, message string) (string, error) {
	if index+1 >= len(args) {
		return "", fmt.Errorf("%s", message)
	}
	return args[index+1], nil
}

func parseNonNegativeIntFlag(args []string, index int, flag string) (int, error) {
	raw, err := parseNextValue(args, index, flag+" requires a number")
	if err != nil {
		return 0, err
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", flag)
	}
	return value, nil
}

func parseNonNegativeFloatFlag(args []string, index int, flag string) (float64, error) {
	raw, err := parseNextValue(args, index, flag+" requires a number")
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative number", flag)
	}
	return value, nil
}
