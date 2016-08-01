package docker

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

func TestBuildImageErrorStream(t *testing.T) {
	d, err := NewClientFromEnv()
	if err != nil {
		t.Error(err)
	}

	dockerfile := `FROM busybox:latest
RUN echo "Regular output message"
RUN echo "This line is printed to standard error" >&2
RUN echo "xxxx" >&2 && exit 1
`

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	header := tar.Header{
		Name: "Dockerfile",
		Mode: 0600,
		Size: int64(len(dockerfile)),
	}

	if err := tw.WriteHeader(&header); err != nil {
		t.Error(err)
	}

	if _, err := tw.Write([]byte(dockerfile)); err != nil {
		t.Error(err)
	}

	if err := tw.Close(); err != nil {
		t.Error(err)
	}

	buildContext := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	outReader, outWriter := io.Pipe()
	defer outWriter.Close()
	errReader, errWriter := io.Pipe()
	defer errWriter.Close()
	opts := BuildImageOptions{
		Name:         "test",
		InputStream:  buildContext,
		NoCache:      true,
		OutputStream: outWriter,
		// ErrorStream:  errWriter,
		RawJSONStream: true,
		Pull:          true,
	}

	go func() {
		outScanner := bufio.NewScanner(outReader)
		for outScanner.Scan() {
			fmt.Printf("stdout: %s\n", outScanner.Text())
		}
		errScanner := bufio.NewScanner(errReader)
		for errScanner.Scan() {
			fmt.Printf("stderr: %s\n", errScanner.Text())
		}
	}()

	fmt.Print("build!\n")
	err = d.BuildImage(opts)
	fmt.Print("build done!\n")

	if err != nil {
		t.Error(err)
	}
}
