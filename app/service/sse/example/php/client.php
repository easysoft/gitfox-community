<?php

// SSE 服务地址
$url = 'http://localhost:8080/sse?space_id=2024';

echo "正在连接 SSE 服务: $url ...\n";

$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, $url);
// 禁用缓存，请求 SSE 类型
curl_setopt($ch, CURLOPT_HTTPHEADER, array('Cache-Control: no-cache', 'Accept: text/event-stream'));
// 关闭 SSL 验证（如果本地是 https）
curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
// 设置连接超时（0 表示无限）
curl_setopt($ch, CURLOPT_CONNECTTIMEOUT, 10);
// 设置读取超时（0 表示无限，保持长连接）
curl_setopt($ch, CURLOPT_TIMEOUT, 0);

// 关键：设置写回调函数，实现流式处理
// 每当有数据到达时，这个函数就会被调用
curl_setopt($ch, CURLOPT_WRITEFUNCTION, function($ch, $data) {
    // data 可能包含多个事件，或者一个事件的多行
    // 简单的按行分割处理
    $lines = explode("\n", $data);
    foreach ($lines as $line) {
        $line = trim($line);
        if (empty($line)) continue;
        
        if (strpos($line, 'event:') === 0) {
            echo "\n[收到事件] 类型: " . substr($line, 7) . "\n";
        } elseif (strpos($line, 'data:') === 0) {
            echo "[收到数据] 内容: " . substr($line, 6) . "\n";
        } elseif (strpos($line, ':') === 0) {
            // 这是 SSE 的注释行（心跳），通常忽略
            // echo "[心跳]\n";
        }
    }
    // 必须返回处理的字节数，否则 curl 会中断
    return strlen($data);
});

$result = curl_exec($ch);

if (curl_errno($ch)) {
    echo 'Curl error: ' . curl_error($ch) . "\n";
}

curl_close($ch);