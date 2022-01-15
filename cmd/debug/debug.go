package main

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"zerosrealm.xyz/tergum/internal/restic"
)

func main() {
	resticExe := restic.New(context.Background(), `restic.exe`)
	repo := `C:\Nextcloud\Development\Go\Tergum\backup`
	password := "1234"
	// snapshots, err := resticExe.Debug(repo, `C:\Users\Mikkel\Downloads\10GB.bin`, password)
	snapshots, err := resticExe.Snapshots(repo, password)
	if err != nil {
		panic(err)
	}
	spew.Dump(snapshots)
}
