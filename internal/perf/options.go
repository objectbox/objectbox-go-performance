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

package perf

type Options struct {
	Path     string
	Count    int
	Runs     int
	ManualGc bool
	Profile  bool
}

var OptionsDefaults = Options{
	// not using field names here so that we don't forget to add default when the Options struct is updated
	"testdata",
	10000,
	10,
	false,
	false,
}
