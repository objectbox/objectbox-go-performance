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

package main

import (
	"github.com/objectbox/go-benchmarks/internal/cmd"
	"github.com/objectbox/go-benchmarks/internal/models"
	"github.com/objectbox/go-benchmarks/internal/perf"
	"github.com/objectbox/go-benchmarks/objectbox/obx"
	"github.com/objectbox/objectbox-go/objectbox"
	"os"
)

func main() {
	var options = cmd.GetOptions()

	var executable = &ObjectBoxPerf{
		path: options.Path,
	}

	var executor = perf.CreateExecutor(executable)
	defer executor.Close()

	executor.Run(options)
}

// perf executable
type ObjectBoxPerf struct {
	path string
	ob   *objectbox.ObjectBox
	box  *obx.EntityBox
}

func (exec *ObjectBoxPerf) Init() error {
	if err := os.RemoveAll(exec.path); err != nil {
		return err
	}

	var builder = objectbox.NewBuilder().
		Directory(exec.path).
		Model(obx.ObjectBoxModel()).
		AlwaysAwaitAsync(false)

	if ob, err := builder.Build(); err != nil {
		return err
	} else {
		exec.ob = ob
		exec.box = obx.BoxForEntity(ob)
	}

	return nil
}

func (exec *ObjectBoxPerf) Close() error {
	exec.ob.Close()

	if err := os.RemoveAll(exec.path); err != nil {
		return err
	}

	return nil
}

func (exec *ObjectBoxPerf) RemoveAll() error {
	return exec.box.RemoveAll()
}

func (exec *ObjectBoxPerf) PutAsync(item *models.Entity) error {
	_, err := exec.box.PutAsync(item)
	return err
}

func (exec *ObjectBoxPerf) AwaitAsyncCompletion() error {
	exec.ob.AwaitAsyncCompletion()
	return nil
}

func (exec *ObjectBoxPerf) PutAll(items []*models.Entity) error {
	_, err := exec.box.PutAll(items)
	return err
}

func (exec *ObjectBoxPerf) ReadAll() ([]*models.Entity, error) {
	return exec.box.GetAll()
}
