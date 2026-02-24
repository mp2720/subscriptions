package main

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	log     *log.Logger
	verbose bool
}

func NewLoger(verbose bool) *Logger {
	return &Logger{
		log:     log.New(os.Stderr, "", log.Ldate|log.Ltime),
		verbose: verbose,
	}
}

func (l *Logger) Verbosef(format string, args ...any) {
	if l.verbose {
		l.log.Printf(format, args...)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.log.Printf("Error: %s", msg)
}
