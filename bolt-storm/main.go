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
	"github.com/asdine/storm"
	"github.com/objectbox/go-benchmarks/internal/cmd"
	"github.com/objectbox/go-benchmarks/internal/models"
	"github.com/objectbox/go-benchmarks/internal/perf"
	"os"
	"path/filepath"
)

func main() {
	var options = cmd.GetOptions()

	var executable = &StormPerf{
		path: options.Path,
	}

	var executor = perf.CreateExecutor(executable)
	defer executor.Close()

	executor.Run(options)
}

// perf executable
type StormPerf struct {
	path string
	db   *storm.DB
	tx   storm.Node // used for PutAsync
}

func (exec *StormPerf) Init() error {
	if err := os.RemoveAll(exec.path); err != nil {
		return err
	}

	if err := os.Mkdir(exec.path, 0777); err != nil {
		return err
	}

	if db, err := storm.Open(filepath.Join(exec.path, "test.db"), storm.BoltOptions(0600, nil)); err != nil {
		return err
	} else {
		exec.db = db
	}

	// initialize the DB schema by saving an object and removing it
	// NOTE not sure if this does take significant amount of time but let's give storm a chance to set-up during Init()
	var proto = &models.Entity{}
	if err := exec.db.Save(proto); err != nil {
		return err
	} else if err = exec.db.DeleteStruct(proto); err != nil {
		return err
	}

	return nil
}

func (exec *StormPerf) Close() error {
	if err := exec.db.Close(); err != nil {
		return err
	}

	if err := os.RemoveAll(exec.path); err != nil {
		return err
	}

	return nil
}

func (exec *StormPerf) Size() (uint64, error) {
	if stat, err := os.Stat(filepath.Join(exec.path, "test.db")); err != nil {
		return 0, err
	} else {
		return uint64(stat.Size()), nil
	}
}

func (exec *StormPerf) RemoveAll() error {
	return exec.db.Select().Delete(&models.Entity{})
}

func (exec *StormPerf) RemoveBulk(items []*models.Entity) error {
	// Using queries - it's slower than TX and iterating, as implemented bellow
	//var ids = make([]uint64, len(items))
	//for k, object := range items {
	//	ids[k] = object.Id
	//}
	//return exec.db.Select(q.In("Id", ids)).Delete(&models.Entity{})

	return exec.runInTx(func(tx storm.Node) error {
		for _, object := range items {
			if err := tx.DeleteStruct(object); err != nil {
				return err
			}
		}
		return nil
	})
}

func (exec *StormPerf) PutAsync(item *models.Entity) error {
	// PutAsync is simulated by reusing a transaction and committing it afterwards
	if exec.tx == nil {
		if tx, err := exec.db.Begin(true); err != nil {
			return err
		} else {
			exec.tx = tx
		}
	}

	if err := exec.tx.Save(item); err != nil {
		if err2 := exec.tx.Rollback(); err2 != nil {
			panic(err2)
		}
		exec.tx = nil
		return err
	}

	return nil
}

func (exec *StormPerf) AwaitAsyncCompletion() error {
	if exec.tx != nil {
		var err = exec.tx.Commit()
		exec.tx = nil
		return err
	}

	return nil
}

// run a callback in a transaction
func (exec *StormPerf) runInTx(fn func(tx storm.Node) error) error {
	tx, err := exec.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (exec *StormPerf) PutBulk(items []*models.Entity) (err error) {
	// until the bulk insert feature request from Aug 10, 2017 https://github.com/asdine/storm/issues/176
	// is implemented, we're using manual transactions
	return exec.runInTx(func(tx storm.Node) error {
		for _, item := range items {
			if err := tx.Save(item); err != nil {
				return err
			}
		}
		return nil
	})
}

func (exec *StormPerf) ReadAll() ([]*models.Entity, error) {
	var items []*models.Entity
	if err := exec.db.All(&items); err != nil {
		return nil, err
	}
	return items, nil
}
