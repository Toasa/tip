package main

import (
	"fmt"
	"os"
)

func main() {
	filename := "test.zip"
	if err := createZip(filename); err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}
}

func createZip(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// End of central directory record
	date := []byte {
		0x06, 0x05, 0x4b, 0x50,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
		0x00,
	}

	_, err = file.Write(date)
	if err != nil {
		return err
	}
	return nil
}