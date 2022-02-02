package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func initialize() (rows int, columns int) {
	var output bytes.Buffer

	cmd := exec.Command("stty", "cbreak", "-echo", "size")
	cmd.Stdout = &output
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	_, err = fmt.Sscanf(output.String(), "%d %d", &rows, &columns)
	if err != nil {
		panic(fmt.Errorf("Unable to read screen size: %s", err))
	}

	return
}

func cleanup() {
	showCursor()
	cmd := exec.Command("stty", "-cbreak", "echo")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func clear() {
	fmt.Print("\x1b[2J")
}

func move(x, y int) {
	fmt.Printf("\x1b[%d;%df", x+1, y+1)
}

func resize() (rows int, columns int) {
	var output bytes.Buffer

	cmd := exec.Command("stty", "size")
	cmd.Stdout = &output
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	_, err = fmt.Sscanf(output.String(), "%d %d", &rows, &columns)
	if err != nil {
		panic(fmt.Errorf("Unable to read screen size: %s", err))
	}

	return
}

func hideCursor() {
	fmt.Print("\x1b[?25l")
}

func showCursor() {
	fmt.Print("\x1b[?25h")
}

func interruptHandling(endrow, endcol int) {
	c := make(chan os.Signal)
	go func() {
		<-c
		move(endrow, endcol)
		cleanup()
		os.Exit(0)
	}()
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
}
