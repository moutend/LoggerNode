package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		return
	}

fmt.Println(http.MethodPost, os.Args[1])

	res, err := http.Post(os.Args[1], "application/json", bytes.NewBufferString(`{
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
