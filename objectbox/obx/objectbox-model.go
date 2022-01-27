// Code generated by ObjectBox; DO NOT EDIT.

package obx

import (
	"github.com/objectbox/objectbox-go/objectbox"
)

// ObjectBoxModel declares and builds the model from all the entities in the package.
// It is usually used when setting-up ObjectBox as an argument to the Builder.Model() function.
func ObjectBoxModel() *objectbox.Model {
	model := objectbox.NewModel()
	model.GeneratorVersion(6)

	model.RegisterBinding(EntityBinding)
	model.LastEntityId(1, 5847816654868029727)

	return model
}
