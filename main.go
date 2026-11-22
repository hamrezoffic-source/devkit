package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const usage = `devkit - a fast, portable developer toolkit

Usage:
  devkit json format [file|-]
  devkit json to-yaml [file|-]
  devkit yaml to-json [file|-]
  devkit port check <host> <port>

Aliases:
  devkit format [file|-]
  devkit json-to-yaml [file|-]
  devkit yaml-to-json [file|-]
  devkit check-port <host> <port>

Use "-" or omit the file argument to read from standard input.
`

func readInput(args []string) ([]byte, error) {
	if len(args) == 0 || args[0] == "-" {
		return io.ReadAll(os.Stdin)
	}
	if len(args) > 1 {
		return nil, errors.New("too many arguments")
	}
	return os.ReadFile(args[0])
}

func jsonDecoder(data []byte) (*json.Decoder, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder, nil
}

func FormatJSON(data []byte) ([]byte, error) {
	decoder, _ := jsonDecoder(data)
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, errors.New("invalid JSON: multiple JSON values")
		}
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	output, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(output, '\n'), nil
}

func JSONToYAML(data []byte) ([]byte, error) {
	decoder, _ := jsonDecoder(data)
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, errors.New("invalid JSON: multiple JSON values")
		}
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	output, err := yaml.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("convert JSON to YAML: %w", err)
	}
	return output, nil
}

func yamlNodeToJSONCompatible(node *yaml.Node) (any, error) {
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		return yamlNodeToJSONCompatible(node.Content[0])
	case yaml.AliasNode:
		return yamlNodeToJSONCompatible(node.Alias)
	case yaml.ScalarNode:
		if node.Tag == "!!null" || node.Value == "" && node.Tag == "" {
			return nil, nil
		}
		switch node.Tag {
		case "!!bool":
			var value bool
			if err := node.Decode(&value); err != nil {
				return nil, err
			}
			return value, nil
		case "!!int":
			text := strings.ReplaceAll(node.Value, "_", "")
			if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
				value, err := strconv.ParseUint(text[2:], 16, 64)
				if err == nil {
					return json.Number(strconv.FormatUint(value, 10)), nil
				}
			}
			if strings.HasPrefix(text, "0o") || strings.HasPrefix(text, "0O") {
				value, err := strconv.ParseUint(text[2:], 8, 64)
				if err == nil {
					return json.Number(strconv.FormatUint(value, 10)), nil
				}
			}
			if strings.HasPrefix(text, "0b") || strings.HasPrefix(text, "0B") {
				value, err := strconv.ParseUint(text[2:], 2, 64)
				if err == nil {
					return json.Number(strconv.FormatUint(value, 10)), nil
				}
			}
			if strings.HasPrefix(text, "-") {
				value, err := strconv.ParseInt(text, 10, 64)
				if err == nil {
					return json.Number(strconv.FormatInt(value, 10)), nil
				}
			}
			value, err := strconv.ParseUint(text, 10, 64)
			if err == nil {
				return json.Number(strconv.FormatUint(value, 10)), nil
			}
			return node.Value, nil
		case "!!float":
			text := strings.ReplaceAll(node.Value, "_", "")
			if text == ".inf" || text == "+.inf" || text == "-.inf" ||
				text == ".nan" || text == "+.nan" || text == "-.nan" {
				return nil, fmt.Errorf("YAML value %q is not representable in JSON", node.Value)
			}
			var value float64
			if err := node.Decode(&value); err != nil {
				return nil, err
			}
			return value, nil
		case "!!timestamp":
			var value time.Time
			if err := node.Decode(&value); err == nil {
				return value.Format(time.RFC3339), nil
			}
			fallthrough
		default:
			return node.Value, nil
		}
	case yaml.SequenceNode:
		value := make([]any, 0, len(node.Content))
		for _, child := range node.Content {
			item, err := yamlNodeToJSONCompatible(child)
			if err != nil {
				return nil, err
			}
			value = append(value, item)
		}
		return value, nil
	case yaml.MappingNode:
		value := make(map[string]any, len(node.Content)/2)
		for i := 0; i < len(node.Content); i += 2 {
			keyNode, child := node.Content[i], node.Content[i+1]
			key, err := yamlNodeToJSONCompatible(keyNode)
			if err != nil {
				return nil, err
			}
			text, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("YAML mapping key %q is not a string", keyNode.Value)
			}
			if _, exists := value[text]; exists {
				return nil, fmt.Errorf("duplicate YAML mapping key %q", text)
			}
			item, err := yamlNodeToJSONCompatible(child)
			if err != nil {
				return nil, err
			}
			value[text] = item
		}
		return value, nil
	default:
		return nil, fmt.Errorf("unsupported YAML node kind")
	}
}

func YAMLToJSON(data []byte) ([]byte, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var documents []any
	for {
		var node yaml.Node
		err := decoder.Decode(&node)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid YAML: %w", err)
		}
		if len(node.Content) == 0 {
			continue
		}
		value, err := yamlNodeToJSONCompatible(&node)
		if err != nil {
			return nil, err
		}
		documents = append(documents, value)
	}
	if len(documents) == 0 {
		return []byte("null\n"), nil
	}
	var value any = documents[0]
	if len(documents) > 1 {
		value = documents
	}
	output, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("convert YAML to JSON: %w", err)
	}
	return append(output, '\n'), nil
}

func CheckTCPPort(host string, port int, timeout time.Duration) error {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	return conn.Close()
}

func writeOutput(data []byte) error {
	_, err := os.Stdout.Write(data)
	return err
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(usage)
		return nil
	}

	switch args[0] {
	case "-h", "--help", "help":
		fmt.Print(usage)
		return nil
	case "json":
		if len(args) < 2 {
			return errors.New("json command requires format, to-yaml, or tojson")
		}
		switch args[1] {
		case "format":
			data, err := readInput(args[2:])
			if err != nil {
				return err
			}
			output, err := FormatJSON(data)
			if err != nil {
				return err
			}
			return writeOutput(output)
		case "to-yaml", "toyaml", "2yaml":
			data, err := readInput(args[2:])
			if err != nil {
				return err
			}
			output, err := JSONToYAML(data)
			if err != nil {
				return err
			}
			return writeOutput(output)
		default:
			return fmt.Errorf("unknown json command %q", args[1])
		}
	case "yaml":
		if len(args) < 2 {
			return errors.New("yaml command requires to-json or tojson")
		}
		switch args[1] {
		case "to-json", "tojson", "2json":
			data, err := readInput(args[2:])
			if err != nil {
				return err
			}
			output, err := YAMLToJSON(data)
			if err != nil {
				return err
			}
			return writeOutput(output)
		default:
			return fmt.Errorf("unknown yaml command %q", args[1])
		}
	case "port":
		if len(args) < 2 || args[1] != "check" {
			return errors.New("port command requires check")
		}
		return runPortCheck(args[2:])
	case "format":
		data, err := readInput(args[1:])
		if err != nil {
			return err
		}
		output, err := FormatJSON(data)
		if err != nil {
			return err
		}
		return writeOutput(output)
	case "json-to-yaml", "jsontoyaml", "json2yaml":
		data, err := readInput(args[1:])
		if err != nil {
			return err
		}
		output, err := JSONToYAML(data)
		if err != nil {
			return err
		}
		return writeOutput(output)
	case "yaml-to-json", "yamltojson", "yaml2json":
		data, err := readInput(args[1:])
		if err != nil {
			return err
		}
		output, err := YAMLToJSON(data)
		if err != nil {
			return err
		}
		return writeOutput(output)
	case "check-port", "portcheck", "check":
		return runPortCheck(args[1:])
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func runPortCheck(args []string) error {
	timeoutSeconds := 3.0
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--timeout" || arg == "-timeout":
			if i+1 >= len(args) {
				return errors.New("timeout requires a value")
			}
			i++
			value, err := strconv.ParseFloat(args[i], 64)
			if err != nil || value <= 0 {
				return errors.New("timeout must be a positive number")
			}
			timeoutSeconds = value
		case strings.HasPrefix(arg, "--timeout=") || strings.HasPrefix(arg, "-timeout="):
			text := arg[strings.IndexByte(arg, '=')+1:]
			value, err := strconv.ParseFloat(text, 64)
			if err != nil || value <= 0 {
				return errors.New("timeout must be a positive number")
			}
			timeoutSeconds = value
		default:
			remaining = append(remaining, arg)
		}
	}
	if len(remaining) != 2 {
		return errors.New("usage: devkit port check <host> <port>")
	}
	port, err := strconv.Atoi(remaining[1])
	if err != nil || port < 1 || port > 65535 {
		return errors.New("port must be an integer between 1 and 65535")
	}
	timeout := time.Duration(timeoutSeconds * float64(time.Second))
	if err := CheckTCPPort(remaining[0], port, timeout); err != nil {
		fmt.Fprintf(os.Stdout, "closed: %s\n", err)
		return err
	}
	fmt.Fprintln(os.Stdout, "open")
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
