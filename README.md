# File-for-AI

File-for-AI is an open-source tool designed to compile all text files in a directory into a single file. This is particularly useful for preparing data to be added to AI models or chat GPT.

## Description

This tool traverses through a specified directory, reads all the text files (while ignoring non-text files and files specified in .gitignore), and compiles them into a single output file. This output file can then be used for various purposes such as training AI models or chat GPT.


## Usage

Download the latest binary for the os of your coice from the [releases page](https://github.com/num30/file-for-ai/releases) it's under `Assets` section.

To use this tool, you need to provide at least a directory path as an argument. If an output file name is not provided, it will default to "file-for-ai.txt".

```bash
file-for-ai <directory> 
```

For example:

```bash
file-for-ai /path/to/your/directory
```

This will create a file named "file-for-ai.txt" in your current directory with the contents of all text files in the specified directory.

## Flags
The tool accepts the following flags:

- `--model` or `-m`: Specifies the model to use for toke counting. It should be one of the available models in [tiktoken-go](https://github.com/pkoukk/tiktoken-go?tab=readme-ov-file#available-encodings). Default is "gpt-4".
- `--output` or `-o`: Specifies the name of the output file. Default is "file-for-ai.txt".

For example:

``` bash
file-for-ai /path/to/your/directory --model gpt-4o --output my-output.txt
```


This will create a file named "my-output.txt" in your current directory with the contents of all text files in the specified directory, token counting using the "gpt-4" model.
