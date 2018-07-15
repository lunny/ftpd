package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/goftp/file-driver"
	"github.com/goftp/leveldb-auth"
	"github.com/goftp/server"
	"github.com/lunny/log"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	testPORT = 4242
)

var (
	ftpServer  *server.Server
	rootFolder string
	auth       *ldbauth.LDBAuth
)

func TestMain(m *testing.M) {
	//Setup
	startFTPServer()
	//Do tests
	retVal := m.Run()
	//Clean up
	clean()
	//Return
	os.Exit(retVal)
}

func startFTPServer() {
	f, err := ioutil.TempDir("", "ftpd-")
	if err != nil {
		log.Fatal(err)
	}
	rootFolder = f

	log.Printf("Starting ftp server at : %s", rootFolder)
	db, err := leveldb.OpenFile(filepath.Join(rootFolder, "authperm.db"), nil)
	if err != nil {
		log.Fatal(err)
	}
	auth = &ldbauth.LDBAuth{db}
	factory := &filedriver.FileDriverFactory{
		rootFolder,
		server.NewSimplePerm("root", "root"),
	}

	opt := &server.ServerOpts{
		Name:    "Go Ftp Server : Integration Testing",
		Factory: factory,
		Port:    testPORT,
		Auth:    auth,
	}
	// start ftp server
	ftpServer = server.NewServer(opt)
	log.Info("FTP Server", version)
	go func() {
		err = ftpServer.ListenAndServe()
		if err != nil && err.Error() != "Server closed" {
			log.Fatal("Error starting server:", err)
		}
	}()
}

func clean() {
	ftpServer.Shutdown()
	err := os.RemoveAll(rootFolder)
	if err != nil {
		log.Fatal("Error cleaning up:", err)
	}
}

// TestRclone runs the ftp server then runs the tests from Rclone against it.
func TestRclone(t *testing.T) {
	if testing.Short() { //Allow to not be run when short
		t.Skip("skipping test in short mode.")
	}
	t.Log("Init Rclone integration tests")
	auth.AddUser("rclone", "password")
	t.Log("Start Rclone integration tests")

	// Run the ftp tests with an on the fly remote
	args := []string{"test"}
	if testing.Verbose() {
		args = append(args, "-v")
	}

	//Retrieve rclone
	err := exec.Command("go", "get", "-u", "-v", "github.com/ncw/rclone").Run()
	if err != nil {
		t.Fatal(err)
	}
	//Get gopath
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}

	args = append(args, "-remote", "ftptest:")
	cmd := exec.Command("go", args...)
	cmd.Dir = filepath.Join(gopath, "src/github.com/ncw/rclone/backend/ftp")
	cmd.Env = append(os.Environ(),
		"RCLONE_CONFIG_FTPTEST_TYPE=ftp",
		"RCLONE_CONFIG_FTPTEST_HOST=127.0.0.1",
		fmt.Sprintf("RCLONE_CONFIG_FTPTEST_PORT=%d", testPORT),
		"RCLONE_CONFIG_FTPTEST_USER=rclone",
		"RCLONE_CONFIG_FTPTEST_PASS=0HU5Hx42YiLoNGJxppOOP3QTbr-KB_MP", // ./rclone obscure password
	)
	out, err := cmd.CombinedOutput()
	if len(out) != 0 {
		t.Logf("\n----------\n%s----------\n", string(out))
	}
	assert.NoError(t, err, "Running ftp integration tests")

	t.Log("Finish Rclone integration tests")
}
