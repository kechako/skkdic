package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kechako/skkdic"
)

const appName = "skkdic-expr"

func _main() (int, error) {
	var delimiter string
	var outputFile string
	var inputEncoding string
	var outputEncoding string
	flag.StringVar(&delimiter, "d", "", "")
	flag.StringVar(&outputFile, "o", "", "")
	flag.StringVar(&inputEncoding, "i", "", "")
	flag.StringVar(&outputEncoding, "e", "", "")
	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-i input_encoding] [-e output_encoding] [-d delimiter] [-o output] jisyo1 [[+-^] jisyo2]...\n", appName)
	}
	flag.Parse()

	var output io.Writer
	if outputFile == "" {
		output = os.Stdout
	} else {
		file, err := os.Create(outputFile)
		if err != nil {
			return 1, fmt.Errorf("failed to create output file: %s: %w", outputFile, err)
		}
		defer file.Close()
		output = file
	}

	var dicOpts []skkdic.Option
	if delimiter != "" {
		dicOpts = append(dicOpts, skkdic.WithAnnotationDelimiter(delimiter))
	}
	var readOpts []skkdic.ReadOption
	if inputEncoding != "" {
		readOpts = append(readOpts, skkdic.WithInputEncoding(skkdic.Encoding(inputEncoding)))
	}
	var writeOpts []skkdic.WriteOption
	if outputEncoding != "" {
		writeOpts = append(writeOpts, skkdic.WithOutputEncoding(skkdic.Encoding(outputEncoding)))
	}

	dic := skkdic.New(dicOpts...)

	if flag.NArg() == 0 {
		err := dic.Read(os.Stdin, skkdic.Add, readOpts...)
		if err != nil {
			return 1, fmt.Errorf("failed to read dictionary: %w", err)
		}

		err = dic.Write(output, writeOpts...)
		if err != nil {
			return 1, fmt.Errorf("failed to write dictionary: %w", err)
		}

		return 0, nil
	}

	args := flag.Args()
	mode := skkdic.Add
	for i := 0; i < len(args); i++ {
		switch args[i][0] {
		case '+':
			mode = skkdic.Add
			if len(args[i]) > 1 {
				err := dic.ReadFile(args[i][1:], mode, readOpts...)
				if err != nil {
					return 1, fmt.Errorf("failed to read dictionary: %w", err)
				}
			}
		case '-':
			mode = skkdic.Sub
			if len(args[i]) > 1 {
				err := dic.ReadFile(args[i][1:], mode, readOpts...)
				if err != nil {
					return 1, fmt.Errorf("failed to read dictionary: %w", err)
				}
				mode = skkdic.Add
			}
		case '^':
			mode = skkdic.And
			if len(args[i]) > 1 {
				err := dic.ReadFile(args[i][1:], mode, readOpts...)
				if err != nil {
					return 1, fmt.Errorf("failed to read dictionary: %w", err)
				}
				mode = skkdic.Add
			}
		default:
			err := dic.ReadFile(args[i], mode, readOpts...)
			if err != nil {
				return 1, fmt.Errorf("failed to read dictionary: %w", err)
			}
		}
	}

	err := dic.Write(output, writeOpts...)
	if err != nil {
		return 1, fmt.Errorf("failed to write dictionary: %w", err)
	}

	return 0, nil
}
func main() {
	code, err := _main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error : %v\n", err)
	}
	if code != 0 {
		os.Exit(code)
	}
}
