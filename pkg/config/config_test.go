package config

import (
	"os"
	"testing"
)

func TestConfigurationGetString(t *testing.T) {
	os.Setenv("TEST", "Hello")

	x := GetString("TEST", "", false)
	if x != "Hello" {
		t.Fail()
	}
}

func TestConfigurationGetBoolTrue(t *testing.T) {
	os.Setenv("TEST", "true")

	x := GetBool("TEST", false, false)
	if !x {
		t.Fail()
	}
}

func TestConfigurationGetBoolFalse(t *testing.T) {
	os.Setenv("TEST", "true")

	x := GetBool("TEST", true, false)
	if !x {
		t.Fail()
	}
}

func TestConfigurationGetInt(t *testing.T) {
	os.Setenv("TEST", "5324")

	x := GetInt64("TEST", -1, false)

	if x != 5324 {
		t.Fail()
	}
}

func TestConfigurationGetFloat(t *testing.T) {
	os.Setenv("TEST", "5324.33225")

	x := GetFloat64("TEST", -1.0, false)

	if x != 5324.33225 {
		t.Fail()
	}
}

func TestConfigurationGetStringMap(t *testing.T) {
	os.Setenv("TEST", "1=a,2=b,3=c")

	x := GetStringMap("TEST", make(map[string]string), false)
	if x["1"] != "a" {
		t.Fail()
	}
	if x["2"] != "b" {
		t.Fail()
	}
	if x["3"] != "c" {
		t.Fail()
	}
}
