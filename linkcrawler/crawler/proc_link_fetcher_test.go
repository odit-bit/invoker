package crawler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	mock_crawler "github.com/odit-bit/invoker/linkcrawler/mocks"
	"go.uber.org/mock/gomock"
)

var successHttpResponse = func(code int, header string, body []byte) (*http.Response, error) {

	res := httptest.NewRecorder()
	res.Code = code
	res.Header().Set("Content-type", header)
	res.Body = bytes.NewBuffer(body)
	return res.Result(), nil
}

func Test_linkFetcher_error_httpResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlGetter := mock_crawler.NewMockURLGetter(ctrl)
	pnd := mock_crawler.NewMockPrivateNetworkDetector(ctrl)

	// http response status code not 200
	var inputURL = "http://example.com"
	var privateNetwork = false
	urlGetter.
		EXPECT().Get(gomock.Eq(inputURL)).AnyTimes().
		Return(successHttpResponse(404, "html/utf-8", []byte("EXAMPLE CONTENT")))

	pnd.
		EXPECT().IsPrivate(gomock.Any()).AnyTimes().
		Return(privateNetwork, nil)

	p := &payload{URL: inputURL}
	res, err := newLinkFetcher(urlGetter, pnd).Process(context.TODO(), p)

	if err != nil {
		t.Error(err)
	}
	if res != nil {
		t.Error("result should nil", res)
	}

	// non html header
	inputURL = "https://www.nonHtml.com"
	urlGetter.
		EXPECT().Get(gomock.Eq(inputURL)).AnyTimes().
		Return(successHttpResponse(200, "Application/JSON", []byte(`{"EXAMPLE":"CONTENT}"`)))

	p = &payload{URL: inputURL}
	res, err = newLinkFetcher(urlGetter, pnd).Process(context.TODO(), p)

	if err != nil {
		t.Error(err)
	}

	if res != nil {
		t.Error("result should nil", res)
	}

}

func Test_linkFetcher_exclusion_url(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlGetter := mock_crawler.NewMockURLGetter(ctrl)
	pnd := mock_crawler.NewMockPrivateNetworkDetector(ctrl)

	var inputURL = "https://www.example.com/image.png"
	urlGetter.
		EXPECT().Get(gomock.Eq(inputURL)).AnyTimes().
		Return(successHttpResponse(200, "image/", []byte("EXAMPLE CONTENT")))

	pnd.
		EXPECT().IsPrivate(gomock.Any()).AnyTimes().
		Return(false, nil)

	//  error and payload should nil

	p := &payload{URL: "http://example.com/foo.png"}
	res, err := newLinkFetcher(urlGetter, pnd).Process(context.TODO(), p)

	if err != nil {
		t.Error(err)
	}

	if reflect.DeepEqual(res, p) {
		t.Errorf("\n%v\n%v\n", res, p)
	}
}

func Test_linkFetcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlGetter := mock_crawler.NewMockURLGetter(ctrl)
	pnd := mock_crawler.NewMockPrivateNetworkDetector(ctrl)

	htmlContent := []byte(`
		{
			"name":"First",
			"price":123415,
			"tag":[tag1,tag2]
		}
	`)
	urlGetter.EXPECT().Get(gomock.Any()).AnyTimes().
		Return(successHttpResponse(200, "html/utf-8", htmlContent))

	pnd.EXPECT().IsPrivate(gomock.Any()).AnyTimes().
		Return(false, nil)

	p := &payload{URL: "http://example.com/index.html"}
	res, err := newLinkFetcher(urlGetter, pnd).Process(context.TODO(), p)

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(res, p) {
		t.Errorf("\n%v\n%v\n", res, p)
	}

}
