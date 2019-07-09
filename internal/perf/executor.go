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

import (
	"fmt"
	"github.com/objectbox/go-benchmarks/internal/models"
	"github.com/pkg/profile"
	"log"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

type Executor struct {
	exec  Executable
	times map[string][]time.Duration // arrays of runtimes indexed by function name
}

func CreateExecutor(executable Executable) *Executor {
	var result = &Executor{
		times: map[string][]time.Duration{},
		exec:  executable,
	}

	result.Init()
	return result
}

func (perf *Executor) Init() {
	defer perf.trackTime(time.Now())
	assert(perf.exec.Init())
}

func (perf *Executor) Close() {
	defer perf.trackTime(time.Now())
	assert(perf.exec.Close())
}

func removeIds(items []*models.Entity) {
	for _, item := range items {
		item.Id = 0
	}
}

func (perf *Executor) Run(options Options) {
	if options.Profile {
		defer profile.Start().Stop()
	}

	log.Printf("running the test %d times with %d objects", options.Runs, options.Count)

	if options.ManualGc {
		// disable automatic garbage collector
		log.Println("using manual GC management")
		debug.SetGCPercent(-1)
	}

	var inserts = perf.PrepareData(options.Count)
	var size uint64

	for i := 0; i < options.Runs; i++ {
		perf.PutBulk(inserts)
		items := perf.ReadAll(options.Count)
		perf.UpdateBulk(items)

		if size_, err := perf.exec.Size(); err != nil {
			panic(err)
		} else {
			size = size_
		}

		if len(items) >= 100 {
			perf.Query100IdsBetween(items[len(items)-100].Id, items[len(items)-1].Id)
		}

		var prefix = "Entity no. 1"
		var expectedPrefixMatches = 0
		for _, object := range items {
			if strings.HasPrefix(object.String, prefix) {
				expectedPrefixMatches++
			}
		}
		log.Printf("QueryStringPrefix must match %d items", expectedPrefixMatches)
		perf.QueryStringPrefix(prefix, expectedPrefixMatches)

		perf.RemoveAll()

		// insert again and delete by id
		removeIds(inserts)
		perf.PutBulk(inserts)
		perf.RemoveBulk(inserts)

		log.Printf("%d/%d finished", i+1, options.Runs)

		if options.ManualGc {
			// manually invoke GC out of benchmarked time
			runtime.GC()
			log.Printf("%d/%d garbage-collector executed", i+1, options.Runs)
		}
	}

	perf.PrintTimes([]string{
		"Init",
		"PutBulk",
		"ReadAll",
		"UpdateBulk",
		"RemoveAll",
		"RemoveBulk",
		"Query100IdsBetween",
		"QueryStringPrefix",
	})

	fmt.Println(fmt.Sprintf("DB size after update, before remove: %d", size))
}

func (perf *Executor) RemoveAll() {
	defer perf.trackTime(time.Now())
	err := perf.exec.RemoveAll()
	if err != nil {
		panic(err)
	}
}

func (perf *Executor) RemoveBulk(items []*models.Entity) {
	defer perf.trackTime(time.Now())
	assert(perf.exec.RemoveBulk(items))
}

func (perf *Executor) PrepareData(count int) []*models.Entity {
	defer perf.trackTime(time.Now())

	var result = make([]*models.Entity, count)
	for i := 0; i < count; i++ {
		result[i] = &models.Entity{
			String:  fmt.Sprintf("Entity no. %d", i),
			Float64: float64(i),
			Int32:   int32(i),
			Int64:   int64(i),
		}
	}

	return result
}

func (perf *Executor) PutAsync(items []*models.Entity) {
	defer perf.trackTime(time.Now())

	for _, item := range items {
		assert(perf.exec.PutAsync(item))
	}

	assert(perf.exec.AwaitAsyncCompletion())
}

func (perf *Executor) PutBulk(items []*models.Entity) {
	defer perf.trackTime(time.Now())
	assert(perf.exec.PutBulk(items))
}

func (perf *Executor) ReadAll(expectedCount int) []*models.Entity {
	defer perf.trackTime(time.Now())

	if items, err := perf.exec.ReadAll(); err != nil {
		panic(err)
	} else if len(items) != expectedCount {
		panic("invalid number of objects read")
	} else {
		return items
	}
}

func (perf *Executor) ChangeValues(items []*models.Entity) {
	defer perf.trackTime(time.Now())

	count := len(items)
	for i := 0; i < count; i++ {
		items[i].Int64 = items[i].Int64 * 2
	}
}

func (perf *Executor) UpdateBulk(items []*models.Entity) {
	defer perf.trackTime(time.Now())
	assert(perf.exec.PutBulk(items))
}

func (perf *Executor) Query100IdsBetween(min, max uint64) {
	defer perf.trackTime(time.Now())
	if items, err := perf.exec.QueryIdBetween(min, max); err != nil {
		panic(err)
	} else if uint64(len(items)) != max-min+1 {
		panic(fmt.Errorf("invalid number of objects returned by QueryIdBetween(%d, %d): %d", min, max,
			len(items)))
	}
}

func (perf *Executor) QueryStringPrefix(prefix string, expectedCount int) {
	defer perf.trackTime(time.Now())
	if items, err := perf.exec.QueryStringPrefix(prefix); err != nil {
		panic(err)
	} else if len(items) != expectedCount {
		panic(fmt.Errorf("invalid number of objects returned by QueryStringPrefix - %d instead of %d",
			len(items), expectedCount))
	}
}

func (perf *Executor) trackTime(start time.Time) {
	elapsed := time.Since(start)

	pc, _, _, _ := runtime.Caller(1)
	fun := filepath.Ext(runtime.FuncForPC(pc).Name())[1:]
	perf.times[fun] = append(perf.times[fun], elapsed)
}

func (perf *Executor) PrintTimes(functions []string) {
	// print the whole data as a table
	fmt.Println("Function\tRuns\tAverage ms\tAll times")

	if len(functions) == 0 {
		for fun := range perf.times {
			functions = append(functions, fun)
		}
	}

	for _, fun := range functions {
		times := perf.times[fun]

		sum := int64(0)
		for _, duration := range times {
			sum += duration.Nanoseconds()
		}
		fmt.Printf("%s\t%d\t%f", fun, len(times), float64(sum/int64(len(times)))/1000000)

		for _, duration := range times {
			fmt.Printf("\t%f", float64(duration.Nanoseconds())/1000000)
		}
		fmt.Println()
	}
}
