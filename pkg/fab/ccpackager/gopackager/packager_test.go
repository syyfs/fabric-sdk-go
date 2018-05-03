/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gopackager

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path"
	"testing"

	"strings"

	"github.com/stretchr/testify/assert"
)

// Test golang ChainCode packaging
func TestNewCCPackage(t *testing.T) {
	pwd, err := os.Getwd()
	assert.Nil(t, err, "error from os.Getwd %v", err)

	ccPackage, err := NewCCPackage("github.com", path.Join(pwd, "../../../../test/fixtures/testdata"))
	assert.Nil(t, err, "error from Create %v", err)

	r := bytes.NewReader(ccPackage.Code)

	gzf, err := gzip.NewReader(r)
	assert.Nil(t, err, "error from gzip.NewReader %v", err)

	tarReader := tar.NewReader(gzf)
	i := 0
	var exampleccExist, eventMetaInfExists, examplecc1MetaInfExists, fooMetaInfoExists, metaInfFooExists bool
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		assert.Nil(t, err, "error from tarReader.Next() %v", err)

		exampleccExist = exampleccExist || header.Name == "src/github.com/example_cc/example_cc.go"
		eventMetaInfExists = eventMetaInfExists || header.Name == "META-INF/sample-json/event.json"
		examplecc1MetaInfExists = examplecc1MetaInfExists || header.Name == "META-INF/example1.json"
		fooMetaInfoExists = fooMetaInfoExists || strings.HasPrefix(header.Name, "foo-META-INF")
		metaInfFooExists = metaInfFooExists || strings.HasPrefix(header.Name, "META-INF-foo")

		i++
	}

	assert.True(t, exampleccExist, "src/github.com/example_cc/example_cc.go does not exists in tar file")
	assert.True(t, eventMetaInfExists, "META-INF/event.json does not exists in tar file")
	assert.True(t, examplecc1MetaInfExists, "META-INF/example1.json does not exists in tar file")
	assert.False(t, fooMetaInfoExists, "invalid root directory found")
	assert.False(t, metaInfFooExists, "invalid root directory found")
}

// Test Package Go ChainCode
func TestEmptyCreate(t *testing.T) {

	_, err := NewCCPackage("", "")
	if err == nil {
		t.Fatalf("Package Empty GoLang CC must return an error.")
	}
}

// Test Bad Package Path for ChainCode packaging
func TestBadPackagePathGoLangCC(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error from os.Getwd %v", err)
	}

	_, err = NewCCPackage("github.com", path.Join(pwd, "../../../../test/fixturesABC"))
	if err == nil {
		t.Fatalf("error expected from Create %v", err)
	}
}

// Test isSource set to true for any go readable files used in ChainCode packaging
func TestIsSourcePath(t *testing.T) {
	keep = []string{}
	isSrcVal := isSource("../")

	if isSrcVal {
		t.Fatalf("error expected when calling isSource %v", isSrcVal)
	}

	// reset keep
	keep = []string{".go", ".c", ".h"}
}

// Test packEntry and generateTarGz with empty file Descriptor
func TestEmptyPackEntry(t *testing.T) {
	emptyDescriptor := &Descriptor{"NewFile", ""}
	err := packEntry(nil, nil, emptyDescriptor)
	if err == nil {
		t.Fatal("packEntry call with empty descriptor info must throw an error")
	}

	_, err = generateTarGz([]*Descriptor{emptyDescriptor})
	if err == nil {
		t.Fatal("generateTarGz call with empty descriptor info must throw an error")
	}

}
