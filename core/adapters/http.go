package adapters

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/smartcontractkit/chainlink/core/store"
	"github.com/smartcontractkit/chainlink/core/store/models"
)

// HTTPGet requires a URL which is used for a GET request when the adapter is called.
type HTTPGet struct {
	URL     models.WebURL `json:"url"`
	GET     models.WebURL `json:"get"`
	Headers http.Header   `json:"headers"`
}

// Perform ensures that the adapter's URL responds to a GET request without
// errors and returns the response body as the "value" field of the result.
func (hga *HTTPGet) Perform(input models.RunResult, store *store.Store) models.RunResult {
	request, err := http.NewRequest("GET", hga.GetURL(), nil)
	if err != nil {
		input.SetError(err)
		return input
	}
	setHeaders(request, hga.Headers, "")
	return sendRequest(input, request, store.Config.DefaultHTTPLimit())
}

// GetURL retrieves the GET field if set otherwise returns the URL field
func (hga *HTTPGet) GetURL() string {
	if hga.GET.String() != "" {
		return hga.GET.String()
	}
	return hga.URL.String()
}

// HTTPPost requires a URL which is used for a POST request when the adapter is called.
type HTTPPost struct {
	URL     models.WebURL `json:"url"`
	POST    models.WebURL `json:"post"`
	Headers http.Header   `json:"headers"`
}

// Perform ensures that the adapter's URL responds to a POST request without
// errors and returns the response body as the "value" field of the result.
func (hpa *HTTPPost) Perform(input models.RunResult, store *store.Store) models.RunResult {
	reqBody := bytes.NewBufferString(input.Data.String())
	request, err := http.NewRequest("POST", hpa.GetURL(), reqBody)
	if err != nil {
		input.SetError(err)
		return input
	}
	setHeaders(request, hpa.Headers, "application/json")
	return sendRequest(input, request, store.Config.DefaultHTTPLimit())
}

// GetURL retrieves the POST field if set otherwise returns the URL field
func (hpa *HTTPPost) GetURL() string {
	if hpa.POST.String() != "" {
		return hpa.POST.String()
	}
	return hpa.URL.String()
}

func setHeaders(request *http.Request, headers http.Header, contentType string) {
	if headers != nil {
		request.Header = headers
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
}

func sendRequest(input models.RunResult, request *http.Request, limit int64) models.RunResult {
	tr := &http.Transport{
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(request)
	if err != nil {
		input.SetError(err)
		return input
	}

	defer response.Body.Close()

	source := newMaxBytesReader(response.Body, limit)
	bytes, err := ioutil.ReadAll(source)
	if err != nil {
		input.SetError(err)
		return input
	}

	responseBody := string(bytes)
	if response.StatusCode >= 400 {
		input.SetError(errors.New(responseBody))
		return input
	}

	input.ApplyResult(responseBody)
	return input
}

// maxBytesReader is inspired by
// https://github.com/gin-contrib/size/blob/master/size.go
type maxBytesReader struct {
	rc               io.ReadCloser
	limit, remaining int64
	sawEOF           bool
}

func newMaxBytesReader(rc io.ReadCloser, limit int64) *maxBytesReader {
	return &maxBytesReader{
		rc:        rc,
		limit:     limit,
		remaining: limit,
	}
}

func (mbr *maxBytesReader) Read(p []byte) (n int, err error) {
	toRead := mbr.remaining
	if mbr.remaining == 0 {
		if mbr.sawEOF {
			return mbr.tooLarge()
		}
		// The underlying io.Reader may not return (0, io.EOF)
		// at EOF if the requested size is 0, so read 1 byte
		// instead. The io.Reader docs are a bit ambiguous
		// about the return value of Read when 0 bytes are
		// requested, and {bytes,strings}.Reader gets it wrong
		// too (it returns (0, nil) even at EOF).
		toRead = 1
	}
	if int64(len(p)) > toRead {
		p = p[:toRead]
	}
	n, err = mbr.rc.Read(p)
	if err == io.EOF {
		mbr.sawEOF = true
	}
	if mbr.remaining == 0 {
		// If we had zero bytes to read remaining (but hadn't seen EOF)
		// and we get a byte here, that means we went over our limit.
		if n > 0 {
			return mbr.tooLarge()
		}
		return 0, err
	}
	mbr.remaining -= int64(n)
	if mbr.remaining < 0 {
		mbr.remaining = 0
	}
	return
}

func (mbr *maxBytesReader) tooLarge() (int, error) {
	return 0, fmt.Errorf("HTTP request too large, must be less than %d bytes", mbr.limit)
}

func (mbr *maxBytesReader) Close() error {
	return mbr.rc.Close()
}
