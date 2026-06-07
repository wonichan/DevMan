package versionmanager

import "testing"

func TestGenerateShimQuotesTargetAndForwardsArgs(t *testing.T) {
	shim, err := GenerateShim(`D:\Tools With Spaces\Go\bin\go.exe`)
	if err != nil {
		t.Fatalf("GenerateShim failed: %v", err)
	}

	expected := "@echo off\r\n\"D:\\Tools With Spaces\\Go\\bin\\go.exe\" %*\r\n"
	if shim != expected {
		t.Fatalf("shim = %q, want %q", shim, expected)
	}
}

func TestGenerateShimRejectsQuoteInTarget(t *testing.T) {
	_, err := GenerateShim(`D:\tools\go" & calc.exe`)
	if err == nil {
		t.Fatal("expected quote validation error")
	}
}

func TestShimTargetsIncludesFlutterDart(t *testing.T) {
	targets, err := ShimTargets("flutter", `D:\sdks\flutter`)
	if err != nil {
		t.Fatalf("ShimTargets failed: %v", err)
	}

	if targets["dart.cmd"] != `D:\sdks\flutter\bin\dart.bat` {
		t.Fatalf("dart target = %q", targets["dart.cmd"])
	}
}
