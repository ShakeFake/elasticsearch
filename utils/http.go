package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type UserInfo struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Request struct {
	// 请求参数
	Url    string
	Method string
	Header map[string][]string
	Body   []byte

	// 返回结果
	Err     error
	Code    int
	Message string

	UserInfo *UserInfo
}

// GetRequest 获得一个request
func GetRequest(url string, method string, header map[string][]string, body []byte) *Request {
	if header == nil {
		return &Request{Url: url, Method: method, Body: body}
	}
	return &Request{Url: url, Method: method, Header: header, Body: body}
}

func (r *Request) SetUrl(url string) *Request {
	r.Url = url
	return r
}

func (r *Request) SetMethod(method string) *Request {
	r.Method = method
	return r
}

func (r *Request) SetHeader(header map[string][]string) *Request {
	r.Header = header
	return r
}

// SetUserInfo 用来做简单的用户身份验证
func (r *Request) SetUserInfo(name string, password string) *Request {
	r.UserInfo = &UserInfo{Name: name, Password: password}
	return r
}

// SetUserInfoS 接受一个userInfo结构，设置用户信息
func (r *Request) SetUserInfoS(info *UserInfo) *Request {
	r.UserInfo = info
	return r
}

// Do 根据现有条件发起请求
func (r *Request) Do() *Request {
	request, err := http.NewRequest(r.Method, r.Url, bytes.NewReader(r.Body))
	if err != nil {
		r.Err = err
		return r
	}

	// 设置头
	for k, vs := range r.Header {
		for _, v := range vs {
			request.Header.Set(k, v)
		}
	}

	// 设置基础的用户名和密码
	if r.UserInfo != nil {
		request.SetBasicAuth(r.UserInfo.Name, r.UserInfo.Password)
	}

	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		r.Err = err
		return r
	}

	r.Code = resp.StatusCode
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		r.Err = err
		return r
	}
	r.Message = string(body)
	return r
}
