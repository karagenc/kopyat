package main

import (
	"fmt"
	"os"

	"github.com/kardianos/service"
)

func main() {
	config := &service.Config{
		Name:        "kopyashipd",
		DisplayName: "Kopyaship daemon",
	}
	v := &svice{}
	s, err := service.New(v, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	log, err := s.Logger(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	err = s.Run()
	if err != nil {
		log.Error(err)
	}
}
