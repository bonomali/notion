package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/tmc/notion"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	flagVerbose = flag.Bool("v", false, "verbose")
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		fmt.Fprintln(os.Stderr, "please provide text block id as parameter")
		os.Exit(1)
	}
	if err := run(flag.Args()[0]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(id string) error {
	if terminal.IsTerminal(0) {
		flag.Usage()
		log.Fatalln("stdin appears to be a tty device. This tool is meant to be invoked and have stdin provided by a pipe")
	}
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	opts := []notion.ClientOption{
		notion.WithToken(os.Getenv("NOTION_TOKEN")),
	}
	if *flagVerbose {
		opts = append(opts, notion.WithDebugLogging())
	}
	c, err := notion.NewClient(opts...)
	if err != nil {
		return err
	}
	pageInfo, err := c.GetRecordValues(notion.Record{Table: "block", ID: id})
	if err != nil {
		return err
	}
	if pageInfo[0].Value == nil {
		return fmt.Errorf("issue fetching content, Role=%v", pageInfo[0].Role)
	}
	b, err := c.GetBlock(pageInfo[0].Value.ID)
	if err != nil {
		return err
	}
	if *flagVerbose {
		json.NewEncoder(os.Stderr).Encode(b)
	}

	content := string(bytes.TrimSpace(data))
	if err := c.UpdateBlock(b.ID, "properties.title", content); err != nil {
		return err
	}
	fmt.Println(content) // echo back out for editor use
	return nil
}
