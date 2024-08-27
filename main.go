package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/denormal/go-gitignore"
	"github.com/yargevad/filepathx"

	"github.com/num30/config"
	"github.com/pkoukk/tiktoken-go"
)

var ignore gitignore.GitIgnore

type Config struct {
	Model           string `flag:"model" envvar:"MODEL" default:"gpt-4"`
	OutputFile      string `flag:"output" envvar:"FILE" default:"file-for-ai.txt"`
	IgnoreGitIgnore bool   `flag:"ignore-gitignore" envvar:"IGNORE_GITIGNORE" default:"false"`
	ProcessNonText  bool   `flag:"process-non-text" envvar:"PROCESS_NON_TEXT" default:"false"`
}

var conf Config

func main() {

	err := config.NewConfReader("file-for-ai").Read(&conf)
	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Error: Directory path or glob pattern is required.")
		fmt.Println("Usage: file-for-ai <directory|pattern> --output [output file]")
		fmt.Println("\nExamples:")
		fmt.Println("  file-for-ai /path/to/directory")
		fmt.Println("  file-for-ai './*.txt'")
		fmt.Println("  file-for-ai /path/to/directory --output custom-output.txt")
		fmt.Println("\nFor more information, visit https://github.com/num30/file-for-ai?tab=readme-ov-file#file-for-ai")
		os.Exit(1)
	}

	inputPath := os.Args[1]

	outputFileName := conf.OutputFile

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	encoding := conf.Model

	tkm, err := tiktoken.EncodingForModel(encoding)
	if err != nil {
		fmt.Println("Error getting encoding for model:", err)
		os.Exit(1)
	}
	tokens := 0

	fmt.Println("Merging files:")

	if isDirectory(inputPath) {

		if !conf.IgnoreGitIgnore {
			fmt.Print("Filtering files using .gitignore... ")
			ignore, err = gitignore.NewRepository(inputPath)
			if err != nil {
				fmt.Println("Error creating gitignore repository:", err)
				os.Exit(1)
			}
		}

		err = filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println("Error accessing path:", path, err)
				return err
			}
			return processFile(inputPath, path, info, outputFile, tkm, &tokens, outputFileName)
		})
	} else {
		files, err := filepathx.Glob(inputPath)
		if err != nil {
			fmt.Println("Error parsing glob pattern:", err)
			os.Exit(1)
		}

		for _, path := range files {
			info, err := os.Stat(path)
			if err != nil {
				fmt.Println("Error accessing file:", path, err)
				continue
			}
			err = processFile(filepath.Dir(path), path, info, outputFile, tkm, &tokens, outputFileName)
			if err != nil {
				fmt.Println("Error processing file:", path, err)
				continue
			}
		}
	}

	fmt.Println()
	if err != nil {
		fmt.Println("Error walking through the directory or processing pattern:", err)
		return
	}

	fmt.Printf("Files merged successfully into %s\n", outputFileName)
	fmt.Printf("Total tokens for model %s: %s\n", conf.Model, formatIntNumber(tokens))
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func processFile(basePath, path string, info os.FileInfo, outputFile *os.File, tkm *tiktoken.Tiktoken, tokens *int, outputFileName string) error {
	if info.Name() == outputFileName {
		return nil
	}

	relativePath, err := filepath.Rel(basePath, path)
	if err != nil {
		fmt.Println("Error processing relative path:", path, err)
		return err
	}

	if !info.IsDir() && !isGitIgnored(relativePath, info.IsDir()) && extensionBasedFilter(path) && !strings.HasPrefix(path, ".") {
		fileContents, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Error reading file:", path, err)
			return err
		}
		fmt.Println(relativePath)
		*tokens += len(tkm.Encode(string(fileContents), nil, nil))

		separator := fmt.Sprintf("\n\n>>>>>> %s <<<<<<\n\n", relativePath)
		if _, err := outputFile.WriteString(separator); err != nil {
			fmt.Println("Error writing separator to output file:", err)
			return err
		}

		if _, err := outputFile.Write(fileContents); err != nil {
			fmt.Println("Error writing file contents to output file:", err)
			return err
		}
	}

	return nil
}

func formatIntNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result string
	for i, c := range s {
		if i != 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

func isGitIgnored(path string, isDir bool) bool {
	if ignore == nil {
		return false
	}

	m := ignore.Relative(path, isDir)
	if m != nil {
		return m.Ignore()
	}
	return false
}

func extensionBasedFilter(path string) bool {
	if conf.ProcessNonText {
		return true
	}

	ext := strings.ToLower(filepath.Ext(path))
	return !nonTextFileExtensions[ext]
}

// List of common non-text file extensions based on MIME types and practical file handling.
var nonTextFileExtensions = map[string]bool{
	// Images
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".tiff": true, ".svg": true, ".psd": true, ".ai": true, ".webp": true, ".heic": true,

	// Audio
	".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true, ".m4a": true, ".wma": true,

	// Video
	".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".flv": true, ".wmv": true, ".m4v": true, ".mpg": true, ".mpeg": true, ".h264": true,

	// Compressed files
	".zip": true, ".rar": true, ".7z": true, ".gz": true, ".tar": true, ".bz2": true, ".xz": true, ".tgz": true, ".zipx": true, ".iso": true,

	// Executables and binaries
	".exe": true, ".bin": true, ".dll": true, ".so": true, ".rpm": true, ".deb": true, ".dmg": true, ".bat": true, ".jar": true,

	// Documents and PDFs
	".pdf": true, ".doc": true, ".docx": true, ".ppt": true, ".pptx": true, ".xls": true, ".xlsx": true, ".odt": true, ".ods": true,
	".odp": true, ".epub": true, ".mobi": true,

	// 3D Models
	".obj": true, ".stl": true, ".dae": true, ".blend": true,

	// Database files
	".sqlite": true, ".db": true, ".sql": true, ".mdb": true, ".accdb": true,

	// Code and Script binaries
	".pyc": true, ".class": true, ".o": true, ".a": true, ".dylib": true, ".lib": true,

	// Other formats
	".ps": true, ".eps": true, ".xps": true, ".swf": true, ".fla": true, // Adobe & Microsoft
	".indd": true,

	// Cryptographic files
	".pem": true, ".key": true, ".cert": true, ".crt": true,

	// Virtual Machine files
	".vmdk": true, ".ovf": true, ".vdi": true,

	// Game files
	".pak": true, ".bsp": true, ".wad": true,

	// Fonts
	".ttf": true, ".otf": true, ".woff": true, ".woff2": true,

	// Email
	".pst": true, ".eml": true, ".msg": true,

	// Disk Images
	".img": true, ".vhdx": true,

	// No extension (commonly used for binary files)
	"": true, // representing no-extension files as potentially non-text

	// Additional generic binary/data files
	".dat": true,

	// Specific project files
	".xd":     true, // Adobe XD
	".sketch": true, // Sketch App

	// CAD Files
	".dwg": true, ".dxf": true,
}
