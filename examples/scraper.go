package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/phr3nzy/tango"
)

func main() {
	type state struct {
		Attempts          int
		PluginState       string
		RequestBodyReader *bufio.Reader
		Content           []byte
	}

	type services struct {
		http *http.Client
	}

	ctx := &tango.MachineContext[services, state]{
		Services: services{
			http: &http.Client{
				Timeout: time.Second * 10,
			},
		},
		State: state{
			Attempts:    0,
			PluginState: "plugin state",
		},
		PreviousResult: nil,
	}

	config := &tango.MachineConfig[services, state]{
		Log:      true,
		LogLevel: "info",
		Plugins:  []tango.Plugin[services, state]{},
	}

	steps := []tango.Step[services, state]{
		{
			Name: "visit website",
			Execute: func(ctx *tango.MachineContext[services, state]) (*tango.Response[services, state], error) {
				ctx.State.Attempts++
				resp, err := ctx.Services.http.Get("https://google.com/")
				if err != nil {
					return ctx.Machine.Error(err), nil
				}
				defer resp.Body.Close()

				ctx.State.RequestBodyReader = bufio.NewReader(resp.Body)
				ctx.State.Content, err = io.ReadAll(ctx.State.RequestBodyReader)
				if err != nil {
					return ctx.Machine.Error(err.Error()), nil
				}

				if resp.StatusCode != 200 {
					return ctx.Machine.Error(fmt.Sprintf(
						"status code: %d, body: %s",
						resp.StatusCode,
						string(ctx.State.Content),
					)), nil
				}

				return ctx.Machine.Next("page visited"), nil
			},
			Compensate: func(ctx *tango.MachineContext[services, state]) (*tango.Response[services, state], error) {
				return ctx.Machine.Done("Compensate"), nil
			},
		},
		{
			Name: "write to disk",
			Execute: func(ctx *tango.MachineContext[services, state]) (*tango.Response[services, state], error) {
				// if the file exists, return an error
				_, err := os.Stat("examples/output.html")
				if err == nil {
					return ctx.Machine.Error("file already exists"), nil
				}

				f, err := os.Create("examples/output.html")
				if err != nil {
					return ctx.Machine.Error(err.Error()), nil
				}
				defer f.Close()

				_, err = f.Write(ctx.State.Content)
				if err != nil {
					return ctx.Machine.Error(err.Error()), nil
				}

				return ctx.Machine.Done("wrote the file to disk"), nil
			},
			Compensate: func(ctx *tango.MachineContext[services, state]) (*tango.Response[services, state], error) {
				_, err := os.Stat("examples/output.html")
				if err == nil {
					fmt.Println("removing file")
					err = os.Remove("examples/output.html")
					if err != nil {
						return ctx.Machine.Error(err.Error()), nil
					}
				}

				return ctx.Machine.Next("compensated"), nil
			},
		},
	}

	machine := tango.NewMachine[services, state](
		"scraper",
		steps,
		ctx,
		config,
		&tango.ConcurrentStrategy[services, state]{
			Concurrency: 1,
		},
	)
	ctx.Machine = machine

	response, err := machine.Run()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(response)
}
