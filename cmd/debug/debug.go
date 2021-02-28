package main

import (
	"github.com/davecgh/go-spew/spew"
	"zerosrealm.xyz/tergum/internal/restic"
)

func main() {
	resticExe := restic.New("restic.exe")
	repo := `C:\Nextcloud\Development\Go\Tergum\backup`
	password := "1234"
	snapshots, err := resticExe.Snapshots(repo, password)
	spew.Dump(snapshots, err)
}
