// Copyright 2020 - 2025 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html/charset"
)

// CollectReferencedFiles recursively collects all files referenced by import and include statements
func CollectReferencedFiles(filePath, inputDir string, allFiles map[string]bool, options *Options) error {
	// Skip if already processed
	if allFiles[filePath] {
		return nil
	}

	// Mark as being processed
	allFiles[filePath] = true

	fileDir := filepath.Dir(filePath)
	fi, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return nil
	}

	xmlFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer xmlFile.Close()

	decoder := xml.NewDecoder(xmlFile)
	decoder.CharsetReader = charset.NewReaderLabel

	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}

		switch element := token.(type) {
		case xml.StartElement:
			switch element.Name.Local {
			case "import":
				schemaLocation := getSchemaLocation(element)
				if schemaLocation != "" && !isValidURL(schemaLocation) {
					refFile := filepath.Join(fileDir, schemaLocation)
					if _, err := os.Stat(refFile); err == nil {
						// Recursively collect from referenced file
						if err := CollectReferencedFiles(refFile, inputDir, allFiles, options); err != nil {
							return err
						}
					}
				}
			case "include":
				schemaLocation := getSchemaLocation(element)
				if schemaLocation != "" && !isValidURL(schemaLocation) {
					refFile := filepath.Join(fileDir, schemaLocation)
					if _, err := os.Stat(refFile); err == nil {
						// Recursively collect from referenced file
						if err := CollectReferencedFiles(refFile, inputDir, allFiles, options); err != nil {
							return err
						}
					}
				}
			}
		default:
		}
	}

	return nil
}

// getSchemaLocation extracts the schemaLocation attribute from an XML element
func getSchemaLocation(element xml.StartElement) string {
	for _, attr := range element.Attr {
		if attr.Name.Local == "schemaLocation" {
			return strings.TrimSpace(attr.Value)
		}
	}
	return ""
}
