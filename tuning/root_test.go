package tuning_test

import (
	"testing"

	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
)

func TestRootForms(t *testing.T) {
	t.Parallel()

	t.Run("class", func(t *testing.T) {
		t.Parallel()
		class := pitchclass.MustNew(9, 12)
		root := tuning.RootFromClass(class)
		assert.True(t, root.IsValid())
		gotClass, ok := root.Class()
		assert.True(t, ok)
		assert.Equal(t, class, gotClass)
		gotName, ok := root.Name()
		assert.False(t, ok)
		assert.False(t, gotName.IsValid())
		assert.Equal(t, "Root(9 mod 12)", root.String())
	})

	t.Run("name", func(t *testing.T) {
		t.Parallel()
		root := tuning.RootFromName(spelling.C())
		assert.True(t, root.IsValid())
		gotName, ok := root.Name()
		assert.True(t, ok)
		assert.Equal(t, spelling.C(), gotName)
		gotClass, ok := root.Class()
		assert.False(t, ok)
		assert.False(t, gotClass.IsValid())
		assert.Equal(t, "Root(C)", root.String())
	})
}

func TestRootZeroAndEquality(t *testing.T) {
	t.Parallel()
	var zero tuning.Root
	assert.False(t, zero.IsValid())
	assert.Equal(t, "Root(unset)", zero.String())
	assert.True(t, zero.Equal(tuning.Root{}))
	assert.True(t, zero == tuning.Root{})
	_, hasClass := zero.Class()
	assert.False(t, hasClass)
	_, hasName := zero.Name()
	assert.False(t, hasName)

	assert.False(t, tuning.RootFromClass(pitchclass.Class{}).IsValid())
	assert.False(t, tuning.RootFromName(spelling.Spelling{}).IsValid())
	assert.True(t, tuning.RootFromClass(pitchclass.Class{}).Equal(zero))
	assert.True(t, tuning.RootFromName(spelling.Spelling{}).Equal(zero))

	classC := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	nameC := tuning.RootFromName(spelling.C())
	assert.False(t, classC.Equal(nameC), "different Root forms are not equal")
	assert.False(t, classC == nameC)
	assert.False(t, classC.Equal(zero))
	assert.False(t, nameC.Equal(zero))
	assert.True(t, classC.Equal(tuning.RootFromClass(pitchclass.MustNew(0, 12))))
	assert.True(t, nameC.Equal(tuning.RootFromName(spelling.C())))

	byRoot := map[tuning.Root]string{classC: "class", nameC: "name"}
	assert.Equal(t, "class", byRoot[tuning.RootFromClass(pitchclass.MustNew(0, 12))])
	assert.Equal(t, "name", byRoot[tuning.RootFromName(spelling.C())])
}
