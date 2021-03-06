// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package actions

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/sylabs/singularity/e2e/internal/e2e"
	"github.com/sylabs/singularity/internal/pkg/test/exec"
)

type testingEnv struct {
	// base env for running tests
	CmdPath     string `split_words:"true"`
	TestDir     string `split_words:"true"`
	RunDisabled bool   `default:"false"`
	//  base image for tests
	ImagePath string `split_words:"true"`
}

var testenv testingEnv

// run tests min fuctionality for singularity run
func actionRun(t *testing.T) {
	tests := []struct {
		name string
		argv []string
		exit int
	}{
		{
			name: "NoCommand",
			argv: []string{testenv.ImagePath},
			exit: 0,
		},
		{
			name: "True",
			argv: []string{testenv.ImagePath, "true"},
			exit: 0,
		},
		{
			name: "False",
			argv: []string{testenv.ImagePath, "false"},
			exit: 1,
		},
		{
			name: "ScifTestAppGood",
			argv: []string{"--app", "testapp", testenv.ImagePath},
			exit: 0,
		},
		{
			name: "ScifTestAppBad",
			argv: []string{"--app", "fakeapp", testenv.ImagePath},
			exit: 1,
		},
	}

	for _, tt := range tests {
		e2e.RunSingularity(
			t,
			tt.name,
			e2e.WithCommand("run"),
			e2e.WithArgs(tt.argv...),
			e2e.ExpectExit(tt.exit),
		)
	}
}

// exec tests min fuctionality for singularity exec
func actionExec(t *testing.T) {
	user := e2e.CurrentUser(t)

	// Create a temp testfile
	tmpfile, err := ioutil.TempFile("", "testSingularityExec.tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	testfile, err := tmpfile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		argv []string
		exit int
	}{
		{
			name: "NoCommand",
			argv: []string{testenv.ImagePath},
			exit: 1,
		},
		{
			name: "True",
			argv: []string{testenv.ImagePath, "true"},
			exit: 0,
		},
		{
			name: "TrueAbsPAth",
			argv: []string{testenv.ImagePath, "/bin/true"},
			exit: 0,
		},
		{
			name: "False",
			argv: []string{testenv.ImagePath, "false"},
			exit: 1,
		},
		{
			name: "FalseAbsPath",
			argv: []string{testenv.ImagePath, "/bin/false"},
			exit: 1,
		},
		// Scif apps tests
		{
			name: "ScifTestAppGood",
			argv: []string{"--app", "testapp", testenv.ImagePath, "testapp.sh"},
			exit: 0,
		},
		{
			name: "ScifTestAppBad",
			argv: []string{"--app", "fakeapp", testenv.ImagePath, "testapp.sh"},
			exit: 1,
		},
		{
			name: "ScifTestfolderOrg",
			argv: []string{testenv.ImagePath, "test", "-d", "/scif"},
			exit: 0,
		},
		{
			name: "ScifTestfolderOrg",
			argv: []string{testenv.ImagePath, "test", "-d", "/scif/apps"},
			exit: 0,
		},
		{
			name: "ScifTestfolderOrg",
			argv: []string{testenv.ImagePath, "test", "-d", "/scif/data"},
			exit: 0,
		},
		{
			name: "ScifTestfolderOrg",
			argv: []string{testenv.ImagePath, "test", "-d", "/scif/apps/foo"},
			exit: 0,
		},
		{
			name: "ScifTestfolderOrg",
			argv: []string{testenv.ImagePath, "test", "-d", "/scif/apps/bar"},
			exit: 0,
		},
		// blocked by issue [scif-apps] Files created at install step fall into an unexpected path #2404
		{
			name: "ScifTestfolderOrg",
			argv: []string{testenv.ImagePath, "test", "-f", "/scif/apps/foo/filefoo.exec"},
			exit: 0,
		},
		{
			name: "ScifTestfolderOrg",
			argv: []string{testenv.ImagePath, "test", "-f", "/scif/apps/bar/filebar.exec"},
			exit: 0,
		},
		{
			name: "ScifTestfolderOrg",
			argv: []string{testenv.ImagePath, "test", "-d", "/scif/data/foo/output"},
			exit: 0,
		},
		{
			name: "ScifTestfolderOrg",
			argv: []string{testenv.ImagePath, "test", "-d", "/scif/data/foo/input"},
			exit: 0,
		},
		{
			name: "WorkdirContain",
			argv: []string{"--contain", testenv.ImagePath, "test", "-f", tmpfile.Name()},
			exit: 1,
		},
		{
			name: "Workdir",
			argv: []string{"--workdir", "testdata", testenv.ImagePath, "test", "-f", tmpfile.Name()},
			exit: 0,
		},
		{
			name: "PwdGood",
			argv: []string{"--pwd", "/etc", testenv.ImagePath, "true"},
			exit: 0,
		},
		{
			name: "Home",
			argv: []string{"--home", pwd + "testdata", testenv.ImagePath, "test", "-f", tmpfile.Name()},
			exit: 0,
		},
		{
			name: "HomePath",
			argv: []string{"--home", "/tmp:/home", testenv.ImagePath, "test", "-f", "/home/" + testfile.Name()},
			exit: 0,
		},
		{
			name: "HomeTmp",
			argv: []string{"--home", "/tmp", testenv.ImagePath, "true"},
			exit: 0,
		},
		{
			name: "HomeTmpExplicit",
			argv: []string{"--home", "/tmp:/home", testenv.ImagePath, "true"},
			exit: 0,
		},
		{
			name: "UserBind",
			argv: []string{"--bind", "/tmp:/var/tmp", testenv.ImagePath, "test", "-f", "/var/tmp/" + testfile.Name()},
			exit: 0,
		},
		{
			name: "NoHome",
			argv: []string{"--no-home", testenv.ImagePath, "ls", "-ld", user.Dir},
			exit: 2,
		},
	}

	for _, tt := range tests {
		e2e.RunSingularity(
			t,
			tt.name,
			e2e.WithCommand("exec"),
			e2e.WithDir("/tmp"),
			e2e.WithArgs(tt.argv...),
			e2e.ExpectExit(tt.exit),
		)
	}
}

// Shell interaction tests
func actionShell(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
		t.Fatalf("could not get hostname: %s", err)
	}

	tests := []struct {
		name       string
		argv       []string
		consoleOps []e2e.SingularityConsoleOp
		exit       int
	}{
		{
			name: "ShellExit",
			argv: []string{testenv.ImagePath},
			consoleOps: []e2e.SingularityConsoleOp{
				e2e.ConsoleExpect("Singularity"),
				e2e.ConsoleSendLine("exit"),
			},
			exit: 0,
		},
		{
			name: "ShellHostname",
			argv: []string{testenv.ImagePath},
			consoleOps: []e2e.SingularityConsoleOp{
				e2e.ConsoleSendLine("hostname"),
				e2e.ConsoleExpect(hostname),
				e2e.ConsoleSendLine("exit"),
			},
			exit: 0,
		},
		{
			name: "ShellBadCommand",
			argv: []string{testenv.ImagePath},
			consoleOps: []e2e.SingularityConsoleOp{
				e2e.ConsoleSendLine("_a_fake_command"),
				e2e.ConsoleSendLine("exit"),
			},
			exit: 127,
		},
	}

	for _, tt := range tests {
		e2e.RunSingularity(
			t,
			tt.name,
			e2e.WithCommand("shell"),
			e2e.WithArgs(tt.argv...),
			e2e.ConsoleRun(tt.consoleOps...),
			e2e.ExpectExit(tt.exit),
		)
	}
}

// STDPipe tests pipe stdin/stdout to singularity actions cmd
func STDPipe(t *testing.T) {
	stdinTests := []struct {
		name    string
		command string
		argv    []string
		input   string
		exit    int
	}{
		{
			name:    "TrueSTDIN",
			command: "exec",
			argv:    []string{testenv.ImagePath, "grep", "hi"},
			input:   "hi",
			exit:    0,
		},
		{
			name:    "FalseSTDIN",
			command: "exec",
			argv:    []string{testenv.ImagePath, "grep", "hi"},
			input:   "bye",
			exit:    1,
		},
		{
			name:    "TrueLibrary",
			command: "shell",
			argv:    []string{"library://busybox"},
			input:   "true",
			exit:    0,
		},
		{
			name:    "FalseLibrary",
			command: "shell",
			argv:    []string{"library://busybox"},
			input:   "false",
			exit:    1,
		},
		{
			name:    "TrueDocker",
			command: "shell",
			argv:    []string{"docker://busybox"},
			input:   "true",
			exit:    0,
		},
		{
			name:    "FalseDocker",
			command: "shell",
			argv:    []string{"docker://busybox"},
			input:   "false",
			exit:    1,
		},
		{
			name:    "TrueShub",
			command: "shell",
			argv:    []string{"shub://singularityhub/busybox"},
			input:   "true",
			exit:    0,
		},
		{
			name:    "FalseShub",
			command: "shell",
			argv:    []string{"shub://singularityhub/busybox"},
			input:   "false",
			exit:    1,
		},
	}

	var input bytes.Buffer

	for _, tt := range stdinTests {
		e2e.RunSingularity(
			t,
			tt.name,
			e2e.WithCommand(tt.command),
			e2e.WithArgs(tt.argv...),
			e2e.WithStdin(&input),
			e2e.PreRun(func(t *testing.T) {
				input.WriteString(tt.input)
			}),
			e2e.ExpectExit(tt.exit),
		)
		input.Reset()
	}

	user := e2e.CurrentUser(t)
	stdoutTests := []struct {
		name    string
		command string
		argv    []string
		output  string
		exit    int
	}{
		{
			name:    "AppsFoo",
			command: "run",
			argv:    []string{"--app", "foo", testenv.ImagePath},
			output:  "FOO",
			exit:    0,
		},
		{
			name:    "PwdPath",
			command: "exec",
			argv:    []string{"--pwd", "/etc", testenv.ImagePath, "pwd"},
			output:  "/etc",
			exit:    0,
		},
		{
			name:    "Arguments",
			command: "run",
			argv:    []string{testenv.ImagePath, "foo"},
			output:  "foo",
			exit:    127,
		},
		{
			name:    "Permissions",
			command: "exec",
			argv:    []string{testenv.ImagePath, "id", "-un"},
			output:  user.Name,
			exit:    0,
		},
	}
	for _, tt := range stdoutTests {
		e2e.RunSingularity(
			t,
			tt.name,
			e2e.WithCommand(tt.command),
			e2e.WithArgs(tt.argv...),
			e2e.ExpectExit(tt.exit, e2e.ExpectOutput(tt.output)),
		)
	}
}

// RunFromURI tests min fuctionality for singularity run/exec URI://
func RunFromURI(t *testing.T) {
	runScript := "testdata/runscript.sh"
	bind := fmt.Sprintf("%s:/.singularity.d/runscript", runScript)

	fi, err := os.Stat(runScript)
	if err != nil {
		t.Fatalf("can't find %s", runScript)
	}
	size := strconv.Itoa(int(fi.Size()))

	tests := []struct {
		name    string
		command string
		argv    []string
		exit    int
	}{
		// Run from supported URI's and check the runscript call works
		{
			name:    "RunFromDockerOK",
			command: "run",
			argv:    []string{"--bind", bind, "docker://busybox:latest", size},
			exit:    0,
		},
		{
			name:    "RunFromLibraryOK",
			command: "run",
			argv:    []string{"--bind", bind, "library://busybox:latest", size},
			exit:    0,
		},
		{
			name:    "RunFromShubOK",
			command: "run",
			argv:    []string{"--bind", bind, "shub://singularityhub/busybox", size},
			exit:    0,
		},
		{
			name:    "RunFromOrasOK",
			command: "run",
			argv:    []string{"--bind", bind, "oras://localhost:5000/oras_test_sif:latest", size},
			exit:    0,
		},
		{
			name:    "RunFromDockerKO",
			command: "run",
			argv:    []string{"--bind", bind, "docker://busybox:latest", "0"},
			exit:    1,
		},
		{
			name:    "RunFromLibraryKO",
			command: "run",
			argv:    []string{"--bind", bind, "library://busybox:latest", "0"},
			exit:    1,
		},
		{
			name:    "RunFromShubKO",
			command: "run",
			argv:    []string{"--bind", bind, "shub://singularityhub/busybox", "0"},
			exit:    1,
		},
		{
			name:    "RunFromOrasKO",
			command: "run",
			argv:    []string{"--bind", bind, "oras://localhost:5000/oras_test_sif:latest", "0"},
			exit:    1,
		},

		// exec from a supported URI's and check the exit code
		{
			name:    "ExecTrueDocker",
			command: "exec",
			argv:    []string{"docker://busybox:latest", "true"},
			exit:    0,
		},
		{
			name:    "ExecTrueLibrary",
			command: "exec",
			argv:    []string{"library://busybox:latest", "true"},
			exit:    0,
		},
		{
			name:    "ExecTrueShub",
			command: "exec",
			argv:    []string{"shub://singularityhub/busybox", "true"},
			exit:    0,
		},
		{
			name:    "ExecTrueOras",
			command: "exec",
			argv:    []string{"oras://localhost:5000/oras_test_sif:latest", "true"},
			exit:    0,
		},
		{
			name:    "ExecFalseDocker",
			command: "exec",
			argv:    []string{"docker://busybox:latest", "false"},
			exit:    1,
		},
		{
			name:    "ExecFalseLibrary",
			command: "exec",
			argv:    []string{"library://busybox:latest", "false"},
			exit:    1,
		},
		{
			name:    "ExecFalseShub",
			command: "exec",
			argv:    []string{"shub://singularityhub/busybox", "false"},
			exit:    1,
		},
		{
			name:    "ExecFalseOras",
			command: "exec",
			argv:    []string{"oras://localhost:5000/oras_test_sif:latest", "false"},
			exit:    1,
		},

		// exec from URI with user namespace enabled
		{
			name:    "ExecTrueDockerUserns",
			command: "exec",
			argv:    []string{"--userns", "docker://busybox:latest", "true"},
			exit:    0,
		},
		{
			name:    "ExecTrueLibraryUserns",
			command: "exec",
			argv:    []string{"--userns", "library://busybox:latest", "true"},
			exit:    0,
		},
		{
			name:    "ExecTrueShubUserns",
			command: "exec",
			argv:    []string{"--userns", "shub://singularityhub/busybox", "true"},
			exit:    0,
		},
		{
			name:    "ExecTrueOrasUserns",
			command: "exec",
			argv:    []string{"--userns", "oras://localhost:5000/oras_test_sif:latest", "true"},
			exit:    0,
		},
		{
			name:    "ExecFalseDockerUserns",
			command: "exec",
			argv:    []string{"--userns", "docker://busybox:latest", "false"},
			exit:    1,
		},
		{
			name:    "ExecFalseLibraryUserns",
			command: "exec",
			argv:    []string{"--userns", "library://busybox:latest", "false"},
			exit:    1,
		},
		{
			name:    "ExecFalseShubUserns",
			command: "exec",
			argv:    []string{"--userns", "shub://singularityhub/busybox", "false"},
			exit:    1,
		},
		{
			name:    "ExecFalseOrasUserns",
			command: "exec",
			argv:    []string{"--userns", "oras://localhost:5000/oras_test_sif:latest", "false"},
			exit:    1,
		},
	}

	for _, tt := range tests {
		e2e.RunSingularity(
			t,
			tt.name,
			e2e.WithCommand(tt.command),
			e2e.WithArgs(tt.argv...),
			e2e.ExpectExit(tt.exit),
		)
	}
}

// PersistentOverlay test the --overlay function
func PersistentOverlay(t *testing.T) {
	const squashfsImage = "squashfs.simg"

	dir, err := ioutil.TempDir(testenv.TestDir, "overlay_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create dirfs for squashfs
	squashDir, err := ioutil.TempDir(testenv.TestDir, "overlay_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(squashDir)

	content := []byte("temporary file's content")
	tmpfile, err := ioutil.TempFile(squashDir, "bogus")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	cmd := exec.Command("mksquashfs", squashDir, squashfsImage, "-noappend", "-all-root")
	if res := cmd.Run(t); res.Error != nil {
		t.Fatalf("Unexpected error while running command.\n%s", res)
	}
	defer os.RemoveAll(squashfsImage)

	//  Create the overlay ext3 fs
	cmd = exec.Command("dd", "if=/dev/zero", "of=ext3_fs.img", "bs=1M", "count=768", "status=none")
	if res := cmd.Run(t); res.Error != nil {
		t.Fatalf("Unexpected error while running command.\n%s", res)
	}

	cmd = exec.Command("mkfs.ext3", "-q", "-F", "ext3_fs.img")
	if res := cmd.Run(t); res.Error != nil {
		t.Fatalf("Unexpected error while running command.\n%s", res)
	}

	defer os.Remove("ext3_fs.img")

	tests := []struct {
		name       string
		argv       []string
		exit       int
		privileged bool
	}{
		{
			name:       "overlay_create",
			argv:       []string{"--overlay", dir, testenv.ImagePath, "touch", "/dir_overlay"},
			exit:       0,
			privileged: true,
		},
		{
			name:       "overlay_find",
			argv:       []string{"--overlay", dir, testenv.ImagePath, "test", "-f", "/dir_overlay"},
			exit:       0,
			privileged: true,
		},
		{
			name:       "overlay_ext3_create",
			argv:       []string{"--overlay", "ext3_fs.img", testenv.ImagePath, "touch", "/ext3_overlay"},
			exit:       0,
			privileged: true,
		},
		{
			name:       "overlay_ext3_find",
			argv:       []string{"--overlay", "ext3_fs.img", testenv.ImagePath, "test", "-f", "/ext3_overlay"},
			exit:       0,
			privileged: true,
		},
		{
			name:       "overlay_squashFS_find",
			argv:       []string{"--overlay", squashfsImage, testenv.ImagePath, "test", "-f", fmt.Sprintf("/%s", tmpfile.Name())},
			exit:       0,
			privileged: true,
		},
		{
			name:       "overlay_multiple_create",
			argv:       []string{"--overlay", "ext3_fs.img", "--overlay", squashfsImage, testenv.ImagePath, "touch", "/multiple_overlay_fs"},
			exit:       0,
			privileged: true,
		},
		{
			name:       "overlay_multiple_find_ext3",
			argv:       []string{"--overlay", "ext3_fs.img", "--overlay", squashfsImage, testenv.ImagePath, "test", "-f", "/multiple_overlay_fs"},
			exit:       0,
			privileged: true,
		},
		{
			name:       "overlay_multiple_find_squashfs",
			argv:       []string{"--overlay", "ext3_fs.img", "--overlay", squashfsImage, testenv.ImagePath, "test", "-f", fmt.Sprintf("/%s", tmpfile.Name())},
			exit:       0,
			privileged: true,
		},
		{
			name:       "overlay_noroot",
			argv:       []string{"--overlay", dir, testenv.ImagePath, "test", "-f", "/foo_overlay"},
			exit:       255,
			privileged: false,
		},
		{
			name:       "overlay_noflag",
			argv:       []string{testenv.ImagePath, "test", "-f", "/foo_overlay"},
			exit:       1,
			privileged: true,
		},
	}

	for _, tt := range tests {
		e2e.RunSingularity(
			t,
			tt.name,
			e2e.WithCommand("exec"),
			e2e.WithArgs(tt.argv...),
			e2e.WithPrivileges(tt.privileged),
			e2e.ExpectExit(tt.exit),
		)
	}
}

// RunE2ETests is the main func to trigger the test suite
func RunE2ETests(t *testing.T) {
	e2e.LoadEnv(t, &testenv)

	// singularity run
	t.Run("run", actionRun)
	// singularity exec
	t.Run("exec", actionExec)
	// stdin/stdout pipe
	t.Run("STDPIPE", STDPipe)
	// action_URI
	t.Run("action_URI", RunFromURI)
	// Persistent Overlay
	t.Run("Persistent_Overlay", PersistentOverlay)
	// shell interaction
	t.Run("Shell", actionShell)
}
