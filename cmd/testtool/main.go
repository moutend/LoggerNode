package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func main() {
	res, err := http.Post("http://localhost:7901/v1/log", "application/json", bytes.NewBufferString(`{
  "messages": [
    {
      "level": "INFO",
      "source": "testtool",
      "version": "develop",
      "message": "Hello, World!",
      "threadId": 12345
    }
  ]
  }`))

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body := &bytes.Buffer{}

	if _, err := io.Copy(body, res.Body); err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", body.Bytes())
}
