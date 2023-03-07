# GoGPT

GoGPT is a program written in Go that allows you to chat with the GPT chatbot in your terminal.

![GoGPT](screenshot.png)

## Installation

To install GoGPT, follow these steps:
1. Install Go on your system.
2. Clone this repository.
3. Run `go build` to build the program.
4. Run the executable file to start chatting with ChatGPT.

## Usage

To chat with ChatGPT, run the executable file and follow the prompts. You can optionally pass in the `-k` flag to specify API Key to use. Here's an example:
```
./GoGPT -k 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
```

## Configuration

```
  -k string
        Your API Key
  -m string
        The Model to chat with (default "gpt-3.5-turbo")
  -p string
        Proxy address, eg http://127.0.0.1:7890 or sock5://127.0.0.1:7890
```

## Contributing

If you'd like to contribute to GoGPT, please submit a pull request with your changes. Before submitting, make sure your code passes the tests and is properly formatted.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
