// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/rykov/paperboy/cmd"
	"github.com/rykov/paperboy/config"

	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Populated by goreleaser
var (
	version = "???"
	date    = "???"
)

// Commands managed by Cobra
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop() // unregistered signal handlers

	bi := config.BuildInfo{
		Version:   version,
		BuildDate: date,
	}

	if err := cmd.New(bi).ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, context.Canceled) {
			os.Exit(130) // Standard SIGINT exit code
		} else {
			os.Exit(1)
		}
	}
}
