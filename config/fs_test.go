package config

import "testing"

func TestIsYaml(t *testing.T) {
	fs := Fs{}
	if !fs.IsYaml("file.yaml") {
		t.Error("file.yaml should be yaml")
	}
	if !fs.IsYaml("file.yml") {
		t.Error("file.yml should be yaml")
	}
	if fs.IsYaml("file.json") {
		t.Error("file.json should not be yaml")
	}
}

func TestIsCsv(t *testing.T) {
	fs := Fs{}
	if !fs.IsCsv("file.csv") {
		t.Error("file.csv should be csv")
	}
	if fs.IsYaml("file.json") {
		t.Error("file.json should not be csv")
	}
}
