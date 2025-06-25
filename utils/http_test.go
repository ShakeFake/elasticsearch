package utils

import (
	"fmt"
	"testing"
)

func TestRequest_SetHeader(t *testing.T) {
	url := "https://d17dp3wyck5yxn.cloudfront.net/ts/3CB5559B28436AFA/55524C20455854200000018ABAEC3611/240516-08-44-45-272/1/play.m3u8"
	cookie := "CloudFront-Key-Pair-Id=K3SM2GVV9JFF5Y;CloudFront-Policy=eyJTdGF0ZW1lbnQiOlt7IlJlc291cmNlIjoiaHR0cHM6Ly9kMTdkcDN3eWNrNXl4bi5jbG91ZGZyb250Lm5ldC90cy8zQ0I1NTU5QjI4NDM2QUZBLzU1NTI0QzIwNDU1ODU0MjAwMDAwMDE4QUJBRUMzNjExLzI0MDUxNi0wOC00NC00NS0yNzIvMS8qIiwiQ29uZGl0aW9uIjp7IkRhdGVHcmVhdGVyVGhhbiI6eyJBV1M6RXBvY2hUaW1lIjoxNzE3NDA3MjI3fSwiRGF0ZUxlc3NUaGFuIjp7IkFXUzpFcG9jaFRpbWUiOjE3MTc0MTA4Mjd9fX1dfQ__;CloudFront-Signature=LCH8K1PnPLchM8X5ZJkLUiRO3Wc1VZRaMhxhTotTwFDyJ9XeZI08on8C41hXKc92ChnMkvw4ZrHVZlNGWu1XTH0sBR-cst6S~JQ~cPCM85R1kPtfgfznX0~XH30F-jcV8i-RE~np-TjuNV5hdVJHMv9Vxxd5y8z4LUS6TsOToXxnoQ9tLKuJxuaBeEAW0QD7sF4OO6V3ZAPh8oJadVQh4SL9YUu0zaz5IQfUVfF6mJE0jZ-IaUZeImHgYqmMlw02B-~E2XAi7yDQ82Zp2bJPGHGnyeFThyGl6jyIv5ZPYWybUQ8kxzE3iXR0QTuRtSHuZx2TR52yK~XVbid-uTuhZg__"

	headers := make(map[string][]string)
	headers["cookie"] = []string{cookie}

	r := GetRequest(url, "GET", headers, nil)
	r.Do()

	fmt.Println(r.Err, r.Message)

}
