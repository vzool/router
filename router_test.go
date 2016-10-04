// Copyright (c) Jeevanandam M (https://github.com/jeevatkm)
// go-aah/aah source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aahframework.org/aah/ahttp"
	"aahframework.org/essentials"
	"aahframework.org/test/assert"
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Test Path Params methods
//___________________________________

func TestPathParamGet(t *testing.T) {
	pathParameters := PathParams{
		PathParam{"dir", "js"},
		PathParam{"filepath", "/inc/framework.js"},
	}

	fp := pathParameters.Get("filepath")
	assert.Equal(t, "/inc/framework.js", fp)

	notfound := pathParameters.Get("notfound")
	assert.Equal(t, "", notfound)
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Test Router methods
//___________________________________

func TestRouterLoadConfiguration(t *testing.T) {
	router := createRouter("routes.conf")

	err := router.Load()
	assert.FailNowOnError(t, err, "")

	// After loading just couple assertion

	reqCancelBooking1 := createHTTPRequest("localhost:8000", "/hotels/12345/cancel")
	reqCancelBooking1.Method = ahttp.MethodPost
	domain := router.Domain(reqCancelBooking1)
	route, pathParam, rts := domain.Lookup(reqCancelBooking1)
	assert.Equal(t, "cancel_booking", route.Name)
	assert.Equal(t, "12345", pathParam.Get("id"))
	assert.False(t, rts)

	// possible redirect trailing slash
	reqCancelBooking2 := createHTTPRequest("localhost:8000", "/hotels/12345/cancel/")
	reqCancelBooking2.Method = ahttp.MethodPost
	domain = router.Domain(reqCancelBooking2)
	_, _, rts = domain.Lookup(reqCancelBooking2)
	assert.True(t, rts)

	// Lookup by name
	cancelBooking := domain.LookupByName("cancel_booking")
	assert.Equal(t, "hotels_group", cancelBooking.ParentName)
	assert.Equal(t, "cancel_booking", cancelBooking.Name)
	assert.Equal(t, "Hotels", cancelBooking.Controller)
	assert.Equal(t, "POST", cancelBooking.Method)
}

func TestRouterStaticLoadConfiguration(t *testing.T) {
	router := createRouter("routes.conf")

	err := router.Load()
	assert.FailNowOnError(t, err, "")

	// After loading just couple assertion for static

	// /favicon.ico
	req1 := createHTTPRequest("localhost:8000", "/favicon.ico")
	req1.Method = ahttp.MethodGet
	domain := router.Domain(req1)
	route, pathParam, rts := domain.Lookup(req1)
	assert.NotNil(t, pathParam)
	assert.False(t, rts)
	assert.True(t, route.IsStatic)
	assert.Equal(t, "/public/img/favicon.png", route.File)
	assert.Equal(t, "", route.Dir)

	// /static/img/aahframework.png
	req2 := createHTTPRequest("localhost:8000", "/static/img/aahframework.png")
	req2.Method = ahttp.MethodGet
	domain = router.Domain(req2)
	route, pathParam, rts = domain.Lookup(req2)
	assert.NotNil(t, pathParam)
	assert.False(t, rts)
	assert.True(t, route.IsStatic)
	assert.Equal(t, "/public", route.Dir)
	assert.Equal(t, "/img/aahframework.png", pathParam.Get("filepath"))
	assert.Equal(t, "", route.File)
}

func TestRouterErrorLoadConfiguration(t *testing.T) {
	router := createRouter("routes-error.conf")

	err := router.Load()
	assert.NotNilf(t, err, "expected error loading '%v'", "routes-error.conf")
}

func TestRouterErrorHostLoadConfiguration(t *testing.T) {
	router := createRouter("routes-no-hostname.conf")

	err := router.Load()
	assert.NotNilf(t, err, "expected error loading '%v'", "routes-no-hostname.conf")
}

func TestRouterErrorPathLoadConfiguration(t *testing.T) {
	router := createRouter("routes-path-error.conf")

	err := router.Load()
	assert.NotNilf(t, err, "expected error loading '%v'", "routes-path-error.conf")
}

func TestRouterErrorControllerLoadConfiguration(t *testing.T) {
	router := createRouter("routes-controller-error.conf")

	err := router.Load()
	assert.NotNilf(t, err, "expected error loading '%v'", "routes-controller-error.conf")
}

func TestRouterErrorStaticPathLoadConfiguration(t *testing.T) {
	router := createRouter("routes-static-path-error.conf")

	err := router.Load()
	assert.NotNilf(t, err, "expected error loading '%v'", "routes-static-path-error.conf")
}

func TestRouterErrorStaticPathPatternLoadConfiguration(t *testing.T) {
	router := createRouter("routes-static-path-pattern-error.conf")

	err := router.Load()
	assert.NotNilf(t, err, "expected error loading '%v'", "routes-static-path-pattern-error.conf")
}

func TestRouterErrorStaticDirFileLoadConfiguration(t *testing.T) {
	router := createRouter("routes-static-dir-file-error.conf")

	err := router.Load()
	assert.NotNilf(t, err, "expected error loading '%v'", "routes-static-dir-file-error.conf")
}

func TestRouterErrorStaticNoDirFileLoadConfiguration(t *testing.T) {
	router := createRouter("routes-static-no-dir-file-error.conf")

	err := router.Load()
	assert.NotNilf(t, err, "expected error loading '%v'", "routes-static-no-dir-file-error.conf")
}

func TestRouterNoDomainRoutesFound(t *testing.T) {
	router := createRouter("routes-no-domains.conf")

	err := router.Load()
	assert.Equal(t, ErrNoRoutesConfigFound, err)
}

func TestRouterReloadConfiguration(t *testing.T) {
	router := createRouter("routes.conf")

	err := router.Reload()
	assert.FailNowOnError(t, err, "")

	// TODO validate routes after reload
}

func TestRouterDomainConfig(t *testing.T) {
	router := createRouter("routes.conf")
	err := router.Load()
	assert.FailNowOnError(t, err, "")

	domain := router.Domain(createHTTPRequest("localhost:8000", ""))
	assert.NotNil(t, domain)

	domain = router.Domain(createHTTPRequest("www.aahframework.org", ""))
	assert.Nil(t, domain)
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Test Domain methods
//___________________________________

func TestDomainAllowed(t *testing.T) {
	router := createRouter("routes.conf")
	err := router.Load()
	assert.FailNowOnError(t, err, "")

	req := createHTTPRequest("localhost:8000", "/")
	domain := router.Domain(req)
	allow := domain.Allowed(ahttp.MethodGet, "/")
	assert.NotNil(t, allow)
	assert.True(t, strings.Contains(allow, ahttp.MethodOptions))

	domain = router.Domain(req)
	allow = domain.Allowed(ahttp.MethodPost, "*")
	assert.NotNil(t, allow)
	assert.True(t, strings.Contains(allow, ahttp.MethodOptions))

	// domain not exists
	reqNotExists := createHTTPRequest("notexists:8000", "/")
	domain = router.Domain(reqNotExists)
	assert.Nil(t, domain)

	// TODO do more
}

func TestDomainReverseURL(t *testing.T) {
	router := createRouter("routes.conf")
	err := router.Load()
	assert.FailNowOnError(t, err, "")

	req := createHTTPRequest("localhost:8000", "/")
	domain := router.Domain(req)

	// route name not exists
	emptyURL := domain.Reverse("not_exists_routename")
	assert.Equal(t, "", emptyURL)

	// non key-value pair value leads to error
	emptyURL = domain.Reverse("book_hotels", 12345678)
	assert.Equal(t, "", emptyURL)

	// not enough arguments
	emptyURL = domain.Reverse("book_hotels")
	assert.Equal(t, "", emptyURL)

	// incorrect key name scenario
	emptyURL = domain.Reverse("book_hotels", map[string]string{
		"idvalue": "12345678",
	})
	assert.Equal(t, "", emptyURL)

	// static URL
	loginURL := domain.Reverse("login")
	assert.Equal(t, "/login", loginURL)

	// success scenario
	bookingURL := domain.Reverse("book_hotels", map[string]string{
		"id": "12345678",
	})
	assert.Equal(t, "/hotels/12345678/booking", bookingURL)

	bookingURL = domain.Reverse("book_hotels", map[string]string{
		"id":     "12345678",
		"param1": "param1value",
		"param2": "param2value",
	})
	assert.Equal(t, "/hotels/12345678/booking?param1=param1value&param2=param2value", bookingURL)
}

func TestDomainAddRoute(t *testing.T) {
	domain := &Domain{
		Host: "aahframework.org",
	}

	route1 := &Route{
		Name:       "route1",
		Path:       "/info/:user/project/:project",
		Method:     "GET",
		Controller: "Info",
		Action:     "ShowProject",
	}
	err := domain.addRoute(route1)
	assert.FailNowOnError(t, err, "unexpected error")

	route2 := &Route{
		Name:       "index",
		Path:       "/",
		Method:     "GET",
		Controller: "App",
		Action:     "Index",
	}
	err = domain.addRoute(route2)
	assert.FailNowOnError(t, err, "unexpected error")

	routeError := &Route{
		Name:       "route_error",
		Path:       "/",
		Method:     "GET",
		Controller: "App",
		Action:     "Index",
	}
	err = domain.addRoute(routeError)
	assert.True(t, strings.Contains(err.Error(), "value is already registered"))
}

func createRouter(filename string) *Router {
	wd, _ := os.Getwd()
	return New(filepath.Join(wd, "testdata", filename))
}

func createHTTPRequest(host, path string) *ahttp.Request {
	req := &ahttp.Request{
		Request: &http.Request{Host: host},
	}

	if !ess.IsStrEmpty(path) {
		req.URL = &url.URL{Path: path}
	}

	return req
}
