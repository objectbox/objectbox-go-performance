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

func (exec *GormPerf) RemoveAll() error {
	return exec.db.Delete(models.Entity{}).Error
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

func (exec *GormPerf) PutAll(items []*models.Entity) (err error) {
	// until the bulk insert feature request from Oct 16, 2014 https://github.com/jinzhu/gorm/issues/255
	// is implemented, we're using manual transactions, see http://gorm.io/docs/transactions.html
	tx := exec.db.Begin().Model(&models.Entity{})
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if err == nil {
				err = tx.Error
			}
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	for _, item := range items {
		if item.Id == 0 {
			tx.Create(item)
		} else {
			tx.Save(item)
		}
		if tx.Error != nil {
			tx.Rollback()
			return tx.Error
		}
	}

	return tx.Commit().Error
}

func (exec *GormPerf) ReadAll() ([]*models.Entity, error) {
	var items []*models.Entity
	exec.db.Find(&items)
	return items, exec.db.Error
}
