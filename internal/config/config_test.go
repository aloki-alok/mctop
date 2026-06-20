package config

import "testing"

func TestLoadDefaultsVimOn(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if !Load().Vim {
		t.Fatal("vim should default on when no config exists")
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := Save(Config{Vim: false}); err != nil {
		t.Fatal(err)
	}
	if Load().Vim {
		t.Fatal("saved vim:false should load as false")
	}
	if err := Save(Config{Vim: true}); err != nil {
		t.Fatal(err)
	}
	if !Load().Vim {
		t.Fatal("saved vim:true should load as true")
	}
}
