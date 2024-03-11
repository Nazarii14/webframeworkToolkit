package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var tools Tools

	s := tools.RandomString(10)
	if len(s) != 10 {
		t.Errorf("RandomString() = %v; want 10", len(s))
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	{name: "allowed no rename",
		allowedTypes:  []string{"image/png", "image/jpeg"},
		renameFile:    false,
		errorExpected: false,
	},
	{name: "allowed rename",
		allowedTypes:  []string{"image/png", "image/jpeg"},
		renameFile:    true,
		errorExpected: false,
	},
	{name: "not allowed",
		allowedTypes:  []string{"image/jpeg"},
		renameFile:    false,
		errorExpected: true,
	},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", "./testdata/test.png")
			if err != nil {
				t.Errorf("CreateFormFile() error = %v; want nil", err)
			}

			f, err := os.Open("./testdata/test.png")

			if err != nil {
				t.Errorf("os.Open() error = %v; want nil", err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Errorf("image.Decode() error = %v; want nil", err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Errorf("png.Encode() error = %v; want nil", err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
			}

			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error expected but none received", e.name)
		}

		wg.Wait()
	}
}

func TestTools_UploadFile(t *testing.T) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer writer.Close()

		part, err := writer.CreateFormFile("file", "./testdata/test.png")
		if err != nil {
			t.Errorf("CreateFormFile() error = %v; want nil", err)
		}

		f, err := os.Open("./testdata/test.png")

		if err != nil {
			t.Errorf("os.Open() error = %v; want nil", err)
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			t.Errorf("image.Decode() error = %v; want nil", err)
		}

		err = png.Encode(part, img)
		if err != nil {
			t.Errorf("png.Encode() error = %v; want nil", err)
		}
	}()

	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	var testTools Tools

	uploadedFiles, err := testTools.UploadOneFile(request, "./testdata/uploads/", true)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName)); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", err.Error())
	}

	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName))
}

func TestTools_CreateDirIfNotExists(t *testing.T) {
	var testTool Tools

	err := testTool.CreateDirIfNotExists("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

	err = testTool.CreateDirIfNotExists("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

	_ = os.Remove("./testdata/myDir")
}

var slugsTests = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{name: "valid string", s: "now is the time 123", expected: "now-is-the-time-123", errorExpected: false},
	{name: "empty string", s: "", expected: "", errorExpected: true},
	{name: "complex string", s: "Complex String!@#$1234", expected: "complex-string-1234", errorExpected: false},
}

func TestTools_Slugify(t *testing.T) {
	var testTools Tools

	for _, e := range slugsTests {
		s, err := testTools.Slugify(e.s)
		if err != nil && !e.errorExpected {
			t.Errorf("%s error not expected; got %v", e.name, err)
		}

		if !e.errorExpected && s != e.expected {
			t.Errorf("expected %s; got %s", e.expected, s)
		}
	}
}

func TestTools_DownloadStaticFile(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	var testTool Tools
	testTool.DownloadStaticFile(rr, req, "./testdata", "pic.jpg", "pic1.jpg")
	res := rr.Result()
	defer res.Body.Close()

	if res.Header["Content-Length"][0] != "100" {
		t.Error("wrong content length of", res.Header["Content-Length"][0])
	}
}
