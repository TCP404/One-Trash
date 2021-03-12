package main

import (
	"Trash-clir/recorder"
	"fmt"
	"os"
	"path"
	"path/filepath"

	clir "github.com/leaanthony/clir"
)

var (
	// HOME is the home directory
	HOME string
	// TRASHHOME is the trash can directory
	TRASHHOME string
	// TRASHLOGPATH files records the Deletion date, Original path, permission
	TRASHLOGPATH string
	// TRASHCONFIGPATH file saves some User-defined infomation.
	TRASHCONFIGPATH string
)

// init() checkes the trash can path and create it when it is not exist.
func init() {
	HOME = os.Getenv("HOME")
	// TRASHHOME = HOME + "/.local/share/Trash"
	TRASHHOME = HOME + "/.Trash"
	TRASHLOGPATH = HOME + "/.trash/deleted"
	TRASHCONFIGPATH = HOME + "/.trash/config"

	_, err := os.Stat(TRASHHOME)
	if os.IsNotExist(err) {
		err := os.MkdirAll(TRASHHOME, os.ModePerm)
		dealError(err, TRASHHOME)
	}

	_, err = os.Stat(HOME + "/.trash")
	if os.IsNotExist(err) {
		err := os.Mkdir(HOME+"/.trash", os.ModePerm)
		dealError(err)
	}

	_, err = os.Stat(TRASHLOGPATH)
	if os.IsNotExist(err) {
		_, err := os.Create(TRASHLOGPATH)
		dealError(err)
	}

	_, err = os.Stat(TRASHCONFIGPATH)
	if os.IsNotExist(err) {
		_, err := os.Create(TRASHCONFIGPATH)
		dealError(err)
	}
}

func main() {
	report := recorder.NewReporter(TRASHHOME, TRASHLOGPATH, TRASHCONFIGPATH)

	var delTarget string
	var undelTarget bool
	var list bool
	var clear bool

	cli := clir.NewCli("rms", "A safe deletion command for linux", "v0.0.2")
	cli.StringFlag("r", "Delete the files or directorys", &delTarget)
	cli.BoolFlag("u", "Undelete the files or directorys", &undelTarget)
	cli.BoolFlag("l", "List the trash can", &list)
	cli.BoolFlag("c", "Clear the trash can", &clear)

	cli.Action(func() error {
		if delTarget != "" {
			del(cli, report, delTarget)
		}

		if undelTarget {
			undel(cli, report)
		}
		if list {
			listTrash(report)
		}

		if clear {
			clearTrash(report)
		}
		return nil
	})
	cli.Run()
}

// del deal with the delete feature
func del(cli *clir.Cli, report *recorder.Reporter, delTarget string) {
	defer func() {
		if msg := recover(); msg != nil {
			fmt.Println(msg)
		}
	}()

	err := _del(report, delTarget)
	dealError(err)

	if len(cli.OtherArgs()) <= 0 {
		return
	}

	for _, f := range cli.OtherArgs() {
		err := _del(report, f)
		dealError(err)
	}
}

func _del(report *recorder.Reporter, delTarget string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// record
	report.Add(delTarget)

	// rename
	var oldpath string
	var newpath string
	oldpath = path.Join(pwd, delTarget)
	newpath = path.Join(TRASHHOME, path.Base(delTarget))
	err = os.Rename(oldpath, newpath)
	if err != nil {
		return err
	}
	return nil
}

func undel(cli *clir.Cli, report *recorder.Reporter) {
	// TODO 恢复到原来的位置可以作为另外的选项，如 -u 默认恢复到当前目录，-uo恢复到原来的位置

	newPath := report.Sub()
	if newPath == "" {
		fmt.Println("Trash can may be is EMPTY.")
		return
	}
	// Check OVERRIDE
	_, err := os.Stat(newPath)
	if os.IsExist(err) {
		var choice string
		fmt.Printf("%s is existing, are you going to OVERRIDE it [y|N]? ", newPath)
		fmt.Scanln(&choice)
		if choice[0] != 'y' && choice[0] != 'Y' {
			return
		}
	}
	// Undel
	oldPath := path.Join(TRASHHOME, filepath.Base(newPath))
	err = os.Rename(oldPath, newPath)
	dealError(err)
}

func listTrash(r *recorder.Reporter) {
	fmt.Println("List trash can")
	fmt.Println("   Perm       Deletion-Date        Path")
	if n, s := r.List(); n <= 0 {
		fmt.Println("[ Trash can is EMPTY. ]")
	} else {
		fmt.Println(s)
	}
}

func clearTrash(r *recorder.Reporter) {
	var choice string
	fmt.Printf("This operation can't recover Clear Trash? [y|N]")
	fmt.Scanln(&choice)
	if choice[0] != 'y' && choice[0] != 'Y' {
		return
	}
	err := os.RemoveAll(TRASHHOME + "/")
	dealError(err)
	r.Clear()
}

func dealError(err error, v ...interface{}) {
	switch err.(type) {
	case *os.PathError:
		fmt.Println("PathError", err)
	case *os.LinkError:
		fmt.Println("LinkError", err)
	}
}
