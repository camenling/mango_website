export const apiPath = '/api/sign/blacklist-lookup'

export const payloadSample = `{
  "udid": "<设备 UDID>"
}`

export const noneSample = `{
  "status": "none"
}`

export const channelSample = `{
  "status": "channel",
  "channel_count": 3
}`

export const globalSample = `{
  "status": "global"
}`

export const errorSample = `{
  "error": "<错误码>"
}`

export const pythonSample = `#!/usr/bin/env python3
import json, urllib.request

URL = "https://<本站域名>/api/sign/blacklist-lookup"

def lookup(udid):
    # 步骤 1：准备请求数据
    body = json.dumps({"udid": udid}).encode()
    # 步骤 2：POST JSON，不要再包一层
    req = urllib.request.Request(
        URL,
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req) as resp:
        # 步骤 3：解析返回
        return resp.status, json.loads(resp.read())

status, data = lookup("00008020-001234567890ABCD")
print(status, data)
`

export const phpSample = `<?php
const URL = 'https://<本站域名>/api/sign/blacklist-lookup';

function lookup($udid)
{
    // 步骤 1：准备请求数据
    $body = json_encode(['udid' => $udid], JSON_UNESCAPED_SLASHES);
    // 步骤 2：POST JSON，不要再包一层
    $ch = curl_init(URL);
    curl_setopt_array($ch, [
        CURLOPT_POST => true,
        CURLOPT_HTTPHEADER => ['Content-Type: application/json'],
        CURLOPT_POSTFIELDS => $body,
        CURLOPT_RETURNTRANSFER => true,
    ]);
    $raw = curl_exec($ch);
    $status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    // 步骤 3：解析返回
    return [$status, json_decode($raw, true)];
}

list($status, $data) = lookup('00008020-001234567890ABCD');
echo $status, PHP_EOL, json_encode($data), PHP_EOL;
`

export const goSample = `package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const apiURL = "https://<本站域名>/api/sign/blacklist-lookup"

func lookup(udid string) (int, map[string]any, error) {
	// 步骤 1：准备请求数据
	body, err := json.Marshal(map[string]string{"udid": udid})
	if err != nil {
		return 0, nil, err
	}
	// 步骤 2：POST JSON，不要再包一层
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	// 步骤 3：解析返回
	var data map[string]any
	if err = json.Unmarshal(raw, &data); err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, data, nil
}

func main() {
	status, data, err := lookup("00008020-001234567890ABCD")
	if err != nil {
		panic(err)
	}
	fmt.Println(status, data)
}
`

export const javaSample = `import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

class LookupBlacklist {
    static final String URL = "https://<本站域名>/api/sign/blacklist-lookup";

    static HttpResponse<String> lookup(String udid) throws Exception {
        // 步骤 1：准备请求数据
        String body = "{\\"udid\\":\\"" + udid + "\\"}";
        // 步骤 2：POST JSON，不要再包一层
        HttpRequest req = HttpRequest.newBuilder(URI.create(URL))
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build();
        // 步骤 3：解析返回
        return HttpClient.newHttpClient().send(req, HttpResponse.BodyHandlers.ofString());
    }

    public static void main(String[] args) throws Exception {
        HttpResponse<String> resp = lookup("00008020-001234567890ABCD");
        System.out.println(resp.statusCode());
        System.out.println(resp.body());
    }
}
`
