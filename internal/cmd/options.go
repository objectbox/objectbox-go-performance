/*
 * Copyright 2019 ObjectBox Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"flag"
	"github.com/objectbox/objectbox-go-performance/internal/perf"
)

func GetOptions() perf.Options {
	// start with a copy of defaults
	var o = perf.OptionsDefaults

	flag.StringVar(&o.Path, "db", o.Path, "database directory")
	flag.IntVar(&o.Count, "count", o.Count, "number of objects")
	flag.IntVar(&o.Runs, "runs", o.Runs, "number of times the tests should be executed")
	flag.BoolVar(&o.Profile, "profile", o.Profile, "enable profiling")
	flag.Parse()

	return o
}
