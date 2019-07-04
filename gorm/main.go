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
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/objectbox/go-benchmarks/internal/cmd"
	"github.com/objectbox/go-benchmarks/internal/models"
	"github.com/objectbox/go-benchmarks/internal/perf"
	"os"
	"path/filepath"
)

func main() {
	var options = cmd.GetOptions()

	var executable = &GormPerf{
		path: options.Path,
	}

	var executor = perf.CreateExecutor(executable)
	defer executor.Close()

	executor.Run(options)
}

// perf executable
type GormPerf struct {
	path string
	db   *gorm.DB
	tx   *gorm.DB // used for PutAsync

}

func (exec *GormPerf) Init() error {
	if err := os.RemoveAll(exec.path); err != nil {
		return err
	}

	if err := os.Mkdir(exec.path, 0777); err != nil {
		return err
	}

	if db, err := gorm.Open("sqlite3", filepath.Join(exec.path, "test.db")); err != nil {
		return err
	} else {
		exec.db = db
	}

	exec.db.AutoMigrate(&models.Entity{})

	return exec.db.Error
}

func (exec *GormPerf) Close() error {
	if err := exec.db.Close(); err != nil {
		return err
	}

	if err := os.RemoveAll(exec.path); err != nil {
		return err
	}

	return nil
}

func (exec *GormPerf) Size() (uint64, error) {
	if stat, err := os.Stat(filepath.Join(exec.path, "test.db")); err != nil {
		return 0, err
	} else {
		return uint64(stat.Size()), nil
	}
}

func (exec *GormPerf) RemoveAll() error {
	return exec.db.Delete(models.Entity{}).Error
}

func (exec *GormPerf) RemoveBulk(items []*models.Entity) error {
	return exec.runInTx(func(tx *gorm.DB) error {
		// sqlite takes at most 999 variables by default, see SQLITE_MAX_VARIABLE_NUMBER
		// if we pass more, we get an error "too many sql variables"
		// therefore, we're running a delete for at most 999 items at a time
		const limit = 999
		for i := 0; i <= len(items)/limit; i++ {
			var ids = make([]uint64, limit)

			for j := 0; j < limit; j++ {
				var idx = i*limit + j
				if idx >= len(items) {
					break
				}

				ids[j] = items[i*limit+j].Id
			}

			if err := tx.Where(ids).Delete(models.Entity{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (exec *GormPerf) PutAsync(item *models.Entity) error {
	// PutAsync is simulated by reusing a transaction and committing it afterwards
	if exec.tx == nil {
		exec.tx = exec.db.Begin().Model(&models.Entity{})
	}

	if item.Id == 0 {
		exec.tx.Create(item)
	} else {
		exec.tx.Save(item)
	}

	if err := exec.tx.Error; err != nil {
		exec.tx.Rollback()
		exec.tx = nil
		return err
	}

	return nil
}

func (exec *GormPerf) AwaitAsyncCompletion() error {
	if exec.tx != nil {
		var err = exec.tx.Commit().Error
		exec.tx = nil
		return err
	}

	return nil
}

// run a callback in a transaction
func (exec *GormPerf) runInTx(fn func(tx *gorm.DB) error) (err error) {
	// see http://gorm.io/docs/transactions.html
	tx := exec.db.Begin().Model(&models.Entity{})
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if err == nil {
				err = tx.Error
			}

			if err == nil {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (exec *GormPerf) PutBulk(items []*models.Entity) (err error) {
	// until the bulk insert feature request from Oct 16, 2014 https://github.com/jinzhu/gorm/issues/255
	// is implemented, we're using manual transactions

	return exec.runInTx(func(tx *gorm.DB) error {
		for _, item := range items {
			if item.Id == 0 {
				tx.Create(item)
			} else {
				tx.Save(item)
			}
			if tx.Error != nil {
				return tx.Error
			}
		}
		return nil
	})
}

func (exec *GormPerf) ReadAll() ([]*models.Entity, error) {
	var items []*models.Entity
	exec.db.Find(&items)
	return items, exec.db.Error
}
