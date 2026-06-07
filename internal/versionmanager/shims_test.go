package versionmanager

import "testing"

func TestGenerateShimQuotesTargetAndForwardsArgs(t *testing.T) {
	shim := GenerateShim(`D:\Tools With Spaces\Go\bin\go.exe`)

	expected := "@echo off\r\n\"D:\\Tools With Spaces\\Go\\bin\\go.exe\" %*\r\n"
	if shim != expected {
		t.Fatalf("shim = %q, want %q", shim, expected)
	}
}

func TestShimTargetsIncludesFlutterDart(t *testing.T) {
	targets, err := ShimTargets("flutter", `D:\sdks\flutter`)
	if err != nil {
		t.Fatalf("ShimTargets failed: %v", err)
	}

	if targets["dart.cmd"] != `D:\sdks\flutter\bin\dart.exe` {
		t.Fatalf("dart target = %q", targets["dart.cmd"])
	}
}
