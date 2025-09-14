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
	"github.com/sirupsen/logrus"

	"fmt"
	"os"
)

// Populated by goreleaser
var (
	version = "???"
	date    = "???"
)

func init() {
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	// LOG_LEVEL not set, let's default to info
	if !ok {
		lvl = "info"
	}
	// parse string, this is built-in feature of logrus
	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.DebugLevel
	}
	// set global log level
	logrus.SetLevel(ll)
}

// Commands managed by Cobra
func main() {
	bi := config.BuildInfo{version, date}
	if err := cmd.New(bi).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
