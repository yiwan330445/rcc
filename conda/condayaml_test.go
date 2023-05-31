package conda_test

import (
	"testing"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/hamlet"
)

func TestCanParseDependencies(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	must_be.Nil(conda.AsDependency(""))
	wont_be.Nil(conda.AsDependency("python"))
	must_be.Equal("python", conda.AsDependency("python").Name)
	must_be.Equal("", conda.AsDependency("python").Qualifier)
	must_be.Equal("", conda.AsDependency("python").Versions)
	wont_be.Nil(conda.AsDependency("python=3.9.13"))
	must_be.Equal("python=3.7.4", conda.AsDependency("python=3.7.4").Original)
	must_be.Equal("python", conda.AsDependency("python=3.7.4").Name)
	must_be.Equal("=", conda.AsDependency("python=3.7.4").Qualifier)
	must_be.Equal("3.7.4", conda.AsDependency("python=3.7.4").Versions)
	must_be.Equal("1.8", conda.AsDependency("kamalia 1.8").Versions)
}

func TestCanCompareDependencies(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	first := conda.AsDependency("python")
	second := conda.AsDependency("python=3.7.7")
	third := conda.AsDependency("python=3.9.13")
	fourth := conda.AsDependency("robotframework=3.2")

	wont_be.True(first.IsExact())
	must_be.True(second.IsExact())

	must_be.True(first.SameAs(second))
	must_be.True(first.SameAs(third))
	wont_be.True(third.SameAs(fourth))

	chosen, err := first.ChooseSpecific(second)
	must_be.Nil(err)
	wont_be.Nil(chosen)
	must_be.Same(second, chosen)

	chosen, err = third.ChooseSpecific(first)
	must_be.Nil(err)
	wont_be.Nil(chosen)
	must_be.Same(third, chosen)

	chosen, err = first.ChooseSpecific(fourth)
	wont_be.Nil(err)
	must_be.Equal("Not same component: python vs. robotframework", err.Error())
	must_be.Nil(chosen)

	chosen, err = second.ChooseSpecific(third)
	wont_be.Nil(err)
	must_be.Equal("Wont choose between dependencies: python=3.7.7 vs. python=3.9.13", err.Error())
	must_be.Nil(chosen)
}

func TestCanCreateCondaYamlFromEmptyByteSlice(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut, err := conda.CondaYamlFrom([]byte(""))
	must_be.Nil(err)
	wont_be.Nil(sut)
	must_be.Equal("", sut.Name)
	must_be.Equal("", sut.Prefix)
	must_be.Equal(0, len(sut.Channels))
	must_be.Equal(0, len(sut.Conda))
	must_be.Equal(0, len(sut.Pip))
	must_be.Equal(0, len(sut.PostInstall))
}

func TestCanReadCondaYaml(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut, err := conda.ReadCondaYaml("testdata/conda.yaml")
	must_be.Nil(err)
	wont_be.Nil(sut)
	must_be.Equal("", sut.Name)
	must_be.Equal("", sut.Prefix)
	must_be.Equal(2, len(sut.Channels))
	must_be.Equal(4, len(sut.Conda))
	must_be.Equal(1, len(sut.Pip))
}

func TestCanMergeTwoEnvironments(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	left, err := conda.ReadCondaYaml("testdata/third.yaml")
	must_be.Nil(err)
	wont_be.Nil(left)
	right, err := conda.ReadCondaYaml("testdata/other.yaml")
	must_be.Nil(err)
	wont_be.Nil(right)
	sut, err := left.Merge(right)
	must_be.Nil(err)
	wont_be.Nil(sut)
	must_be.Equal("", sut.Name)
	must_be.Equal(2, len(sut.Channels))
	must_be.Equal(4, len(sut.Conda))
	must_be.Equal(1, len(sut.Pip))
	content, err := sut.AsYaml()
	must_be.Nil(err)
	must_be.True(len(content) > 100)
	pure, err := sut.AsPureConda().AsYaml()
	must_be.Nil(err)
	must_be.True(len(pure) > 100)
	must_be.True(len(content) > len(pure))
}

func TestCanCreateEmptyEnvironment(t *testing.T) {
	_, wont_be := hamlet.Specifications(t)

	sut := conda.SummonEnvironment("tmp/missing.yaml")
	wont_be.Nil(sut)
}

func TestCanGetLayersFromCondaYaml(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut, err := conda.ReadCondaYaml("testdata/layers.yaml")
	must_be.Nil(err)
	wont_be.Nil(sut)

	layers := sut.AsLayers()
	wont_be.Nil(layers)
	wont_be.Equal(len(layers[0]), 0)
	must_be.True(len(layers[0]) < len(layers[1]))
	must_be.True(len(layers[1]) < len(layers[2]))
	wont_be.Equal(layers[0], layers[1])
	wont_be.Equal(layers[0], layers[2])
	wont_be.Equal(layers[1], layers[2])

	must_be.Equal("0d8cc85130420984", common.BlueprintHash([]byte(layers[0])))
	must_be.Equal("5be3e197c8c2c67d", common.BlueprintHash([]byte(layers[1])))
	must_be.Equal("d310697aca0840a1", common.BlueprintHash([]byte(layers[2])))

	fingerprints := sut.FingerprintLayers()
	must_be.Equal("0d8cc85130420984", fingerprints[0])
	must_be.Equal("5be3e197c8c2c67d", fingerprints[1])
	must_be.Equal("d310697aca0840a1", fingerprints[2])
}
