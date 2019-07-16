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
	"fmt"
	"github.com/objectbox/objectbox-go-benchmarks/internal/cmd"
	"github.com/objectbox/objectbox-go-benchmarks/internal/models"
	"github.com/objectbox/objectbox-go-benchmarks/internal/perf"
	"github.com/objectbox/objectbox-go-benchmarks/objectbox/obx"
	"github.com/objectbox/objectbox-go/objectbox"
	"os"
	"path/filepath"
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
		Model(obx.ObjectBoxModel())

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

func (exec *ObjectBoxPerf) Size() (uint64, error) {
	if stat, err := os.Stat(filepath.Join(exec.path, "data.mdb")); err != nil {
		return 0, err
	} else {
		return uint64(stat.Size()), nil
	}
}

func (exec *ObjectBoxPerf) RemoveAll() error {
	return exec.box.RemoveAll()
}

func (exec *ObjectBoxPerf) RemoveBulk(items []*models.Entity) error {
	if count, err := exec.box.RemoveMany(items...); err != nil {
		return err
	} else if count != uint64(len(items)) {
		return fmt.Errorf("removed only %d out of %d objects", count, len(items))
	}
	return nil
}

func (exec *ObjectBoxPerf) PutAsync(item *models.Entity) error {
	_, err := exec.box.PutAsync(item)
	return err
}

func (exec *ObjectBoxPerf) AwaitAsyncCompletion() error {
	exec.ob.AwaitAsyncCompletion()
	return nil
}

func (exec *ObjectBoxPerf) PutBulk(items []*models.Entity) error {
	_, err := exec.box.PutMany(items)
	return err
}

func (exec *ObjectBoxPerf) ReadAll() ([]*models.Entity, error) {
	return exec.box.GetAll()
}

func (exec *ObjectBoxPerf) QueryIdBetween(min, max uint64) ([]*models.Entity, error) {
	return exec.box.Query(obx.Entity_.Id.Between(min, max)).Find()
}

func (exec *ObjectBoxPerf) QueryStringPrefix(prefix string) ([]*models.Entity, error) {
	return exec.box.Query(obx.Entity_.String.HasPrefix(prefix, true)).Find()
}
