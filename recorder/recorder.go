package recorder

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type recorder interface {
	Add(string)
	Sub() string
	List() string
	Clear() bool
}

// Reporter is responsible for maintaining the record
type Reporter struct {
	trashHome       string
	trashLogPath    string
	trashConfigPath string
}

// NewReporter return a reporter which is responsible for manage the trash log file -- .deleted
func NewReporter(trashHome, trashLogPath, trashConfigPath string) *Reporter {
	return &Reporter{
		trashHome:       trashHome,
		trashLogPath:    trashLogPath,
		trashConfigPath: trashConfigPath,
	}
}

// Add will write down a record into trash log file. It includes permission, absolute path, deletion date.
func (r *Reporter) Add(delTarget string) {
	trashLog, _ := os.OpenFile(r.trashLogPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	defer trashLog.Close()

	// deletion Date
	now := time.Now().Format("2006-1-2 15:04:05")
	// original path
	abs, err := filepath.Abs(delTarget)
	dealError(err)
	// permission and file size
	f, err := os.Stat(delTarget)
	dealError(err)
	perm := f.Mode()
	size := strconv.Itoa(int(f.Size()))

	buf := bufio.NewWriter(trashLog)
	_, err = buf.WriteString(perm.String() + " " + size + " " + now + " " + abs + "\n")
	dealError(err)
	buf.Flush()
}

// Sub deletes the last item in trash log file and retuen the undelTarget absolute path.
// Step:
func (r *Reporter) Sub() string {

	// Open the file
	trashLog, err := os.OpenFile(r.trashLogPath, os.O_RDONLY, os.ModePerm)
	dealError(err)
	defer trashLog.Close()

	// Read all content of trash can log file
	b, err := io.ReadAll(trashLog)
	dealError(err)
	if len(b) > 0 {
		// Split the content by '\n', get every line
		bb := bytes.SplitAfter(b, []byte("\n"))
		// Get last line(bb[len(bb)-2]), extract absolute path, and trim the '\n' of right
		undelPath := bytes.TrimRight(bytes.Split(bb[len(bb)-2], []byte(" "))[4], "\n")
		// Write back the first to last-1 lines
		err = ioutil.WriteFile(r.trashLogPath, bytes.Join(bb[:len(bb)-2], []byte("")), os.ModePerm)
		dealError(err)

		return string(undelPath)
	}
	return ""
}

// List show all records to Stdout,
func (r *Reporter) List() (n int, s string) {
	trashLog, err := os.OpenFile(r.trashLogPath, os.O_RDONLY, os.ModePerm)
	dealError(err)
	defer trashLog.Close()

	b, err := io.ReadAll(trashLog)
	dealError(err)

	return len(b), string(b)
}

// Clear all content from trash log files
func (r *Reporter) Clear() bool {
	if err := os.Truncate(r.trashLogPath, 0); err != nil {
		dealError(err)
		return false
	}
	return true
}

func dealError(err error) {
	switch err.(type) {
	case *os.PathError:
		log.Println(err)
	case *os.LinkError:
		log.Println(err)
	}
}
