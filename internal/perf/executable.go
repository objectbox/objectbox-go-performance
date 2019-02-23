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

import "github.com/objectbox/go-benchmarks/internal/models"

type Executable interface {
	Init() error
	Close() error
	RemoveAll() error
	RemoveBulk(items []*models.Entity) error
	PutAsync(*models.Entity) error
	AwaitAsyncCompletion() error
	PutBulk(items []*models.Entity) error
	ReadAll() ([]*models.Entity, error)
}
