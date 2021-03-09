package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	clir "github.com/leaanthony/clir"
)

// HOME is the home directory
var HOME string

// TRASHHOME is the trash bin directory
var TRASHHOME string

func init() {
	HOME = os.Getenv("HOME")
	_, err := os.Stat(HOME + "/.Trash")
	if os.IsNotExist(err) {
		TRASHHOME = HOME + "/.Trash"
		err := os.Mkdir(TRASHHOME, 0755)
		dealError(err, TRASHHOME)
	}
	TRASHHOME = HOME + "/.Trash"
}

func main() {

	var delTarget string
	var undelTarget string
	var list bool
	var clear bool

	cli := clir.NewCli("trash", "A trash bin for linux", "v0.0.1")
	cli.StringFlag("r", "Delete the files or directorys", &delTarget)
	cli.StringFlag("u", "Undelete the files or directorys", &undelTarget)
	cli.BoolFlag("l", "List the trash bin", &list)
	cli.BoolFlag("c", "Clear the trash bin", &clear)

	cli.Action(func() error {
		if delTarget != "" {
			Del(cli, delTarget)
			return nil
		}

		if undelTarget != "" {
			// TODO Recover the target or recover the latest deleted target.
			// ReadDir 默认按文件名排序，得改成按时间排序
			// 建议在 ~/.Trash 中维护一个文件 ~/.Trash/.deleted
			// 删除时在文件中添加一条删除记录，恢复时去文件中读取最后一条记录，执行恢复，然后删除最后一条记录
			// 文件中最好在删除时记录下原本的位置，以便恢复时恢复到原来的位置
			//
			// 恢复到原来的位置可以作为另外的选项，如 -u 默认恢复到当前目录，-uo恢复到原来的位置
			// 恢复时注意检查会不会覆盖影响
			//
			// 另外可以提供 ./trash -un [num] 参数，默认恢复一个文件的到当前目录，写了num 则恢复 num 个，～/.Trash 中不足num个文件时，以实际文件数为准

			// os.file.ReadDir reads the directory named by dirname and returns
			// a list of directory entries sorted by filename.
			// func ReadDir(dirname string) ([]os.FileInfo, error) {
			//     f, err := os.Open(dirname)
			//     if err != nil {
			//         return nil, err
			//     }
			//     list, err := f.Readdir(-1)
			//     f.Close()
			//     if err != nil {
			//         return nil, err
			//     }
			//     sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
			//     return list, nil
			// }
			// 排序部分的 sort.Slice 可以改成按修改时间排序

			// trashList, _ := ioutil.ReadDir(TRASHHOME)
			// undelTarget := trashList[len(trashList)-1].Name()

			oldPath := path.Join(TRASHHOME, undelTarget)
			newPath := path.Join("./", undelTarget)

			err := os.Rename(oldPath, newPath)
			dealError(err)
			if len(cli.OtherArgs()) <= 0 {
				return nil
			}
		}

		if list {
			fmt.Println("list trash bin")
			files, err := ioutil.ReadDir(TRASHHOME)
			dealError(err)
			for i, f := range files {
				fmt.Println(i, f.Mode(), f.Size(), f.ModTime().Format("01-02 15:04"), f.Name())
			}
			fmt.Printf("[Count for %d]\n", len(files))
		}

		if clear {
			var choice string
			fmt.Printf("This operation can't recover Clear Trash? [y|N]")
			fmt.Scanln(&choice)
			if choice[0] != 'y' && choice[0] != 'Y' {
				return nil
			}
			os.RemoveAll(TRASHHOME + "/")
		}
		return nil
	})
	cli.Run()
}

// Del deal with the delete feature
func Del(cli *clir.Cli, delTarget string) {
	del(delTarget)
	if len(cli.OtherArgs()) <= 0 {
		return
	}
	for _, f := range cli.OtherArgs() {
		del(f)
	}
}

func del(delTarget string) {
	pwd, _ := os.Getwd()
	var oldpath string
	var newpath string
	oldpath = path.Join(pwd, delTarget)
	newpath = path.Join(TRASHHOME, path.Base(delTarget))
	err := os.Rename(oldpath, newpath)
	dealError(err)
}

func dealError(err error, v ...interface{}) {
	switch err.(type) {
	case *os.PathError:
		fmt.Println("PathError", err)
	case *os.LinkError:
		fmt.Println("LinkError", err)
	}
}
