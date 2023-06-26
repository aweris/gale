package main

import "os"

func main() {
	srv := NewLocalService(os.TempDir())

	if err := Serve("8080", srv); err != nil {
		panic(err)
	}
}
